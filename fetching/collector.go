package fetching

import (
	"fmt"
	"github.com/goinggo/work"
	"sync"
	"sync/atomic"
)

type orderCollector struct {
	client     OrderFetcher
	pool       *work.Pool
	maxWorkers int
	regionId   int32
	orderChan  chan OrderPayload
	done       chan int32

	workerCount workerCounter
}

func (c *orderCollector) Work(idx int) {
	wg := sync.WaitGroup{}
	endReached := make(chan bool)
	stopSpawning := make(chan bool)
	workerDone := make(chan int32)
	spawnAnother := make(chan bool)

	//We don't want the monitor to start until we've kicked off the first batch.
	monitorStart := sync.WaitGroup{}
	monitorStart.Add(1)

	//Spawner routine, incrementing the counter here because the routine will probably not add to it before we ask it
	//to wait
	wg.Add(1)
	go c.spawner(&wg, endReached, workerDone, spawnAnother, stopSpawning)

	//Monitor routine to monitor for workers shutting down or the end being reached
	wg.Add(1)
	go c.monitor(&wg, &monitorStart, endReached, workerDone, spawnAnother, stopSpawning)

	//kick it off with the initial first batch
	for idx := 0; idx < c.maxWorkers; idx++ {
		fmt.Println("Main: Signaling to spawn")
		spawnAnother <- true
		fmt.Println("Main: Done signaling to spawn")
	}

	fmt.Println("Main: Signaling Monitor")
	monitorStart.Done()
	fmt.Println("Main: Done signaling monitor")

	fmt.Println("Main: Waiting")
	wg.Wait()
	fmt.Println("Main: Done waiting")
	fmt.Printf("Main: Workers running %d\n", c.workerCount.Current())

	for c.workerCount.Current() > 0 {
		fmt.Printf("Main: %d Workers still running, allowing them to exit\n", c.workerCount.Current())
		select {
		case <-workerDone:
			fmt.Println("Main: Worker said they're done")
			c.workerCount.Remove()
		case <-endReached:
			fmt.Println("Main: Worker said end reached")
			c.workerCount.Remove()
		}
	}

	fmt.Println("Main: exiting")
	defer fmt.Println("Main: returning")
	c.done <- c.regionId
}

func (c *orderCollector) spawner(wg *sync.WaitGroup, endReached chan bool, workerDone chan int32, spawnAnother, stopSpawning <-chan bool) {
	page := int32(1)
	exit := false

	for {
		select {
		case <-spawnAnother:
			c.workerCount.Add()
			fmt.Println("Spawner: Spawning")
			c.pool.Run(NewWorker(c.client, "all", page, c.regionId, c.orderChan, endReached, workerDone))
			page++
			fmt.Println("Spawner: Done spawning")
		case <-stopSpawning:
			fmt.Println("Spawner: No more workers")
			exit = true
		}

		if exit {
			break
		}
	}

	fmt.Println("Spawner: exiting")
	wg.Done()
}

func (c *orderCollector) monitor(wg *sync.WaitGroup, monitorStart *sync.WaitGroup, endReached <-chan bool, workerDone <-chan int32, spawnAnother, stopSpawning chan<- bool) {
	monitorStart.Wait()
	fmt.Println("Monitor: Starting")
	exit := false

	for {
		select {
		case page := <-workerDone:
			c.workerCount.Remove()
			spawnAnother <- true
			fmt.Printf("Monitor: Worker finished (page: %d)\n", page)
		case <-endReached:
			stopSpawning <- true
			c.workerCount.Remove()
			fmt.Println("Monitor: End Reached")
			exit = true
		}

		if exit {
			break
		}
	}

	fmt.Println("Monitor: exiting")
	wg.Done()
}

func NewCollector(client OrderFetcher, pool *work.Pool, maxWorkers int, done chan int32, regionId int32, all chan OrderPayload) work.Worker {
	return &orderCollector{
		client:     client,
		pool:       pool,
		maxWorkers: maxWorkers,
		regionId:   regionId,
		orderChan:  all,
		done:       done,
	}
}

type workerCounter struct {
	count int64
	lock  sync.Mutex
}

func (wc *workerCounter) Add() {
	wc.lock.Lock()
	atomic.AddInt64(&wc.count, 1)
	wc.lock.Unlock()
}

func (wc *workerCounter) Remove() {
	wc.lock.Lock()
	atomic.AddInt64(&wc.count, -1)
	wc.lock.Unlock()
}

func (wc *workerCounter) Current() int64 {
	wc.lock.Lock()
	defer wc.lock.Unlock()

	return atomic.LoadInt64(&wc.count)
}

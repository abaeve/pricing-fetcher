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

	workerCount int64
}

func (c *orderCollector) Work(idx int) {
	wg := sync.WaitGroup{}
	endReached := make(chan bool)
	stopSpawning := make(chan bool)
	workerDone := make(chan int)
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
	fmt.Printf("Main: Workers running %d\n", atomic.LoadInt64(&c.workerCount))

	for atomic.LoadInt64(&c.workerCount) > 0 {
		fmt.Println("Main: Workers still running, allowing them to exit")
		select {
		case <-workerDone:
			fmt.Println("Main: Worker said they're done")
			atomic.AddInt64(&c.workerCount, -1)
		case <-endReached:
			fmt.Println("Main: Worker said end reached")
			atomic.AddInt64(&c.workerCount, -1)
		}
	}

	c.done <- c.regionId
}

func (c *orderCollector) spawner(wg *sync.WaitGroup, endReached chan bool, workerDone chan int, spawnAnother, stopSpawning <-chan bool) {
	page := 1

NoMore:
	for {
		select {
		case <-spawnAnother:
			atomic.AddInt64(&c.workerCount, 1)
			fmt.Println("Spawner: Spawning")
			c.pool.Run(NewWorker(c.client, "all", page, c.regionId, c.orderChan, endReached, workerDone))
			page++
			fmt.Println("Spawner: Done spawning")
		case <-stopSpawning:
			fmt.Println("Spawner: No more workers")
			break NoMore
		}
	}

	fmt.Println("Spawner: exiting")
	wg.Done()
}

func (c *orderCollector) monitor(wg *sync.WaitGroup, monitorStart *sync.WaitGroup, endReached <-chan bool, workerDone <-chan int, spawnAnother, stopSpawning chan<- bool) {
	monitorStart.Wait()
	fmt.Println("Monitor: Starting")

EndReached:
	for {
		select {
		case page := <-workerDone:
			atomic.AddInt64(&c.workerCount, -1)
			spawnAnother <- true
			fmt.Printf("Monitor: Worker finished (page: %d)\n", page)
		case <-endReached:
			stopSpawning <- true
			atomic.AddInt64(&c.workerCount, -1)
			fmt.Println("Monitor: End Reached")
			break EndReached
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

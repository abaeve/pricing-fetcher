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

	endReached   chan bool
	stopSpawning chan bool
	workerDone   chan int32
	spawnAnother chan bool

	workerCount int64
}

func (c *orderCollector) Work(idx int) {
	wg := sync.WaitGroup{}

	//We don't want the monitor to start until we've kicked off the first batch.
	monitorStart := sync.WaitGroup{}
	monitorStart.Add(1)

	//Spawner routine, incrementing the counter here because the routine will probably not add to it before we ask it
	//to wait
	wg.Add(1)
	go c.spawner(&wg)

	//Monitor routine to monitor for workers shutting down or the end being reached
	wg.Add(1)
	go c.monitor(&wg, &monitorStart)

	//kick it off with the initial first batch
	for idx := 0; idx < c.maxWorkers; idx++ {
		fmt.Println("Main: Signaling to spawn")
		c.spawnAnother <- true
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
		fmt.Printf("Main: %d Workers still running, allowing them to exit\n", atomic.LoadInt64(&c.workerCount))
		select {
		case <-c.workerDone:
			fmt.Println("Main: Worker said they're done")
			atomic.AddInt64(&c.workerCount, -1)
		case <-c.endReached:
			fmt.Println("Main: Worker said end reached")
			atomic.AddInt64(&c.workerCount, -1)
		}
	}

	fmt.Println("Main: exiting")
	defer fmt.Println("Main: returning")
	c.done <- c.regionId
}

func (c *orderCollector) spawner(wg *sync.WaitGroup) {
	page := int32(1)
	exit := false

	for {
		select {
		case <-c.spawnAnother:
			atomic.AddInt64(&c.workerCount, 1)
			fmt.Println("Spawner: Spawning")
			c.pool.Run(NewWorker(c.client, "all", page, c.regionId, c.orderChan, c.endReached, c.workerDone))
			page++
			fmt.Println("Spawner: Done spawning")
		case <-c.stopSpawning:
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

func (c *orderCollector) monitor(wg *sync.WaitGroup, monitorStart *sync.WaitGroup) {
	monitorStart.Wait()
	fmt.Println("Monitor: Starting")
	exit := false

	for {
		select {
		case page := <-c.workerDone:
			atomic.AddInt64(&c.workerCount, -1)
			c.spawnAnother <- true
			fmt.Printf("Monitor: Worker finished (page: %d)\n", page)
		case <-c.endReached:
			c.stopSpawning <- true
			atomic.AddInt64(&c.workerCount, -1)
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
		client:      client,
		pool:        pool,
		maxWorkers:  maxWorkers,
		regionId:    regionId,
		orderChan:   all,
		done:        done,
		workerCount: 0,

		endReached:   make(chan bool),
		workerDone:   make(chan int32),
		stopSpawning: make(chan bool),
		spawnAnother: make(chan bool),
	}
}

package fetching

import (
	"errors"
	"fmt"
	"github.com/antihax/goesi/v1"
	"github.com/goinggo/work"
	"time"
)

//Save this for later... some day I'll either decide to parse the error out of the http.Response or the go swagger generator will not produce garbage code
var RegionNotFoundError error = errors.New("Did not find the specified region")

type OrderController interface {
	Fetch(regionId int32) error
	GetDoneChannel() chan int32
	Stop()
}

type orderController struct {
	regionsFetcher RegionsFetcher
	orderClient    OrderFetcher
	pool           *work.Pool
	publisher      OrderPublisher
	collectors     map[int32]work.Worker

	workerDone chan int32
	orders     chan OrderPayload

	fetchedRegions map[int32]*goesiv1.GetUniverseRegionsRegionIdOk
	maxWorkers     int
	maxRegions     int
	start          chan int32
	stop           chan bool

	clientDone chan int32

	publishingBinder publisherBinding
	doPublish        chan bool
}

func (o *orderController) Fetch(regionId int32) error {
	if o.fetchedRegions[regionId] == nil {
		region, _, err := o.regionsFetcher.GetUniverseRegionsRegionId(regionId, nil)

		if err != nil {
			return err
		}

		o.fetchedRegions[regionId] = &region
	}

	if o.collectors[regionId] == nil {
		o.collectors[regionId] = NewCollector(o.orderClient, o.pool, o.maxWorkers/o.maxRegions, o.workerDone, regionId, o.orders)
	}

	//Not in the mood to wait for it to queue it... just do it
	o.start <- regionId
	go func() {
		o.pool.Run(o.collectors[regionId])
	}()

	return nil
}

func (o *orderController) GetDoneChannel() chan int32 {
	o.doPublish <- true
	return o.clientDone
}

func (o *orderController) Stop() {
	o.stop <- true
	o.pool.Shutdown()
}

func NewController(regionFetcher RegionsFetcher, orderFetcher OrderFetcher, orderPublisher OrderPublisher, maxRegions int, maxDownloaders int, logFunc func(string)) (OrderController, error) {
	if logFunc == nil {
		logFunc = defaultLogFunc
	}

	//Our pool will contain one routine per download, one per region collector and one for the publisher (or rather the routine that calls the publishers method
	pool, _ := work.New(maxDownloaders+maxRegions+1, time.Second, logFunc)

	if maxDownloaders == 0 || maxRegions == 0 {
		return nil, errors.New("Could not create a pool with only 1 worker")
	}

	collectors := make(map[int32]work.Worker)
	doneChan := make(chan int32)
	ordersChan := make(chan OrderPayload)

	regionsCache := make(map[int32]*goesiv1.GetUniverseRegionsRegionIdOk)
	startChan := make(chan int32)
	stopChan := make(chan bool)
	clientDoneChan := make(chan int32)
	doPublishToClient := make(chan bool)

	binder := publisherBinding{
		publisher:  orderPublisher,
		start:      startChan,
		done:       doneChan,
		orders:     ordersChan,
		stop:       stopChan,
		clientDone: clientDoneChan,

		publishDoneToClient: false,
		doPublish:           doPublishToClient,
	}

	//Now let the publishing binder do its magic
	pool.Run(&binder)

	return &orderController{
		regionsFetcher: regionFetcher,
		orderClient:    orderFetcher,
		pool:           pool,
		publisher:      orderPublisher,
		collectors:     collectors,
		workerDone:     doneChan,
		orders:         ordersChan,

		fetchedRegions: regionsCache,
		maxWorkers:     maxDownloaders,
		maxRegions:     maxRegions,
		start:          startChan,
		stop:           stopChan,
		clientDone:     clientDoneChan,

		publishingBinder: binder,
		doPublish:        doPublishToClient,
	}, nil
}

type publisherBinding struct {
	publisher OrderPublisher
	start     chan int32
	done      chan int32
	orders    chan OrderPayload

	stop       chan bool
	clientDone chan int32

	publishDoneToClient bool
	doPublish           chan bool
}

func (pb *publisherBinding) Work(id int) {
Exiting:
	for {
		select {
		case order := <-pb.orders:
			//fmt.Println("Publisher: Publishing an order")
			pb.publisher.PublishOrder(&order)
		case finishedRegion := <-pb.done:
			fmt.Println("Publisher: Publishing State End")
			pb.publisher.PublishStateEnd(finishedRegion)
			if pb.publishDoneToClient {
				fmt.Println("Trying to publish to client")
				pb.clientDone <- finishedRegion
			}
		case startedRegion := <-pb.start:
			fmt.Println("Publisher: Publishing State Begin")
			pb.publisher.PublishStateBegin(startedRegion)
		case <-pb.stop:
			fmt.Println("Publisher: Exiting")
			break Exiting
		case pb.publishDoneToClient = <-pb.doPublish:
		}
	}
}

func defaultLogFunc(message string) {
	//I do nothing on purpose
}

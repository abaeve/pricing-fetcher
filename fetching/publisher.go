package fetching

import (
	"context"
	"encoding/json"
	"github.com/micro/go-micro/broker"
	"strconv"
	"strings"
	"sync"
)

var defaultRegion string = "unknown"

type OrderPublisher interface {
	PublishOrder(order *OrderPayload)
	PublishStateBegin(regionInfo RegionInfo)
	PublishStateEnd(regionInfo RegionInfo)
}

type orderPublisher struct {
	broker        broker.Broker
	regionFetcher RegionsFetcher

	regionLock  map[int32]*sync.WaitGroup
	regionCache map[int32]string
}

func (op *orderPublisher) PublishOrder(order *OrderPayload) {
	payload, err := json.Marshal(order)

	if err != nil {
		//Damn I really need a logging framework...
		//TODO: FIND A DAMN LOGGING FRAMEWORK!
		return
	}

	orderType := "sell"
	if order.IsBuyOrder {
		orderType = "buy"
	}

	//Only wait to use the broker because it's the only thing that needs to region's name
	op.regionLock[order.RegionId].Wait()
	op.broker.Publish(orderType+"."+strconv.Itoa(int(order.RegionId)), &broker.Message{
		Body: []byte(payload),
	})
}

func (op *orderPublisher) PublishStateBegin(regionInfo RegionInfo) {
	op.regionLock[regionInfo.regionId] = &sync.WaitGroup{}
	op.regionLock[regionInfo.regionId].Add(1)
	if len(op.regionCache[regionInfo.regionId]) == 0 || op.regionCache[regionInfo.regionId] == defaultRegion {
		region, _, err := op.regionFetcher.GetUniverseRegionsRegionId(context.Background(), regionInfo.regionId, nil)

		if err != nil {
			//I really need to find a logging framework... not much I can do here besides pick a default region?
			op.regionCache[region.RegionId] = defaultRegion
		}

		regionName := strings.ToLower(region.Name)
		regionName = strings.Replace(regionName, " ", "-", -1)
		op.regionCache[region.RegionId] = regionName
	}
	op.regionLock[regionInfo.regionId].Done()

	op.broker.Publish(strconv.Itoa(int(regionInfo.regionId))+".state.begin", &broker.Message{
		Body: []byte("Starting"),
	})
}

func (op *orderPublisher) PublishStateEnd(regionInfo RegionInfo) {
	op.broker.Publish(strconv.Itoa(int(regionInfo.regionId))+".state.end", &broker.Message{
		Body: []byte(regionInfo.fetchRequestId),
	})
}

func NewPublisher(regionFetcher RegionsFetcher, brkr broker.Broker) OrderPublisher {
	//We absolutely have to publish the state begin before any orders fly through the wire
	//I debated on a Mutex or WaitGroup here, the reasoning behind the wait group is that I'm only waiting on the
	//region cache to populate per region and don't want concurrent downloader routines to block each other.  I ONLY
	//want to wait for the cache (which is the only thing that does Add() and Done())
	regionLock := make(map[int32]*sync.WaitGroup)
	regionCache := make(map[int32]string)

	return &orderPublisher{
		broker:        brkr,
		regionFetcher: regionFetcher,
		regionLock:    regionLock,
		regionCache:   regionCache,
	}
}

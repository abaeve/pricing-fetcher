package fetching

import (
	"fmt"
	"github.com/antihax/goesi/v1"
	"github.com/goinggo/work"
	"github.com/golang/mock/gomock"
	"testing"
	"time"
)

func TestOrderCollector_Fetch_2PagesAnd2Workers(t *testing.T) {
	//t.SkipNow()
	mockCtrl := gomock.NewController(t)
	mockOrderFetcher := NewMockOrderFetcher(mockCtrl)
	defer mockCtrl.Finish()

	//BGN Expectations
	pageOne := make(map[string]interface{})
	pageOne["page"] = 1

	mockOrderFetcher.EXPECT().GetMarketsRegionIdOrders("all", int32(12345), pageOne).Return(
		[]goesiv1.GetMarketsRegionIdOrders200Ok{
			{
				VolumeTotal:  20,
				VolumeRemain: 20,
				TypeId:       1,
				Range_:       "region",
				Price:        1.1,
				OrderId:      int64(1),
				MinVolume:    2,
				LocationId:   123456,
				Issued:       time.Now(),
				IsBuyOrder:   false,
				Duration:     40,
			},
			{
				VolumeTotal:  20,
				VolumeRemain: 20,
				TypeId:       1,
				Range_:       "region",
				Price:        1.1,
				OrderId:      int64(2),
				MinVolume:    2,
				LocationId:   1234567,
				Issued:       time.Now(),
				IsBuyOrder:   true,
				Duration:     40,
			},
		}, nil, nil,
	)

	pageTwo := make(map[string]interface{})
	pageTwo["page"] = 2

	mockOrderFetcher.EXPECT().GetMarketsRegionIdOrders("all", int32(12345), pageTwo).Return(
		[]goesiv1.GetMarketsRegionIdOrders200Ok{
			{
				VolumeTotal:  20,
				VolumeRemain: 20,
				TypeId:       1,
				Range_:       "region",
				Price:        1.1,
				OrderId:      int64(3),
				MinVolume:    2,
				LocationId:   1234567,
				Issued:       time.Now(),
				IsBuyOrder:   false,
				Duration:     40,
			},
			{
				VolumeTotal:  20,
				VolumeRemain: 20,
				TypeId:       1,
				Range_:       "region",
				Price:        1.1,
				OrderId:      int64(4),
				MinVolume:    2,
				LocationId:   1234567,
				Issued:       time.Now(),
				IsBuyOrder:   true,
				Duration:     40,
			},
		}, nil, nil,
	)

	//We have to allow both pages 3 and 4 because this interface is stupid and this case is spawning workers at a time
	//I guess CCP don't really want people threading these requests easily?
	pageThree := make(map[string]interface{})
	pageThree["page"] = 3

	mockOrderFetcher.EXPECT().GetMarketsRegionIdOrders("all", int32(12345), pageThree).Return(
		[]goesiv1.GetMarketsRegionIdOrders200Ok{}, nil, nil,
	).MaxTimes(1)

	pageFour := make(map[string]interface{})
	pageFour["page"] = 4

	mockOrderFetcher.EXPECT().GetMarketsRegionIdOrders("all", int32(12345), pageFour).Return(
		[]goesiv1.GetMarketsRegionIdOrders200Ok{}, nil, nil,
	).MaxTimes(1)
	//END Expectations

	//We're reading from a channel, need a way to time the test out so we don't hang something up
	//https://blog.golang.org/go-concurrency-patterns-timing-out-and
	timeout := make(chan bool, 1)
	go func() {
		time.Sleep(5 * time.Second)
		timeout <- true
	}()

	//Make the pool
	pool, err := work.New(2, time.Millisecond*250, func(message string) { fmt.Println(message) })

	if err != nil {
		t.Fatalf("Error instantiating pool: %s", err.Error())
	}

	//Build the collector and everything it needs
	//It's designed to be run as a go routine inside a pool so the variables for it's execution start
	//need to be set inside the struct
	done := make(chan int32)
	all := make(chan OrderPayload)
	collector := NewCollector(mockOrderFetcher, pool, 2, done, 12345, all)

	go collector.Work(1)

	foundBuyOrders := 0
	foundSellOrders := 0

	for {
		select {
		case <-timeout:
			t.Fatal("Test timed out")
		case order := <-all:
			if order.IsBuyOrder && (order.OrderId == int64(2) || order.OrderId == int64(4)) {
				foundBuyOrders++
			} else if !order.IsBuyOrder && (order.OrderId == int64(1) || order.OrderId == int64(3)) {
				foundSellOrders++
			} else {
				t.Errorf("Unmatched order id and type: %d", order.OrderId)
			}
		}

		fmt.Println("Test: Got _something_, checking if its what we want")

		if foundBuyOrders == 2 && foundSellOrders == 2 {
			break
		}
	}

	fmt.Println("Test: Timing out or waiting for done signal")

	select {
	case <-timeout:
		t.Fatal("Test timed out waiting for done signal")
	case regionId := <-done:
		if regionId != int32(12345) {
			t.Errorf("Expected region id: (%d) but received (%d)", 12345, regionId)
		}
		fmt.Println("Test: Done received")
	}

	dontPanic := make(chan bool)

	go func(dontPanic chan bool) {
		time.Sleep(time.Second * 1)

		select {
		case <-dontPanic:
		default:
			panic("Timed out waiting for the pool to shutdown")
		}
	}(dontPanic)

	time.Sleep(time.Millisecond * 500)
	fmt.Println("Test: Shutting down the pool")
	pool.Shutdown()
	close(dontPanic)
}

func TestOrderCollector_Fetch_20PagesAnd10Workers(t *testing.T) {
	//t.SkipNow()
	mockCtrl := gomock.NewController(t)
	mockOrderFetcher := NewMockOrderFetcher(mockCtrl)
	defer mockCtrl.Finish()

	//Expectations
	pageOne := make(map[string]interface{})
	pageOne["page"] = 1

	mockOrderFetcher.EXPECT().GetMarketsRegionIdOrders("all", int32(12345), pageOne).Return(
		[]goesiv1.GetMarketsRegionIdOrders200Ok{
			{
				VolumeTotal:  20,
				VolumeRemain: 20,
				TypeId:       1,
				Range_:       "region",
				Price:        1.1,
				OrderId:      int64(1),
				MinVolume:    2,
				LocationId:   123456,
				Issued:       time.Now(),
				IsBuyOrder:   false,
				Duration:     40,
			},
			{
				VolumeTotal:  20,
				VolumeRemain: 20,
				TypeId:       1,
				Range_:       "region",
				Price:        1.1,
				OrderId:      int64(2),
				MinVolume:    2,
				LocationId:   1234567,
				Issued:       time.Now(),
				IsBuyOrder:   true,
				Duration:     40,
			},
		}, nil, nil,
	)

	pageTwo := make(map[string]interface{})
	pageTwo["page"] = 2

	mockOrderFetcher.EXPECT().GetMarketsRegionIdOrders("all", int32(12345), pageTwo).Return(
		[]goesiv1.GetMarketsRegionIdOrders200Ok{
			{
				VolumeTotal:  20,
				VolumeRemain: 20,
				TypeId:       1,
				Range_:       "region",
				Price:        1.1,
				OrderId:      int64(3),
				MinVolume:    2,
				LocationId:   1234567,
				Issued:       time.Now(),
				IsBuyOrder:   false,
				Duration:     40,
			},
			{
				VolumeTotal:  20,
				VolumeRemain: 20,
				TypeId:       1,
				Range_:       "region",
				Price:        1.1,
				OrderId:      int64(4),
				MinVolume:    2,
				LocationId:   1234567,
				Issued:       time.Now(),
				IsBuyOrder:   true,
				Duration:     40,
			},
		}, nil, nil,
	)

	//We have to allow both pages 3 and 4 because this interface is stupid and this case is spawning workers at a time
	//I guess CCP don't really want people threading these requests easily?
	pageThree := make(map[string]interface{})
	pageThree["page"] = 3

	mockOrderFetcher.EXPECT().GetMarketsRegionIdOrders("all", int32(12345), pageThree).Return(
		[]goesiv1.GetMarketsRegionIdOrders200Ok{}, nil, nil,
	).MaxTimes(1)

	pageFour := make(map[string]interface{})
	pageFour["page"] = 4

	mockOrderFetcher.EXPECT().GetMarketsRegionIdOrders("all", int32(12345), pageFour).Return(
		[]goesiv1.GetMarketsRegionIdOrders200Ok{}, nil, nil,
	).MaxTimes(1)

	//We're reading from a channel, need a way to time the test out so we don't hang something up
	//https://blog.golang.org/go-concurrency-patterns-timing-out-and
	timeout := make(chan bool, 1)
	go func() {
		time.Sleep(5 * time.Second)
		timeout <- true
	}()

	//Make the pool
	pool, err := work.New(2, time.Millisecond*250, func(message string) { fmt.Println(message) })

	if err != nil {
		t.Fatalf("Error instantiating pool: %s", err.Error())
	}

	//Build the collector and everything it needs
	//It's designed to be run as a go routine inside a pool so the variables for it's execution start
	//need to be set inside the struct
	done := make(chan int32)
	all := make(chan OrderPayload)
	collector := NewCollector(mockOrderFetcher, pool, 2, done, 12345, all)

	go collector.Work(1)

	foundBuyOrders := 0
	foundSellOrders := 0

	for {
		select {
		case <-timeout:
			t.Fatal("Test timed out")
		case order := <-all:
			if order.IsBuyOrder && (order.OrderId == int64(2) || order.OrderId == int64(4)) {
				foundBuyOrders++
			} else if !order.IsBuyOrder && (order.OrderId == int64(1) || order.OrderId == int64(3)) {
				foundSellOrders++
			} else {
				t.Errorf("Unmatched order id and type: %d", order.OrderId)
			}
		}
		if foundBuyOrders == 2 && foundSellOrders == 2 {
			break
		}
	}

	select {
	case <-timeout:
		t.Fatal("Test timed out waiting for done signal")
	case regionId := <-done:
		if regionId != int32(12345) {
			t.Errorf("Expected region id: (%d) but received (%d)", 12345, regionId)
		}
		fmt.Println("Test: Done received")
	}

	dontPanic := make(chan bool)

	go func(dontPanic chan bool) {
		time.Sleep(time.Second * 1)

		select {
		case <-dontPanic:
		default:
			panic("Timed out waiting for the pool to shutdown")
		}
	}(dontPanic)

	time.Sleep(time.Millisecond * 500)
	fmt.Println("Test: Shutting down the pool")
	pool.Shutdown()
	close(dontPanic)
}

// A simple helper function to give me a bunch of orders for larger tests.  The numberOfOrders param should be even as this'll
// add half the number of sell orders and half the number of buy orders
func addPageWithExpectations(page int, numberOfOrders, nextOrderId int64, regionId int32, mockOrderFetcher MockOrderFetcher) {
	orders := []goesiv1.GetMarketsRegionIdOrders200Ok{}

	for idx := int64(0); idx < numberOfOrders/2; idx++ {
		orders = append(orders, goesiv1.GetMarketsRegionIdOrders200Ok{
			VolumeTotal:  20,
			VolumeRemain: 20,
			TypeId:       1,
			Range_:       "region",
			Price:        1.1,
			OrderId:      nextOrderId,
			MinVolume:    2,
			LocationId:   1234567,
			Issued:       time.Now(),
			IsBuyOrder:   false,
			Duration:     40,
		})
		nextOrderId++
	}

	for idx := int64(0); idx < numberOfOrders/2; idx++ {
		orders = append(orders, goesiv1.GetMarketsRegionIdOrders200Ok{
			VolumeTotal:  20,
			VolumeRemain: 20,
			TypeId:       1,
			Range_:       "region",
			Price:        1.1,
			OrderId:      nextOrderId,
			MinVolume:    2,
			LocationId:   1234567,
			Issued:       time.Now(),
			IsBuyOrder:   true,
			Duration:     40,
		})
		nextOrderId++
	}

	options := make(map[string]interface{})
	options["page"] = page

	mockOrderFetcher.EXPECT().GetMarketsRegionIdOrders("all", regionId, options).Return(orders, nil, nil).MaxTimes(1)
}

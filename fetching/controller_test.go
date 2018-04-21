package fetching

import (
	"errors"
	"fmt"
	goesiv1 "github.com/antihax/goesi/esi"
	"github.com/antihax/goesi/optional"
	"github.com/golang/mock/gomock"
	"testing"
	"time"
)

func TestNewController(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	mockRegionFetcher := NewMockRegionsFetcher(mockCtrl)
	mockOrderFetcher := NewMockOrderFetcher(mockCtrl)
	mockPublisher := NewMockOrderPublisher(mockCtrl)
	defer mockCtrl.Finish()

	var controller *orderController

	ctrler, err := NewController(mockRegionFetcher, mockOrderFetcher, mockPublisher, 5, 10, nil, time.Millisecond*250)

	if err != nil {
		t.Error("Received an error when none were expected")
	}

	controller = ctrler.(*orderController)

	if controller.publisher != mockPublisher {
		t.Error("NewController didn't properly set internal publisher field")
	}

	if controller.regionsFetcher != mockRegionFetcher {
		t.Error("NewController didn't properly set regionsFetcher field")
	}

	if controller.orderClient != mockOrderFetcher {
		t.Error("NewController didn't properly set orderClient field")
	}
}

func TestOrderController_Fetch(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	mockRegionFetcher := NewMockRegionsFetcher(mockCtrl)
	mockOrderFetcher := NewMockOrderFetcher(mockCtrl)
	mockPublisher := NewMockOrderPublisher(mockCtrl)
	defer mockCtrl.Finish()

	//BGN Expectations
	mockRegionFetcher.EXPECT().GetUniverseRegions(gomock.Any(), gomock.Nil()).Return([]int32{12345}, nil, nil).MaxTimes(1)
	mockRegionFetcher.EXPECT().GetUniverseRegionsRegionId(gomock.Any(), int32(12345), gomock.Nil()).Return(goesiv1.GetUniverseRegionsRegionIdOk{
		Name:           "The Forge",
		Constellations: []int32{},
		Description:    "Herpa derpa blerpa Jita",
		RegionId:       int32(12345),
	}, nil, nil)

	orderOne := goesiv1.GetMarketsRegionIdOrders200Ok{
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
	}
	orderTwo := goesiv1.GetMarketsRegionIdOrders200Ok{
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
	}
	orderThree := goesiv1.GetMarketsRegionIdOrders200Ok{
		VolumeTotal:  20,
		VolumeRemain: 20,
		TypeId:       1,
		Range_:       "region",
		Price:        1.1,
		OrderId:      int64(3),
		MinVolume:    2,
		LocationId:   12345678,
		Issued:       time.Now(),
		IsBuyOrder:   false,
		Duration:     40,
	}
	orderFour := goesiv1.GetMarketsRegionIdOrders200Ok{
		VolumeTotal:  20,
		VolumeRemain: 20,
		TypeId:       1,
		Range_:       "region",
		Price:        1.1,
		OrderId:      int64(4),
		MinVolume:    2,
		LocationId:   123456789,
		Issued:       time.Now(),
		IsBuyOrder:   true,
		Duration:     40,
	}

	pageOne := &goesiv1.GetMarketsRegionIdOrdersOpts{}
	pageOne.Page = optional.NewInt32(int32(1))

	mockOrderFetcher.EXPECT().GetMarketsRegionIdOrders(gomock.Any(), "all", int32(12345), pageOne).Return(
		[]goesiv1.GetMarketsRegionIdOrders200Ok{orderOne, orderTwo}, nil, nil,
	)

	pageTwo := &goesiv1.GetMarketsRegionIdOrdersOpts{}
	pageTwo.Page = optional.NewInt32(int32(2))

	mockOrderFetcher.EXPECT().GetMarketsRegionIdOrders(gomock.Any(), "all", int32(12345), pageTwo).Return(
		[]goesiv1.GetMarketsRegionIdOrders200Ok{orderThree, orderFour}, nil, nil,
	)

	//We have to allow both pages 3 and 4 because this interface is stupid and this case is spawning workers at a time
	//I guess CCP don't really want people threading these requests easily?
	pageThree := &goesiv1.GetMarketsRegionIdOrdersOpts{}
	pageThree.Page = optional.NewInt32(int32(3))

	mockOrderFetcher.EXPECT().GetMarketsRegionIdOrders(gomock.Any(), "all", int32(12345), pageThree).Return(
		[]goesiv1.GetMarketsRegionIdOrders200Ok{}, nil, nil,
	).MaxTimes(1)

	pageFour := &goesiv1.GetMarketsRegionIdOrdersOpts{}
	pageFour.Page = optional.NewInt32(int32(4))

	mockOrderFetcher.EXPECT().GetMarketsRegionIdOrders(gomock.Any(), "all", int32(12345), pageFour).Return(
		[]goesiv1.GetMarketsRegionIdOrders200Ok{}, nil, nil,
	).MaxTimes(1)

	pageFive := &goesiv1.GetMarketsRegionIdOrdersOpts{}
	pageFive.Page = optional.NewInt32(int32(5))

	mockOrderFetcher.EXPECT().GetMarketsRegionIdOrders(gomock.Any(), "all", int32(12345), pageFive).Return(
		[]goesiv1.GetMarketsRegionIdOrders200Ok{}, nil, nil,
	).MaxTimes(1)

	pageSix := &goesiv1.GetMarketsRegionIdOrdersOpts{}
	pageSix.Page = optional.NewInt32(int32(6))

	mockOrderFetcher.EXPECT().GetMarketsRegionIdOrders(gomock.Any(), "all", int32(12345), pageSix).Return(
		[]goesiv1.GetMarketsRegionIdOrders200Ok{}, nil, nil,
	).MaxTimes(1)

	//Replace the newUUID func with something constant
	newUUID = func() (string, error) {
		return "123456", nil
	}

	mockPublisher.EXPECT().PublishStateBegin(int32(12345))
	mockPublisher.EXPECT().PublishOrder(&OrderPayload{
		RegionId:       12345,
		FetchRequestId: "123456",
		OrderId:        orderOne.OrderId,
		TypeId:         orderOne.TypeId,
		LocationId:     orderOne.LocationId,
		SystemId:       orderOne.SystemId,
		VolumeTotal:    orderOne.VolumeTotal,
		VolumeRemain:   orderOne.VolumeRemain,
		MinVolume:      orderOne.MinVolume,
		Price:          orderOne.Price,
		IsBuyOrder:     orderOne.IsBuyOrder,
		Duration:       orderOne.Duration,
		Issued:         orderOne.Issued,
		Range_:         orderOne.Range_,
	})
	mockPublisher.EXPECT().PublishOrder(&OrderPayload{
		RegionId:       12345,
		FetchRequestId: "123456",
		OrderId:        orderTwo.OrderId,
		TypeId:         orderTwo.TypeId,
		LocationId:     orderTwo.LocationId,
		SystemId:       orderTwo.SystemId,
		VolumeTotal:    orderTwo.VolumeTotal,
		VolumeRemain:   orderTwo.VolumeRemain,
		MinVolume:      orderTwo.MinVolume,
		Price:          orderTwo.Price,
		IsBuyOrder:     orderTwo.IsBuyOrder,
		Duration:       orderTwo.Duration,
		Issued:         orderTwo.Issued,
		Range_:         orderTwo.Range_,
	})
	mockPublisher.EXPECT().PublishOrder(&OrderPayload{
		RegionId:       12345,
		FetchRequestId: "123456",
		OrderId:        orderThree.OrderId,
		TypeId:         orderThree.TypeId,
		LocationId:     orderThree.LocationId,
		SystemId:       orderThree.SystemId,
		VolumeTotal:    orderThree.VolumeTotal,
		VolumeRemain:   orderThree.VolumeRemain,
		MinVolume:      orderThree.MinVolume,
		Price:          orderThree.Price,
		IsBuyOrder:     orderThree.IsBuyOrder,
		Duration:       orderThree.Duration,
		Issued:         orderThree.Issued,
		Range_:         orderThree.Range_,
	})
	mockPublisher.EXPECT().PublishOrder(&OrderPayload{
		RegionId:       12345,
		FetchRequestId: "123456",
		OrderId:        orderFour.OrderId,
		TypeId:         orderFour.TypeId,
		LocationId:     orderFour.LocationId,
		SystemId:       orderFour.SystemId,
		VolumeTotal:    orderFour.VolumeTotal,
		VolumeRemain:   orderFour.VolumeRemain,
		MinVolume:      orderFour.MinVolume,
		Price:          orderFour.Price,
		IsBuyOrder:     orderFour.IsBuyOrder,
		Duration:       orderFour.Duration,
		Issued:         orderFour.Issued,
		Range_:         orderFour.Range_,
	})
	mockPublisher.EXPECT().PublishStateEnd(int32(12345))
	//END Expectations

	controller, err := NewController(mockRegionFetcher, mockOrderFetcher, mockPublisher, 1, 4, nil, time.Millisecond*250)

	if err != nil {
		t.Error("Received an error when none were expected")
	}

	//We're reading from a channel, need a way to time the test out so we don't hang something up
	//https://blog.golang.org/go-concurrency-patterns-timing-out-and
	timeout := make(chan bool, 1)
	go func() {
		time.Sleep(5 * time.Second)
		timeout <- true
	}()

	done := controller.GetDoneChannel()

	//This doesn't have to be executed in it's own goroutine but this make it easier in the test
	go controller.Fetch(12345)

	select {
	case <-timeout:
		t.Fatal("Test timed out waiting for done signal")
	case regionId := <-done:
		if regionId != int32(12345) {
			t.Errorf("Expected region id: (%d) but received (%d)", 12345, regionId)
		}
		controller.Stop()
		fmt.Println("Test: Done received")
	}
}

func TestOrderController_Fetch_2Regions(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	mockRegionFetcher := NewMockRegionsFetcher(mockCtrl)
	mockOrderFetcher := NewMockOrderFetcher(mockCtrl)
	mockPublisher := NewMockOrderPublisher(mockCtrl)
	defer mockCtrl.Finish()

	//BGN Expectations
	mockRegionFetcher.EXPECT().GetUniverseRegions(gomock.Any(), gomock.Nil()).Return([]int32{12345}, nil, nil).MaxTimes(1)
	mockRegionFetcher.EXPECT().GetUniverseRegionsRegionId(gomock.Any(), int32(12345), gomock.Nil()).Return(goesiv1.GetUniverseRegionsRegionIdOk{
		Name:           "The Forge",
		Constellations: []int32{},
		Description:    "Herpa derpa blerpa Jita",
		RegionId:       int32(12345),
	}, nil, nil)
	mockRegionFetcher.EXPECT().GetUniverseRegionsRegionId(gomock.Any(), int32(12346), gomock.Nil()).Return(goesiv1.GetUniverseRegionsRegionIdOk{
		Name:           "The Derp",
		Constellations: []int32{},
		Description:    "Herpa derpa blerpa Jita",
		RegionId:       int32(12346),
	}, nil, nil)

	orderOne := goesiv1.GetMarketsRegionIdOrders200Ok{
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
	}
	orderTwo := goesiv1.GetMarketsRegionIdOrders200Ok{
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
	}
	orderThree := goesiv1.GetMarketsRegionIdOrders200Ok{
		VolumeTotal:  20,
		VolumeRemain: 20,
		TypeId:       1,
		Range_:       "region",
		Price:        1.1,
		OrderId:      int64(3),
		MinVolume:    2,
		LocationId:   12345678,
		Issued:       time.Now(),
		IsBuyOrder:   false,
		Duration:     40,
	}
	orderFour := goesiv1.GetMarketsRegionIdOrders200Ok{
		VolumeTotal:  20,
		VolumeRemain: 20,
		TypeId:       1,
		Range_:       "region",
		Price:        1.1,
		OrderId:      int64(4),
		MinVolume:    2,
		LocationId:   123456789,
		Issued:       time.Now(),
		IsBuyOrder:   true,
		Duration:     40,
	}

	pageOne := &goesiv1.GetMarketsRegionIdOrdersOpts{}
	pageOne.Page = optional.NewInt32(int32(1))
	pageTwo := &goesiv1.GetMarketsRegionIdOrdersOpts{}
	pageTwo.Page = optional.NewInt32(int32(2))
	//We have to allow both pages 3 and 4 because this interface is stupid and this case is spawning workers at a time
	//I guess CCP don't really want people threading these requests easily?
	pageThree := &goesiv1.GetMarketsRegionIdOrdersOpts{}
	pageThree.Page = optional.NewInt32(int32(3))
	pageFour := &goesiv1.GetMarketsRegionIdOrdersOpts{}
	pageFour.Page = optional.NewInt32(int32(4))
	pageFive := &goesiv1.GetMarketsRegionIdOrdersOpts{}
	pageFive.Page = optional.NewInt32(int32(5))
	pageSix := &goesiv1.GetMarketsRegionIdOrdersOpts{}
	pageSix.Page = optional.NewInt32(int32(6))

	mockOrderFetcher.EXPECT().GetMarketsRegionIdOrders(gomock.Any(), "all", int32(12345), pageOne).Return(
		[]goesiv1.GetMarketsRegionIdOrders200Ok{orderOne, orderTwo}, nil, nil,
	)
	mockOrderFetcher.EXPECT().GetMarketsRegionIdOrders(gomock.Any(), "all", int32(12345), pageTwo).Return(
		[]goesiv1.GetMarketsRegionIdOrders200Ok{orderThree, orderFour}, nil, nil,
	)
	mockOrderFetcher.EXPECT().GetMarketsRegionIdOrders(gomock.Any(), "all", int32(12345), pageThree).Return(
		[]goesiv1.GetMarketsRegionIdOrders200Ok{}, nil, nil,
	).MaxTimes(1)
	mockOrderFetcher.EXPECT().GetMarketsRegionIdOrders(gomock.Any(), "all", int32(12345), pageFour).Return(
		[]goesiv1.GetMarketsRegionIdOrders200Ok{}, nil, nil,
	).MaxTimes(1)
	mockOrderFetcher.EXPECT().GetMarketsRegionIdOrders(gomock.Any(), "all", int32(12345), pageFive).Return(
		[]goesiv1.GetMarketsRegionIdOrders200Ok{}, nil, nil,
	).MaxTimes(1)
	mockOrderFetcher.EXPECT().GetMarketsRegionIdOrders(gomock.Any(), "all", int32(12345), pageSix).Return(
		[]goesiv1.GetMarketsRegionIdOrders200Ok{}, nil, nil,
	).MaxTimes(1)

	mockOrderFetcher.EXPECT().GetMarketsRegionIdOrders(gomock.Any(), "all", int32(12346), pageOne).Return(
		[]goesiv1.GetMarketsRegionIdOrders200Ok{orderOne, orderTwo}, nil, nil,
	)
	mockOrderFetcher.EXPECT().GetMarketsRegionIdOrders(gomock.Any(), "all", int32(12346), pageTwo).Return(
		[]goesiv1.GetMarketsRegionIdOrders200Ok{orderThree, orderFour}, nil, nil,
	)
	mockOrderFetcher.EXPECT().GetMarketsRegionIdOrders(gomock.Any(), "all", int32(12346), pageThree).Return(
		[]goesiv1.GetMarketsRegionIdOrders200Ok{}, nil, nil,
	).MaxTimes(1)
	mockOrderFetcher.EXPECT().GetMarketsRegionIdOrders(gomock.Any(), "all", int32(12346), pageFour).Return(
		[]goesiv1.GetMarketsRegionIdOrders200Ok{}, nil, nil,
	).MaxTimes(1)
	mockOrderFetcher.EXPECT().GetMarketsRegionIdOrders(gomock.Any(), "all", int32(12346), pageFive).Return(
		[]goesiv1.GetMarketsRegionIdOrders200Ok{}, nil, nil,
	).MaxTimes(1)
	mockOrderFetcher.EXPECT().GetMarketsRegionIdOrders(gomock.Any(), "all", int32(12346), pageSix).Return(
		[]goesiv1.GetMarketsRegionIdOrders200Ok{}, nil, nil,
	).MaxTimes(1)

	newUUID = func() (string, error) {
		return "123456", nil
	}

	mockPublisher.EXPECT().PublishStateBegin(int32(12345))
	mockPublisher.EXPECT().PublishOrder(&OrderPayload{
		RegionId:       12345,
		FetchRequestId: "123456",
		OrderId:        orderOne.OrderId,
		TypeId:         orderOne.TypeId,
		LocationId:     orderOne.LocationId,
		SystemId:       orderOne.SystemId,
		VolumeTotal:    orderOne.VolumeTotal,
		VolumeRemain:   orderOne.VolumeRemain,
		MinVolume:      orderOne.MinVolume,
		Price:          orderOne.Price,
		IsBuyOrder:     orderOne.IsBuyOrder,
		Duration:       orderOne.Duration,
		Issued:         orderOne.Issued,
		Range_:         orderOne.Range_,
	})
	mockPublisher.EXPECT().PublishOrder(&OrderPayload{
		RegionId:       12345,
		FetchRequestId: "123456",
		OrderId:        orderTwo.OrderId,
		TypeId:         orderTwo.TypeId,
		LocationId:     orderTwo.LocationId,
		SystemId:       orderTwo.SystemId,
		VolumeTotal:    orderTwo.VolumeTotal,
		VolumeRemain:   orderTwo.VolumeRemain,
		MinVolume:      orderTwo.MinVolume,
		Price:          orderTwo.Price,
		IsBuyOrder:     orderTwo.IsBuyOrder,
		Duration:       orderTwo.Duration,
		Issued:         orderTwo.Issued,
		Range_:         orderTwo.Range_,
	})
	mockPublisher.EXPECT().PublishOrder(&OrderPayload{
		RegionId:       12345,
		FetchRequestId: "123456",
		OrderId:        orderThree.OrderId,
		TypeId:         orderThree.TypeId,
		LocationId:     orderThree.LocationId,
		SystemId:       orderThree.SystemId,
		VolumeTotal:    orderThree.VolumeTotal,
		VolumeRemain:   orderThree.VolumeRemain,
		MinVolume:      orderThree.MinVolume,
		Price:          orderThree.Price,
		IsBuyOrder:     orderThree.IsBuyOrder,
		Duration:       orderThree.Duration,
		Issued:         orderThree.Issued,
		Range_:         orderThree.Range_,
	})
	mockPublisher.EXPECT().PublishOrder(&OrderPayload{
		RegionId:       12345,
		FetchRequestId: "123456",
		OrderId:        orderFour.OrderId,
		TypeId:         orderFour.TypeId,
		LocationId:     orderFour.LocationId,
		SystemId:       orderFour.SystemId,
		VolumeTotal:    orderFour.VolumeTotal,
		VolumeRemain:   orderFour.VolumeRemain,
		MinVolume:      orderFour.MinVolume,
		Price:          orderFour.Price,
		IsBuyOrder:     orderFour.IsBuyOrder,
		Duration:       orderFour.Duration,
		Issued:         orderFour.Issued,
		Range_:         orderFour.Range_,
	})
	mockPublisher.EXPECT().PublishStateEnd(int32(12345))

	mockPublisher.EXPECT().PublishStateBegin(int32(12346))
	mockPublisher.EXPECT().PublishOrder(&OrderPayload{
		RegionId:       12346,
		FetchRequestId: "123456",
		OrderId:        orderOne.OrderId,
		TypeId:         orderOne.TypeId,
		LocationId:     orderOne.LocationId,
		SystemId:       orderOne.SystemId,
		VolumeTotal:    orderOne.VolumeTotal,
		VolumeRemain:   orderOne.VolumeRemain,
		MinVolume:      orderOne.MinVolume,
		Price:          orderOne.Price,
		IsBuyOrder:     orderOne.IsBuyOrder,
		Duration:       orderOne.Duration,
		Issued:         orderOne.Issued,
		Range_:         orderOne.Range_,
	})
	mockPublisher.EXPECT().PublishOrder(&OrderPayload{
		RegionId:       12346,
		FetchRequestId: "123456",
		OrderId:        orderTwo.OrderId,
		TypeId:         orderTwo.TypeId,
		LocationId:     orderTwo.LocationId,
		SystemId:       orderTwo.SystemId,
		VolumeTotal:    orderTwo.VolumeTotal,
		VolumeRemain:   orderTwo.VolumeRemain,
		MinVolume:      orderTwo.MinVolume,
		Price:          orderTwo.Price,
		IsBuyOrder:     orderTwo.IsBuyOrder,
		Duration:       orderTwo.Duration,
		Issued:         orderTwo.Issued,
		Range_:         orderTwo.Range_,
	})
	mockPublisher.EXPECT().PublishOrder(&OrderPayload{
		RegionId:       12346,
		FetchRequestId: "123456",
		OrderId:        orderThree.OrderId,
		TypeId:         orderThree.TypeId,
		LocationId:     orderThree.LocationId,
		SystemId:       orderThree.SystemId,
		VolumeTotal:    orderThree.VolumeTotal,
		VolumeRemain:   orderThree.VolumeRemain,
		MinVolume:      orderThree.MinVolume,
		Price:          orderThree.Price,
		IsBuyOrder:     orderThree.IsBuyOrder,
		Duration:       orderThree.Duration,
		Issued:         orderThree.Issued,
		Range_:         orderThree.Range_,
	})
	mockPublisher.EXPECT().PublishOrder(&OrderPayload{
		RegionId:       12346,
		FetchRequestId: "123456",
		OrderId:        orderFour.OrderId,
		TypeId:         orderFour.TypeId,
		LocationId:     orderFour.LocationId,
		SystemId:       orderFour.SystemId,
		VolumeTotal:    orderFour.VolumeTotal,
		VolumeRemain:   orderFour.VolumeRemain,
		MinVolume:      orderFour.MinVolume,
		Price:          orderFour.Price,
		IsBuyOrder:     orderFour.IsBuyOrder,
		Duration:       orderFour.Duration,
		Issued:         orderFour.Issued,
		Range_:         orderFour.Range_,
	})
	mockPublisher.EXPECT().PublishStateEnd(int32(12346))
	//END Expectations

	controller, err := NewController(mockRegionFetcher, mockOrderFetcher, mockPublisher, 2, 4, nil, time.Millisecond*250)

	if err != nil {
		t.Error("Received an error when none were expected")
	}

	//We're reading from a channel, need a way to time the test out so we don't hang something up
	//https://blog.golang.org/go-concurrency-patterns-timing-out-and
	timeout := make(chan bool, 1)
	go func() {
		time.Sleep(5 * time.Second)
		timeout <- true
	}()

	done := controller.GetDoneChannel()

	//This doesn't have to be executed in it's own goroutine but this make it easier in the test
	controller.Fetch(12345)
	controller.Fetch(12346)

	for idx := 0; idx < 2; idx++ {
		validRegion := false
		regionId := int32(0)

		select {
		case <-timeout:
			t.Fatal("Test timed out waiting for done signal")
		case regionId = <-done:
			if regionId == int32(12345) || regionId == int32(12346) {
				validRegion = true
			}
		}

		if !validRegion {
			t.Errorf("Expected region id: (%d or %d) but received (%d)", 12345, 12346, regionId)
		}
	}

	controller.Stop()
	fmt.Println("Test: Done received")
}

func TestNewController_Error(t *testing.T) {
	_, err := NewController(nil, nil, nil, 0, 0, nil, time.Millisecond*250)

	if err == nil {
		t.Error("Should have received an error due to pool workers")
	}
}

func TestOrderController_Fetch_PublisherBindingLockCondition(t *testing.T) {
	t.SkipNow()
	mockCtrl := gomock.NewController(t)
	mockRegionFetcher := NewMockRegionsFetcher(mockCtrl)
	mockOrderFetcher := NewMockOrderFetcher(mockCtrl)
	mockPublisher := NewMockOrderPublisher(mockCtrl)
	defer mockCtrl.Finish()

	//BGN Expectations
	mockRegionFetcher.EXPECT().GetUniverseRegions(gomock.Any(), gomock.Nil()).Return([]int32{12345}, nil, nil).MaxTimes(1)
	mockRegionFetcher.EXPECT().GetUniverseRegionsRegionId(gomock.Any(), int32(12345), gomock.Nil()).Return(goesiv1.GetUniverseRegionsRegionIdOk{
		Name:           "The Forge",
		Constellations: []int32{},
		Description:    "Herpa derpa blerpa Jita",
		RegionId:       int32(12345),
	}, nil, nil).MaxTimes(1)

	orderOne := goesiv1.GetMarketsRegionIdOrders200Ok{
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
	}
	orderTwo := goesiv1.GetMarketsRegionIdOrders200Ok{
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
	}
	orderThree := goesiv1.GetMarketsRegionIdOrders200Ok{
		VolumeTotal:  20,
		VolumeRemain: 20,
		TypeId:       1,
		Range_:       "region",
		Price:        1.1,
		OrderId:      int64(3),
		MinVolume:    2,
		LocationId:   12345678,
		Issued:       time.Now(),
		IsBuyOrder:   false,
		Duration:     40,
	}
	orderFour := goesiv1.GetMarketsRegionIdOrders200Ok{
		VolumeTotal:  20,
		VolumeRemain: 20,
		TypeId:       1,
		Range_:       "region",
		Price:        1.1,
		OrderId:      int64(4),
		MinVolume:    2,
		LocationId:   123456789,
		Issued:       time.Now(),
		IsBuyOrder:   true,
		Duration:     40,
	}

	pageOne := make(map[string]interface{})
	pageOne["page"] = int32(1)

	mockOrderFetcher.EXPECT().GetMarketsRegionIdOrders(gomock.Any(), "all", int32(12345), pageOne).Return(
		[]goesiv1.GetMarketsRegionIdOrders200Ok{orderOne, orderTwo}, nil, nil,
	).MaxTimes(1)

	pageTwo := make(map[string]interface{})
	pageTwo["page"] = int32(2)

	mockOrderFetcher.EXPECT().GetMarketsRegionIdOrders(gomock.Any(), "all", int32(12345), pageTwo).Return(
		[]goesiv1.GetMarketsRegionIdOrders200Ok{orderThree, orderFour}, nil, nil,
	).MaxTimes(1)

	//We have to allow both pages 3 and 4 because this interface is stupid and this case is spawning workers at a time
	//I guess CCP don't really want people threading these requests easily?
	pageThree := make(map[string]interface{})
	pageThree["page"] = int32(3)

	mockOrderFetcher.EXPECT().GetMarketsRegionIdOrders(gomock.Any(), "all", int32(12345), pageThree).Return(
		[]goesiv1.GetMarketsRegionIdOrders200Ok{}, nil, nil,
	).MaxTimes(1)

	pageFour := make(map[string]interface{})
	pageFour["page"] = int32(4)

	mockOrderFetcher.EXPECT().GetMarketsRegionIdOrders(gomock.Any(), "all", int32(12345), pageFour).Return(
		[]goesiv1.GetMarketsRegionIdOrders200Ok{}, nil, nil,
	).MaxTimes(1)

	mockPublisher.EXPECT().PublishStateBegin(int32(12345)).MaxTimes(1)
	mockPublisher.EXPECT().PublishOrder(&OrderPayload{
		RegionId:     12345,
		OrderId:      orderOne.OrderId,
		TypeId:       orderOne.TypeId,
		LocationId:   orderOne.LocationId,
		SystemId:     orderOne.SystemId,
		VolumeTotal:  orderOne.VolumeTotal,
		VolumeRemain: orderOne.VolumeRemain,
		MinVolume:    orderOne.MinVolume,
		Price:        orderOne.Price,
		IsBuyOrder:   orderOne.IsBuyOrder,
		Duration:     orderOne.Duration,
		Issued:       orderOne.Issued,
		Range_:       orderOne.Range_,
	}).MaxTimes(1)
	mockPublisher.EXPECT().PublishOrder(&OrderPayload{
		RegionId:     12345,
		OrderId:      orderTwo.OrderId,
		TypeId:       orderTwo.TypeId,
		LocationId:   orderTwo.LocationId,
		SystemId:     orderTwo.SystemId,
		VolumeTotal:  orderTwo.VolumeTotal,
		VolumeRemain: orderTwo.VolumeRemain,
		MinVolume:    orderTwo.MinVolume,
		Price:        orderTwo.Price,
		IsBuyOrder:   orderTwo.IsBuyOrder,
		Duration:     orderTwo.Duration,
		Issued:       orderTwo.Issued,
		Range_:       orderTwo.Range_,
	}).MaxTimes(1)
	mockPublisher.EXPECT().PublishOrder(&OrderPayload{
		RegionId:     12345,
		OrderId:      orderThree.OrderId,
		TypeId:       orderThree.TypeId,
		LocationId:   orderThree.LocationId,
		SystemId:     orderThree.SystemId,
		VolumeTotal:  orderThree.VolumeTotal,
		VolumeRemain: orderThree.VolumeRemain,
		MinVolume:    orderThree.MinVolume,
		Price:        orderThree.Price,
		IsBuyOrder:   orderThree.IsBuyOrder,
		Duration:     orderThree.Duration,
		Issued:       orderThree.Issued,
		Range_:       orderThree.Range_,
	}).MaxTimes(1)
	mockPublisher.EXPECT().PublishOrder(&OrderPayload{
		RegionId:     12345,
		OrderId:      orderFour.OrderId,
		TypeId:       orderFour.TypeId,
		LocationId:   orderFour.LocationId,
		SystemId:     orderFour.SystemId,
		VolumeTotal:  orderFour.VolumeTotal,
		VolumeRemain: orderFour.VolumeRemain,
		MinVolume:    orderFour.MinVolume,
		Price:        orderFour.Price,
		IsBuyOrder:   orderFour.IsBuyOrder,
		Duration:     orderFour.Duration,
		Issued:       orderFour.Issued,
		Range_:       orderFour.Range_,
	}).MaxTimes(1)
	mockPublisher.EXPECT().PublishStateEnd(int32(12345)).MaxTimes(1)
	//END Expectations

	controller, err := NewController(mockRegionFetcher, mockOrderFetcher, mockPublisher, 1, 4, nil, time.Millisecond)

	if err != nil {
		t.Error("Received an error when none were expected")
	}

	//We're reading from a channel, need a way to time the test out so we don't hang something up
	//https://blog.golang.org/go-concurrency-patterns-timing-out-and
	timeout := make(chan bool, 1)
	go func() {
		time.Sleep(3 * time.Second)
		fmt.Println("Timeout routine timed out")
		timeout <- true
	}()

	ctrl := controller.(*orderController)

	done := ctrl.publishingBinder.clientDone

	//This doesn't have to be executed in it's own goroutine but this make it easier in the test
	go controller.Fetch(12345)

	time.Sleep(time.Second * 2)

	controller.Stop()

	select {
	case <-done:
		fmt.Println("Something told me I was done... that shouldn't happen")
		t.Fatal("I shouldn't have been notified of this!")
	case <-timeout:
		fmt.Println("Good test!")
	}

	fmt.Println("Finished blocking scenario")
}

func TestOrderController_Fetch_RegionError(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	mockRegionFetcher := NewMockRegionsFetcher(mockCtrl)
	mockOrderFetcher := NewMockOrderFetcher(mockCtrl)
	mockPublisher := NewMockOrderPublisher(mockCtrl)
	defer mockCtrl.Finish()

	//BGN Expectations
	mockRegionFetcher.EXPECT().GetUniverseRegionsRegionId(gomock.Any(), int32(1237821798), gomock.Nil()).Return(goesiv1.GetUniverseRegionsRegionIdOk{}, nil, errors.New("I'm sorry Dave, I'm afraid I can't do that"))

	controller, err := NewController(mockRegionFetcher, mockOrderFetcher, mockPublisher, 5, 10, nil, time.Millisecond*250)

	if err != nil {
		t.Error("Received an error when none were expected")
	}

	err = controller.Fetch(1237821798)

	if err == nil {
		t.Error("Did NOT receive an error when one was expected")
	}
}

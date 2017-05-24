package fetching

import (
	"errors"
	"fmt"
	"github.com/abaeve/pricing-fetcher/mocks"
	"github.com/antihax/goesi/v1"
	"github.com/golang/mock/gomock"
	"testing"
	"time"
)

func TestNewController(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	mockRegionFetcher := mocks.NewMockRegionsFetcher(mockCtrl)
	mockOrderFetcher := mocks.NewMockOrderFetcher(mockCtrl)
	mockPublisher := mocks.NewMockOrderPublisher(mockCtrl)
	defer mockCtrl.Finish()

	var controller *orderController

	ctrler, err := NewController(mockRegionFetcher, mockOrderFetcher, mockPublisher, 5, 10, nil)

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
	mockRegionFetcher := mocks.NewMockRegionsFetcher(mockCtrl)
	mockOrderFetcher := mocks.NewMockOrderFetcher(mockCtrl)
	mockPublisher := mocks.NewMockOrderPublisher(mockCtrl)
	defer mockCtrl.Finish()

	//BGN Expectations
	mockRegionFetcher.EXPECT().GetUniverseRegions(gomock.Nil()).Return([]int32{12345}, nil, nil).MaxTimes(1)
	mockRegionFetcher.EXPECT().GetUniverseRegionsRegionId(int32(12345), gomock.Nil()).Return(goesiv1.GetUniverseRegionsRegionIdOk{
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

	pageOne := make(map[string]interface{})
	pageOne["page"] = 1

	mockOrderFetcher.EXPECT().GetMarketsRegionIdOrders("all", int32(12345), pageOne).Return(
		[]goesiv1.GetMarketsRegionIdOrders200Ok{orderOne, orderTwo}, nil, nil,
	)

	pageTwo := make(map[string]interface{})
	pageTwo["page"] = 2

	mockOrderFetcher.EXPECT().GetMarketsRegionIdOrders("all", int32(12345), pageTwo).Return(
		[]goesiv1.GetMarketsRegionIdOrders200Ok{orderThree, orderFour}, nil, nil,
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

	mockPublisher.EXPECT().PublishStateBegin(int32(12345))
	mockPublisher.EXPECT().PublishOrder(&orderOne)
	mockPublisher.EXPECT().PublishOrder(&orderTwo)
	mockPublisher.EXPECT().PublishOrder(&orderThree)
	mockPublisher.EXPECT().PublishOrder(&orderFour)
	mockPublisher.EXPECT().PublishStateEnd(int32(12345))
	//END Expectations

	controller, err := NewController(mockRegionFetcher, mockOrderFetcher, mockPublisher, 5, 10, nil)

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

func TestNewController_Error(t *testing.T) {
	_, err := NewController(nil, nil, nil, 0, 0, nil)

	if err == nil {
		t.Error("Should have received an error due to pool workers")
	}
}

func TestOrderController_Fetch_PublishierBindingLockCondition(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	mockRegionFetcher := mocks.NewMockRegionsFetcher(mockCtrl)
	mockOrderFetcher := mocks.NewMockOrderFetcher(mockCtrl)
	mockPublisher := mocks.NewMockOrderPublisher(mockCtrl)
	defer mockCtrl.Finish()

	//BGN Expectations
	mockRegionFetcher.EXPECT().GetUniverseRegions(gomock.Nil()).Return([]int32{12345}, nil, nil).MaxTimes(1)
	mockRegionFetcher.EXPECT().GetUniverseRegionsRegionId(int32(12345), gomock.Nil()).Return(goesiv1.GetUniverseRegionsRegionIdOk{
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
	pageOne["page"] = 1

	mockOrderFetcher.EXPECT().GetMarketsRegionIdOrders("all", int32(12345), pageOne).Return(
		[]goesiv1.GetMarketsRegionIdOrders200Ok{orderOne, orderTwo}, nil, nil,
	).MaxTimes(1)

	pageTwo := make(map[string]interface{})
	pageTwo["page"] = 2

	mockOrderFetcher.EXPECT().GetMarketsRegionIdOrders("all", int32(12345), pageTwo).Return(
		[]goesiv1.GetMarketsRegionIdOrders200Ok{orderThree, orderFour}, nil, nil,
	).MaxTimes(1)

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

	mockPublisher.EXPECT().PublishStateBegin(int32(12345)).MaxTimes(1)
	mockPublisher.EXPECT().PublishOrder(&orderOne).MaxTimes(1)
	mockPublisher.EXPECT().PublishOrder(&orderTwo).MaxTimes(1)
	mockPublisher.EXPECT().PublishOrder(&orderThree).MaxTimes(1)
	mockPublisher.EXPECT().PublishOrder(&orderFour).MaxTimes(1)
	mockPublisher.EXPECT().PublishStateEnd(int32(12345)).MaxTimes(1)
	//END Expectations

	controller, err := NewController(mockRegionFetcher, mockOrderFetcher, mockPublisher, 5, 10, nil)

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

	ctrl := controller.(*orderController)

	done := ctrl.publishingBinder.clientDone

	//This doesn't have to be executed in it's own goroutine but this make it easier in the test
	go controller.Fetch(12345)

	time.Sleep(time.Millisecond * 500)

	controller.Stop()

	select {
	case <-done:
		t.Fatal("I shouldn't have been notified of this!")
	case <-timeout:
	}
}

func TestOrderController_Fetch_RegionError(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	mockRegionFetcher := mocks.NewMockRegionsFetcher(mockCtrl)
	mockOrderFetcher := mocks.NewMockOrderFetcher(mockCtrl)
	mockPublisher := mocks.NewMockOrderPublisher(mockCtrl)
	defer mockCtrl.Finish()

	//BGN Expectations
	mockRegionFetcher.EXPECT().GetUniverseRegionsRegionId(int32(1237821798), gomock.Nil()).Return(goesiv1.GetUniverseRegionsRegionIdOk{}, nil, errors.New("I'm sorry Dave, I'm afraid I can't do that"))

	controller, err := NewController(mockRegionFetcher, mockOrderFetcher, mockPublisher, 5, 10, nil)

	if err != nil {
		t.Error("Received an error when none were expected")
	}

	err = controller.Fetch(1237821798)

	if err == nil {
		t.Error("Did NOT receive an error when one was expected")
	}
}

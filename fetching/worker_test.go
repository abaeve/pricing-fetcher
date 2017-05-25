package fetching

import (
	"errors"
	"github.com/antihax/goesi/v1"
	"github.com/golang/mock/gomock"
	"testing"
	"time"
)

func TestWorker_Work(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	mockOrderFetcher := NewMockOrderFetcher(mockCtrl)
	defer mockCtrl.Finish()

	//We're reading from a channel, need a way to time the test out so we don't hang something up
	//https://blog.golang.org/go-concurrency-patterns-timing-out-and
	timeout := make(chan bool, 1)
	go func() {
		time.Sleep(1 * time.Second)
		timeout <- true
	}()

	options := make(map[string]interface{})
	options["page"] = 1

	mockOrderFetcher.EXPECT().GetMarketsRegionIdOrders("orderChan", int32(123456), options).Return(
		[]goesiv1.GetMarketsRegionIdOrders200Ok{
			{
				Duration:     int32(1),
				IsBuyOrder:   true,
				Issued:       time.Now(),
				LocationId:   123456,
				MinVolume:    2,
				OrderId:      1000000,
				Price:        1.1,
				Range_:       "region",
				TypeId:       54,
				VolumeRemain: 20000,
				VolumeTotal:  20000,
			},
		}, nil, nil,
	)

	out := make(chan OrderPayload)
	endReached := make(chan bool)
	workerDone := make(chan int)

	fetcher := NewWorker(mockOrderFetcher, "orderChan", 1, 123456, out, endReached, workerDone)

	//This is designed to work in a go routine, so do things in a go routine!
	go func() {
		fetcher.Work(1)
	}()

	var result OrderPayload
	var workerFinished int

	for {
		select {
		case result = <-out:
		case workerFinished = <-workerDone:
		case <-timeout:
			t.Fatal("Test timed out")
		}

		if workerFinished == 1 {
			break
		}
	}

	if result.VolumeTotal != 20000 {
		t.Errorf("Expected (%d) but received (%d)", 20000, result.VolumeTotal)
	}
}

func TestWorker_Work_Error(t *testing.T) {
	//This is the actual assertion
	//If the call to Work tries to publish to the closed out channel it will panic, this will catch
	//that panic fail the test
	//https://github.com/golang/go/wiki/PanicAndRecover
	defer func() {
		if r := recover(); r != nil {
			t.Fatal("Shouldn't have panic'ed")
		}
	}()

	mockCtrl := gomock.NewController(t)
	mockOrderFetcher := NewMockOrderFetcher(mockCtrl)
	defer mockCtrl.Finish()

	options := make(map[string]interface{})
	options["page"] = 1

	mockOrderFetcher.EXPECT().GetMarketsRegionIdOrders("orderChan", int32(123456), options).Return(
		[]goesiv1.GetMarketsRegionIdOrders200Ok{
			//Return an actual object just as some extra piece of mind that the error processing is happening as expected
			{
				Duration:     int32(1),
				IsBuyOrder:   true,
				Issued:       time.Now(),
				LocationId:   123456,
				MinVolume:    2,
				OrderId:      1000000,
				Price:        1.1,
				Range_:       "region",
				TypeId:       54,
				VolumeRemain: 20000,
				VolumeTotal:  20000,
			},
		}, nil, errors.New("I'm sorry Dave, I'm afraid I can't do that"),
	)

	out := make(chan OrderPayload)
	endReached := make(chan bool)
	workerDone := make(chan int)

	fetcher := NewWorker(mockOrderFetcher, "orderChan", 1, 123456, out, endReached, workerDone)

	//Close the outbound channel because we want the work function to panic if it publishes something
	close(out)

	fetcher.Work(1)
}

func TestWorker_Work_NoResults(t *testing.T) {
	//This is an assertion
	//If the call to Work tries to publish to the closed out channel it will panic, this will catch
	//that panic fail the test
	//https://github.com/golang/go/wiki/PanicAndRecover
	defer func() {
		if r := recover(); r != nil {
			t.Fatal("Shouldn't have panic'ed")
		}
	}()

	//We're reading from a channel, need a way to time the test out so we don't hang something up
	//https://blog.golang.org/go-concurrency-patterns-timing-out-and
	timeout := make(chan bool, 1)
	go func() {
		time.Sleep(1 * time.Second)
		timeout <- true
	}()

	mockCtrl := gomock.NewController(t)
	mockOrderFetcher := NewMockOrderFetcher(mockCtrl)
	defer mockCtrl.Finish()

	options := make(map[string]interface{})
	options["page"] = 1

	mockOrderFetcher.EXPECT().GetMarketsRegionIdOrders("orderChan", int32(123456), options).Return(
		[]goesiv1.GetMarketsRegionIdOrders200Ok{}, nil, nil,
	)

	out := make(chan OrderPayload)
	doneChan := make(chan bool)
	workerDone := make(chan int)

	fetcher := NewWorker(mockOrderFetcher, "orderChan", 1, 123456, out, doneChan, workerDone)

	//Close the outbound channel because we want the work function to panic if it publishes something to out
	close(out)

	go func() {
		fetcher.Work(1)
	}()

	var done bool

Done:
	for {
		select {
		case done = <-doneChan:
			break Done
		case <-timeout:
			t.Fatal("Test timed out")
		}
	}

	if !done {
		t.Error("Should have said we reached the end")
	}
}

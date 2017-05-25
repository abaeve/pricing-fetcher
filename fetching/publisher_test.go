package fetching

import (
	"encoding/json"
	"github.com/abaeve/pricing-fetcher/mocks"
	"github.com/antihax/goesi/v1"
	"github.com/golang/mock/gomock"
	"github.com/micro/go-micro/broker"
	"testing"
	"time"
)

func TestNewRabbitMQPublisher(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	mockRegionFetcher := NewMockRegionsFetcher(mockCtrl)
	mockBroker := mocks.NewMockBroker(mockCtrl)
	defer mockCtrl.Finish()

	publisher := NewPublisher(mockRegionFetcher, mockBroker).(orderPublisher)

	if mockBroker != publisher.broker {
		t.Errorf("Expected (%+v) but received (%+v) for broker address matching", mockBroker, publisher.broker)
	}
}

func TestOrderPublisher_PublishOrder(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	mockBroker := mocks.NewMockBroker(mockCtrl)
	mockRegionFetcher := NewMockRegionsFetcher(mockCtrl)
	defer mockCtrl.Finish()

	order := OrderPayload{
		RegionId: 123456,
		GetMarketsRegionIdOrders200Ok: goesiv1.GetMarketsRegionIdOrders200Ok{
			IsBuyOrder:   false,
			OrderId:      1,
			Duration:     2000,
			Issued:       time.Now(),
			LocationId:   123456,
			MinVolume:    20,
			Price:        2.1,
			Range_:       "region",
			TypeId:       1,
			VolumeRemain: 40,
			VolumeTotal:  80,
		},
	}

	payload, _ := json.Marshal(order)

	mockRegionFetcher.EXPECT().GetUniverseRegionsRegionId(int32(123456), gomock.Nil()).Return(
		goesiv1.GetUniverseRegionsRegionIdOk{
			RegionId:    123456,
			Description: "Herpa Derpa Blerpa doo",
			Name:        "someregion",
		}, nil, nil,
	)
	mockBroker.EXPECT().Publish("someregion.state.begin", &broker.Message{
		Body: []byte("Starting"),
	})
	mockBroker.EXPECT().Publish("sell.someregion", &broker.Message{
		Body: payload,
	})
	mockBroker.EXPECT().Publish("someregion.state.end", &broker.Message{
		Body: []byte("Ending"),
	})

	publisher := NewPublisher(mockRegionFetcher, mockBroker)

	publisher.PublishStateBegin(123456)
	publisher.PublishOrder(&order)
	publisher.PublishStateEnd(123456)
}

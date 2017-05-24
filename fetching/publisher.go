package fetching

import "github.com/antihax/goesi/v1"

type OrderPublisher interface {
	PublishOrder(order *goesiv1.GetMarketsRegionIdOrders200Ok)
	PublishStateBegin(regionId int32)
	PublishStateEnd(regionId int32)
}

type orderPublisher struct {
}

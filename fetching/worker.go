package fetching

import (
	"github.com/antihax/goesi/v1"
	"github.com/goinggo/work"
)

type orderFetcher struct {
	client    OrderFetcher
	orderType string
	page      int
	regionId  int32
	out       chan goesiv1.GetMarketsRegionIdOrders200Ok
	done      chan bool
}

func (w *orderFetcher) Work(id int) {
	options := make(map[string]interface{})
	options["page"] = w.page
	data, _, err := w.client.GetMarketsRegionIdOrders(w.orderType, w.regionId, options)

	if err != nil {
		return
	}

	if len(data) == 0 {
		w.done <- true
		return
	}

	for _, order := range data {
		w.out <- order
	}
}

func NewWorker(client OrderFetcher, orderType string, page int, regionId int32, out chan goesiv1.GetMarketsRegionIdOrders200Ok, done chan bool) work.Worker {
	return &orderFetcher{client: client, orderType: orderType, page: page, regionId: regionId, out: out, done: done}
}

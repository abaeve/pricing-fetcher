package fetching

import (
	"github.com/antihax/goesi/v1"
	"github.com/goinggo/work"
)

type orderFetcher struct {
	client     OrderFetcher
	orderType  string
	page       int
	regionId   int32
	out        chan goesiv1.GetMarketsRegionIdOrders200Ok
	endReached chan bool
	workerDone chan bool
}

func (w *orderFetcher) Work(id int) {
	options := make(map[string]interface{})
	options["page"] = w.page
	data, _, err := w.client.GetMarketsRegionIdOrders(w.orderType, w.regionId, options)

	if err != nil {
		return
	}

	if len(data) == 0 {
		w.endReached <- true
	}

	for _, order := range data {
		w.out <- order
	}

	w.workerDone <- true
}

func NewWorker(client OrderFetcher, orderType string, page int, regionId int32, out chan goesiv1.GetMarketsRegionIdOrders200Ok, done chan bool, workerDone chan bool) work.Worker {
	return &orderFetcher{client: client, orderType: orderType, page: page, regionId: regionId, out: out, endReached: done, workerDone: workerDone}
}

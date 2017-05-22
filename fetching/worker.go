package fetching

import (
	"fmt"
	"github.com/antihax/goesi/v1"
	"github.com/goinggo/work"
)

type orderFetcher struct {
	client     OrderFetcher
	orderType  string
	page       int
	regionId   int32
	out        chan<- goesiv1.GetMarketsRegionIdOrders200Ok
	endReached chan<- bool
	workerDone chan<- int
}

func (w *orderFetcher) Work(id int) {
	options := make(map[string]interface{})
	options["page"] = w.page
	data, _, err := w.client.GetMarketsRegionIdOrders(w.orderType, w.regionId, options)

	if err != nil {
		return
	}

	if len(data) == 0 {
		fmt.Printf("Worker: Worker %d publishing endReached\n", w.page)
		defer fmt.Printf("Worker: Worker %d returning\n", w.page)
		w.endReached <- true
		return
	}

	for _, order := range data {
		w.out <- order
	}

	fmt.Printf("Worker: Worker %d publishing done\n", w.page)
	defer fmt.Printf("Worker: Worker %d returning\n", w.page)
	w.workerDone <- w.page
}

func NewWorker(client OrderFetcher, orderType string, page int, regionId int32, out chan<- goesiv1.GetMarketsRegionIdOrders200Ok, endReached chan<- bool, workerDone chan<- int) work.Worker {
	return &orderFetcher{client: client, orderType: orderType, page: page, regionId: regionId, out: out, endReached: endReached, workerDone: workerDone}
}

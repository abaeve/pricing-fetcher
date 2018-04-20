package fetching

import (
	"fmt"
	goesiv1 "github.com/antihax/goesi/esi"
	"github.com/antihax/goesi/optional"
	"github.com/goinggo/work"
	"golang.org/x/net/context"
)

type OrderPayload struct {
	goesiv1.GetMarketsRegionIdOrders200Ok

	FetchRequestId string
	RegionId       int32
}

type orderFetcher struct {
	client         OrderFetcher
	orderType      string
	page           int32
	regionId       int32
	out            chan<- OrderPayload
	endReached     chan<- bool
	workerDone     chan<- int32
	fetchRequestId string
}

func (w *orderFetcher) Work(id int) {
	options := &goesiv1.GetMarketsRegionIdOrdersOpts{}
	options.Page = optional.NewInt32(w.page)
	data, _, err := w.client.GetMarketsRegionIdOrders(context.Background(), w.orderType, w.regionId, options)

	if err != nil {
		fmt.Printf("Worker: Worker %d encountered an error\n", w.page)
		w.workerDone <- w.page
		return
	}

	if len(data) == 0 {
		fmt.Printf("Worker: Worker %d publishing endReached\n", w.page)
		defer fmt.Printf("Worker: Worker %d returning\n", w.page)
		w.endReached <- true
		return
	}

	for _, order := range data {
		w.out <- OrderPayload{
			RegionId:                      w.regionId,
			FetchRequestId:                w.fetchRequestId,
			GetMarketsRegionIdOrders200Ok: order,
		}
	}

	fmt.Printf("Worker: Worker %d publishing done\n", w.page)
	defer fmt.Printf("Worker: Worker %d returning\n", w.page)
	w.workerDone <- w.page
}

func NewWorker(client OrderFetcher, orderType string, page int32, regionId int32, out chan<- OrderPayload, endReached chan<- bool, workerDone chan<- int32, fetchRequestId string) work.Worker {
	return &orderFetcher{client: client, orderType: orderType, page: page, regionId: regionId, out: out, endReached: endReached, workerDone: workerDone, fetchRequestId: fetchRequestId}
}

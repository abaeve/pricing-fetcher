package fetching

import (
	"fmt"
	goesiv1 "github.com/antihax/goesi/esi"
	"github.com/antihax/goesi/optional"
	"github.com/goinggo/work"
	"golang.org/x/net/context"
	"time"
)

type OrderPayload struct {
	FetchRequestId string `json:"request_id,omitempty"` /* request_id string */
	RegionId       int32  `json:"region_id,omitempty"`  /* region_id integer */

	OrderId      int64     `json:"order_id,omitempty"`      /* order_id integer */
	TypeId       int32     `json:"type_id,omitempty"`       /* type_id integer */
	LocationId   int64     `json:"location_id,omitempty"`   /* location_id integer */
	SystemId     int32     `json:"system_id,omitempty"`     /* The solar system this order was placed */
	VolumeTotal  int32     `json:"volume_total,omitempty"`  /* volume_total integer */
	VolumeRemain int32     `json:"volume_remain,omitempty"` /* volume_remain integer */
	MinVolume    int32     `json:"min_volume,omitempty"`    /* min_volume integer */
	Price        float64   `json:"price,omitempty"`         /* price number */
	IsBuyOrder   bool      `json:"is_buy_order,omitempty"`  /* is_buy_order boolean */
	Duration     int32     `json:"duration,omitempty"`      /* duration integer */
	Issued       time.Time `json:"issued,omitempty"`        /* issued string */
	Range_       string    `json:"range,omitempty"`         /* range string */
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
			RegionId:       w.regionId,
			FetchRequestId: w.fetchRequestId,
			OrderId:        order.OrderId,
			TypeId:         order.TypeId,
			LocationId:     order.LocationId,
			SystemId:       order.SystemId,
			VolumeTotal:    order.VolumeTotal,
			VolumeRemain:   order.VolumeRemain,
			MinVolume:      order.MinVolume,
			Price:          order.Price,
			IsBuyOrder:     order.IsBuyOrder,
			Duration:       order.Duration,
			Issued:         order.Issued,
			Range_:         order.Range_,
		}
	}

	fmt.Printf("Worker: Worker %d publishing done\n", w.page)
	defer fmt.Printf("Worker: Worker %d returning\n", w.page)
	w.workerDone <- w.page
}

func NewWorker(client OrderFetcher, orderType string, page int32, regionId int32, out chan<- OrderPayload, endReached chan<- bool, workerDone chan<- int32, fetchRequestId string) work.Worker {
	return &orderFetcher{client: client, orderType: orderType, page: page, regionId: regionId, out: out, endReached: endReached, workerDone: workerDone, fetchRequestId: fetchRequestId}
}

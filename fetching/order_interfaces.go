package fetching

import (
	"github.com/antihax/goesi/v1"
	"net/http"
)

type OrderFetcher interface {
	GetMarketsRegionIdOrders(orderType string, regionId int32, localVarOptionals map[string]interface{}) ([]goesiv1.GetMarketsRegionIdOrders200Ok, *http.Response, error)
}

type HistoryFetcher interface {
	GetMarketsRegionIdHistory(regionId int32, typeId int32, localVarOptionals map[string]interface{}) ([]goesiv1.GetMarketsRegionIdHistory200Ok, *http.Response, error)
}

package fetching

import (
	"github.com/antihax/goesi/esi"
	"golang.org/x/net/context"
	"net/http"
)

type RegionsFetcher interface {
	GetUniverseRegions(ctx context.Context, localVarOptionals map[string]interface{}) ([]int32, *http.Response, error)
	GetUniverseRegionsRegionId(ctx context.Context, regionId int32, localVarOptionals map[string]interface{}) (esi.GetUniverseRegionsRegionIdOk, *http.Response, error)
}

type OrderFetcher interface {
	GetMarketsRegionIdOrders(ctx context.Context, orderType string, regionId int32, localVarOptionals map[string]interface{}) ([]esi.GetMarketsRegionIdOrders200Ok, *http.Response, error)
}

type HistoryFetcher interface {
	GetMarketsRegionIdHistory(ctx context.Context, regionId int32, typeId int32, localVarOptionals map[string]interface{}) ([]esi.GetMarketsRegionIdHistory200Ok, *http.Response, error)
}

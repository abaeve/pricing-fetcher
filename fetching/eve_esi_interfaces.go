package fetching

import (
	"github.com/antihax/goesi/esi"
	"golang.org/x/net/context"
	"net/http"
)

type RegionsFetcher interface {
	GetUniverseRegions(ctx context.Context, opts *esi.GetUniverseRegionsOpts) ([]int32, *http.Response, error)
	GetUniverseRegionsRegionId(ctx context.Context, regionId int32, opts *esi.GetUniverseRegionsRegionIdOpts) (esi.GetUniverseRegionsRegionIdOk, *http.Response, error)
}

type OrderFetcher interface {
	GetMarketsRegionIdOrders(ctx context.Context, orderType string, regionId int32, localVarOptionals *esi.GetMarketsRegionIdOrdersOpts) ([]esi.GetMarketsRegionIdOrders200Ok, *http.Response, error)
}

type HistoryFetcher interface {
	GetMarketsRegionIdHistory(ctx context.Context, regionId int32, typeId int32, localVarOptionals *esi.GetMarketsRegionIdHistoryOpts) ([]esi.GetMarketsRegionIdHistory200Ok, *http.Response, error)
}

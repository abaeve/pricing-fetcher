package fetching

//go:generate mockgen -package=fetching -source=eve_esi_interfaces.go -destination=fetcher_mocks.go RegionsFetcher,OrderFetcher,HistoryFetcher,OrderPublisher
//go:generate mockgen -package=fetching -source=publisher.go -destination=publisher_mocks.go OrderPublisher

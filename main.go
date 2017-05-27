package main

import (
	"fmt"
	"github.com/abaeve/pricing-fetcher/fetching"
	"github.com/antihax/goesi"
	"github.com/gregjones/httpcache"
	"github.com/micro/go-plugins/broker/rabbitmq"
	"time"

	_ "net/http/pprof"
	"net/http"
	"log"
)

var controller fetching.OrderController

func init() {
	httpClient := httpcache.NewMemoryCacheTransport().Client()

	// Get the ESI API Client
	apiClient := goesi.NewAPIClient(httpClient, "aba-pricing-fetcher maurer.it@gmail.com https://github.com/abaeve/pricing-fetcher")

	broker := rabbitmq.NewBroker()
	broker.Init()
	broker.Connect()

	orderPublisher := fetching.NewPublisher(apiClient.V1.UniverseApi, broker)

	controller, _ = fetching.NewController(apiClient.V1.UniverseApi, apiClient.V1.MarketApi, orderPublisher, 4, 16, poolLog, time.Second)

	done := controller.GetDoneChannel()

	//I don't really care but reading from this channel is mandatory now
	go func() {
		for {
			<-done
			fmt.Println("Something finished")
		}
	}()

	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()
}

func main() {
	for idx := 0; idx < 6; idx++ {
		//The Forge
		controller.Fetch(10000002)
		//Domain
		controller.Fetch(10000043)
		//Sinq Laison
		controller.Fetch(10000032)
		//Heimatar
		controller.Fetch(10000030)
		time.Sleep(time.Minute * 5)
	}
}

func poolLog(message string) {
	fmt.Println(message)
}

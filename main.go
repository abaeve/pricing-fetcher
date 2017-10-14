package main

import (
	"fmt"
	"github.com/PuerkitoBio/rehttp"
	"github.com/abaeve/pricing-fetcher/fetching"
	"github.com/antihax/goesi"
	"github.com/micro/go-plugins/broker/rabbitmq"
	"time"

	//"log"
	"net/http"
	//_ "net/http/pprof"
)

var controller fetching.OrderController

func initialize() {
	httpClient := &http.Client{
		Transport: rehttp.NewTransport(
			nil,
			rehttp.RetryAll(rehttp.RetryMaxRetries(3), rehttp.RetryStatuses(500)),
			rehttp.ConstDelay(time.Second*3),
		),
	}

	// Get the ESI API Client
	apiClient := goesi.NewAPIClient(httpClient, "aba-pricing-fetcher maurer.it@gmail.com https://github.com/abaeve/pricing-fetcher")

	broker := rabbitmq.NewBroker()
	broker.Init()
	broker.Connect()

	orderPublisher := fetching.NewPublisher(apiClient.ESI.UniverseApi, broker)

	controller, _ = fetching.NewController(apiClient.ESI.UniverseApi, apiClient.ESI.MarketApi, orderPublisher, 4, 16, poolLog, time.Second)

	done := controller.GetDoneChannel()

	//I don't really care but reading from this channel is mandatory now
	go func() {
		for {
			<-done
			fmt.Println("Something finished")
		}
	}()

	//go func() {
	//	log.Println(http.ListenAndServe("localhost:6060", nil))
	//}()
}

func main() {
	initialize()

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

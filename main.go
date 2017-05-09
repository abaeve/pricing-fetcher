package pricing_fetcher

import (
	"github.com/abaeve/services-common/config"
)

const (
	version = "1.0.0"
)

func main() {
	service := config.NewService(version, "pricing-fetcher", initialize)

	service.Run()
}

func initialize(conf *config.Configuration) error {

}

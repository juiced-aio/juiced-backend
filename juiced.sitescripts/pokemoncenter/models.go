package pokemoncenter

import (
	"backend.juicedbot.io/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/juiced.sitescripts/base"
)

// Endpoints
const (
	BaseEndpoint = "https://www.pokemoncenter.com/"

	MonitorEndpoint          = "https://www.pokemoncenter.com/product/%s"
	AddToCartEndpoint        = "https://www.pokemoncenter.com/tpci-ecommweb-api/cart?type=product&format=zoom.nodatalinks"
	AddToCartRefererEndpoint = "https://www.pokemoncenter.com/product/%s"
)

// Monitor info
type Monitor struct {
	Monitor         base.Monitor
	SKUsSentToTask  []string
	RunningMonitors []string
	OutOfStockSKUs  []string
	SKUs            []string
	InStock         []PokemonCenterInStockData
	SKUWithInfo     map[string]entities.PokemonCenterSingleMonitorInfo
}

type PokemonCenterInStockData struct {
	SKU           string
	Price         float64
	ItemName      string
	AddToCartForm string
	ImageURL      string
}

type MonitorResponse struct {
	Props struct {
		IsServer     bool
		InitialState struct {
			Product struct {
				AddToCartForm string
				Availability  string
				Images        struct {
					Original string
				}
				ListPrice struct {
					Amount float64
				}
				Name string
			}
		}
		InitialProps struct {
		}
	}
}

// Task info
type Task struct {
	Task         base.Task
	CheckoutInfo CheckoutInfo
}

type CheckoutInfo struct {
	AddToCartForm string
	AuthId        string
	ImageURL      string
	Price         float64
	ItemName      string
	SKU           string
}

type AddToCartRequest struct {
	ProductUri    string `json:"productUri`
	Quantity      int    `json:"quantity`
	Configuration string `json:"configuration"`
}
type AddToCartResponse struct {
	Type     string `json:"type"`
	Quantity int    `json:"quantity"`
}

package enums

import "fmt"

// Retailer is a list of possible websites that Tasks and Monitors can run for
type Retailer = string

// Choose one per Task/Monitor
const (
	Amazon        Retailer = "Amazon"
	BestBuy       Retailer = "BestBuy"
	BigCartel     Retailer = "BigCartel"
	BoxLunch      Retailer = "BoxLunch"
	Disney        Retailer = "Disney"
	GameStop      Retailer = "GameStop"
	HotTopic      Retailer = "HotTopic"
	Newegg        Retailer = "Newegg"
	PokemonCenter Retailer = "PokemonCenter"
	Shopify       Retailer = "Shopify"
	SquareSpace   Retailer = "SquareSpace"
	Target        Retailer = "Target"
	Topps         Retailer = "Topps"
	Walmart       Retailer = "Walmart"
)

type ShopifyRetailer = string

const (
	GenericShopify  ShopifyRetailer = "GenericShopify"
	MattelCreations ShopifyRetailer = "Mattel Creations"
)

var ShopifySiteURLs = map[ShopifyRetailer]string{
	MattelCreations: "https://creations.mattel.com",
}

var validRetailers = []Retailer{
	Amazon,
	BestBuy,
	BoxLunch,
	Disney,
	GameStop,
	HotTopic,
	Newegg,
	PokemonCenter,
	Target,
	Topps,
	Walmart,

	GenericShopify,
	MattelCreations,
}

type InvalidRetailerError struct {
	Retailer string
}

func (e *InvalidRetailerError) Error() string {
	return fmt.Sprintf("invalid retailer: %s", e.Retailer)
}

func IsValidRetailer(retailer string) error {
	for _, validRetailer := range validRetailers {
		if retailer == validRetailer {
			return nil
		}
	}
	return &InvalidRetailerError{retailer}
}

package enums

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
	GenericShopify ShopifyRetailer = "GenericShopify"
	HotWheels      ShopifyRetailer = "HotWheels"
)

var ShopifySiteURLs = map[ShopifyRetailer]string{
	HotWheels: "https://creations.mattel.com",
}

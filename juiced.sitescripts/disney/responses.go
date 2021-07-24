package disney

type DisneyStockResponse struct {
	Product  ProductResponse  `json:"product"`
	ATCState ATCStateResponse `json:"addToCartButtonState"`
}

type ProductResponse struct {
	ID            string               `json:"id"`
	ProductName   string               `json:"productName"`
	ProductType   string               `json:"productType"`
	ProductURL    string               `json:"selectedProductUrl"`
	Images        ImagesResponse       `json:"images"`
	Price         PriceResponse        `json:"price"`
	Availability  AvailabilityResponse `json:"availability"`
	Available     bool                 `json:"available"`
	QuantityLimit int                  `json:"quantityLimit"`
	Variants      []VariantResponse    `json:"variationAttributes"`
}

type ImagesResponse struct {
	LargeImages   []ImageResponse `json:"large"`
	SmallImages   []ImageResponse `json:"small"`
	HighResImages []ImageResponse `json:"highRes"`
}

type ImageResponse struct {
	ImageURL string `json:"url"`
}

type PriceResponse struct {
	Sales SalesResponse `json:"sales"`
}

type SalesResponse struct {
	Price string `json:"decimalPrice"`
}

type AvailabilityResponse struct {
	IsPreOrder  bool `json:"isPreOrder"`
	IsBackOrder bool `json:"isBackOrder"`
}

type ATCStateResponse struct {
	IsDisabled bool `json:"isDisabled"`
	IsSoldOut  bool `json:"isSoldOut"`
}

type VariantResponse struct {
	Attribute string                 `json:"attributeId"`
	Values    []VariantValueResponse `json:"values"`
}

type VariantValueResponse struct {
	Value      string `json:"value"`
	Selectable bool   `json:"selectable"`
}

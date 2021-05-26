package target

import (
	"juiced/juiced.sitescripts/util"
	"net/http"
)

// AddTargetHeaders adds Target-specific headers to the request
func AddTargetHeaders(request *http.Request, referer ...string) {
	util.AddBaseHeaders(request)
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Origin", "https://www.target.com")
	if len(referer) != 0 {
		request.Header.Set("Referer", referer[0])
	}
}

// GetTCINStockRequestToMap maps a GetTCINStockRequest object to a map[string]string for converting to a query string
func GetTCINStockRequestToMap(getTCINStockRequest GetTCINStockRequest) map[string]string {
	return map[string]string{
		"key":                         getTCINStockRequest.Key,
		"tcins":                       getTCINStockRequest.TCINs,
		"store_id":                    getTCINStockRequest.StoreID,
		"scheduled_delivery_store_id": getTCINStockRequest.ScheduledDeliveryStoreID,
	}
}

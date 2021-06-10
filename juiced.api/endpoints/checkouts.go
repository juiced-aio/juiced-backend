package endpoints

import (
	"strconv"

	"backend.juicedbot.io/juiced.api/responses"
	"backend.juicedbot.io/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/juiced.infrastructure/common/errors"
	"backend.juicedbot.io/juiced.infrastructure/queries"

	"encoding/json"
	"net/http"
)

// GetAllCheckoutsEndpoint handles the GET request at /api/checkout
func GetAllCheckoutsEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	response.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")

	request.ParseForm()
	params := request.Form

	retailer := params.Get("retailer")
	days := params.Get("days")
	daysBack := -1
	var err error
	if days != "" {
		daysBack, err = strconv.Atoi(days)
	}

	errorsList := make([]string, 0)
	checkouts := make([]entities.Checkout, 0)
	if err == nil {
		checkouts, err = queries.GetCheckouts(retailer, daysBack)
		if err != nil {
			errorsList = append(errorsList, errors.GetAllCheckoutsError+err.Error())
		}
	} else {
		errorsList = append(errorsList, errors.GetAllCheckoutsError+err.Error())
	}

	result := &responses.CheckoutResponse{Success: true, Data: checkouts, Errors: make([]string, 0)}
	if len(errorsList) > 0 {
		response.WriteHeader(http.StatusBadRequest)
		result = &responses.CheckoutResponse{Success: false, Data: checkouts, Errors: errorsList}
	}
	json.NewEncoder(response).Encode(result)
}

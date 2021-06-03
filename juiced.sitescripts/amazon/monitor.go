package amazon

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"backend.juicedbot.io/juiced.sitescripts/util"

	"backend.juicedbot.io/juiced.client/http"
	"backend.juicedbot.io/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/juiced.infrastructure/common/enums"
	"backend.juicedbot.io/juiced.infrastructure/common/events"
	"backend.juicedbot.io/juiced.sitescripts/base"

	"github.com/anaskhan96/soup"
	browser "github.com/eddycjy/fake-useragent"
)

// Creating a channel to recieve the Amazon account because half of the monitoring is done with the account
var Accounts = make(chan AccChan)

// CreateAmazonMonitor takes a TaskGroup entity and turns it into a Amazon Monitor
func CreateAmazonMonitor(taskGroup *entities.TaskGroup, proxy entities.Proxy, eventBus *events.EventBus, singleMonitors []entities.AmazonSingleMonitorInfo) (Monitor, error) {
	storedAmazonMonitors := make(map[string]entities.AmazonSingleMonitorInfo)
	amazonMonitor := Monitor{}
	asins := []string{}

	for _, monitor := range singleMonitors {
		client, err := util.CreateClient(proxy)
		if err != nil {
			return amazonMonitor, err
		}
		storedAmazonMonitors[monitor.ASIN] = entities.AmazonSingleMonitorInfo{
			ASIN:        monitor.ASIN,
			OFID:        monitor.OFID,
			MaxPrice:    monitor.MaxPrice,
			MonitorType: monitor.MonitorType,
			Client:      client,
		}
		asins = append(asins, monitor.ASIN)
	}

	for created := false; !created; {
		account := <-Accounts
		// Making sure the accounts group id matches the monitors
		if account.GroupID != taskGroup.GroupID {
			continue
		}

		amazonMonitor = Monitor{
			Monitor: base.Monitor{
				TaskGroup: taskGroup,
				Proxy:     proxy,
				EventBus:  eventBus,
			},

			ASINs:         asins,
			ASINWithInfo:  storedAmazonMonitors,
			AccountClient: account.Client,
			AddressID:     account.AccountInfo.SavedAddressID,
			SessionID:     account.AccountInfo.SessionID,
		}
		created = true
	}
	becameGuest := false
	for !becameGuest {
		needToStop := amazonMonitor.CheckForStop()
		if needToStop {
			return amazonMonitor, nil
		}
		becameGuest = amazonMonitor.BecomeGuest()
		if !becameGuest {
			time.Sleep(1000 * time.Millisecond)
		}
	}

	return amazonMonitor, nil
}

// PublishEvent wraps the EventBus's PublishMonitorEvent function
func (monitor *Monitor) PublishEvent(status enums.MonitorStatus, eventType enums.MonitorEventType) {
	monitor.Monitor.TaskGroup.SetMonitorStatus(status)
	monitor.Monitor.EventBus.PublishMonitorEvent(status, eventType, nil, monitor.Monitor.TaskGroup.GroupID)
}

// CheckForStop checks the stop flag and stops the monitor if it's true
func (monitor *Monitor) CheckForStop() bool {
	if monitor.Monitor.StopFlag {
		monitor.PublishEvent(enums.MonitorIdle, enums.MonitorStop)
		return true
	}
	return false
}

//	Monitoring amazon is the hardest part of botting it with Amazon's Cloudfront BP,
//	I have recently modified the net/http library to perfectly replicate a chrome http2 fingerprint
//	which Cloudfront is recognizing and monitoring has become a lot easier.
//	Header-Order is also very important when getting around Cloudfront.
//	When setting headers normally they are stored in a map which is not
//	ordered. To order them with the modified net/http you will have to use the request.RawHeader.

// So theres a few different ways we can make the monitoring groups for Amazon, for now I'm going to make it so it runs a goroutine for each ASIN
func (monitor *Monitor) RunMonitor() {
	// If the function panics due to a runtime error, recover from it
	defer func() {
		recover()
		// TODO @silent: Let the UI know that a monitor failed
	}()

	if monitor.Monitor.TaskGroup.MonitorStatus == enums.MonitorIdle {
		monitor.PublishEvent(enums.WaitingForProductData, enums.MonitorStart)
	}
	needToStop := monitor.CheckForStop()
	if needToStop {
		return
	}
	for _, asin := range monitor.ASINs {
		if !util.InSlice(monitor.RunningMonitors, asin) {
			// TODO @Humphrey: THIS IS GOING TO CAUSE A MASSIVE MEMORY LEAK -- IF YOU HAVE 2 MONITORS, AND EACH ONE CALLS THE RUNMONITOR FUNCTION FROM WITHIN, YOU'LL START MULTIPLYING AND VERY QUICKLY YOU'LL HAVE THOUSANDS OF MONITORS
			// 		--> We should turn this into a RunSingleMonitor function, and have it call itself from within
			go func(t string) {
				// If the function panics due to a runtime error, recover from it
				defer func() {
					recover()
					// TODO @silent: Re-run this specific monitor
				}()

				somethingInStock := false
				switch monitor.ASINWithInfo[t].MonitorType {
				case enums.SlowSKUMonitor:
					somethingInStock = monitor.TurboMonitor(t)

				case enums.FastSKUMonitor:
					somethingInStock = monitor.OFIDMonitor(t)
				}

				if somethingInStock {
					needToStop := monitor.CheckForStop()
					if needToStop {
						return
					}
					monitor.RunningMonitors = util.RemoveFromSlice(monitor.RunningMonitors, t)
					monitor.PublishEvent(enums.SendingProductInfoToTasks, enums.MonitorUpdate)
					monitor.SendToTasks()
				} else {
					if len(monitor.RunningMonitors) > 0 {
						if monitor.Monitor.TaskGroup.MonitorStatus != enums.WaitingForInStock {
							monitor.PublishEvent(enums.WaitingForInStock, enums.MonitorUpdate)
						}
					}
					time.Sleep(time.Duration(monitor.Monitor.TaskGroup.MonitorDelay) * time.Millisecond)
					monitor.RunMonitor()
				}
			}(asin)
		}

	}

}

// A lot of the stuff that I'm doing either seems useless or dumb but Cloudfront is Ai based and the more entropy/randomness you add to every request
// the better.
func (monitor *Monitor) TurboMonitor(asin string) bool {
	var client http.Client
	currentEndpoint := AmazonEndpoints[util.RandomNumberInt(0, 2)]
	if currentEndpoint == "https://smile.amazon.com" {
		client = monitor.AccountClient
	} else {
		client = monitor.ASINWithInfo[asin].Client
	}
	resp, body, err := util.MakeRequest(&util.Request{
		Client: client,
		Method: "GET",
		URL:    currentEndpoint + fmt.Sprintf(MonitorEndpoints[util.RandomNumberInt(0, len(MonitorEndpoints))], asin) + util.Randomizer("&pldnSite=1"),
		RawHeaders: [][2]string{
			{"rtt", "100"},
			{"downlink", "10"},
			{"ect", "4g"},
			{"sec-ch-ua", `" Not A;Brand";v="99", "Chromium";v="90", "Google Chrome";v="90"`},
			{"sec-ch-ua-mobile", "?0"},
			{"upgrade-insecure-requests", "1"},
			{"user-agent", browser.Chrome()},
			{"accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"},
			{"sec-fetch-site", "none"},
			{"sec-fetch-mode", "navigate"},
			{"sec-fetch-user", "?1"},
			{"sec-fetch-dest", "document"},
			{"accept-encoding", "gzip, deflate, br"},
			{"accept-language", "en-US,en;q=0.9"},
		},
	})
	if err != nil {
		fmt.Println(err.Error())
	}

	switch resp.StatusCode {
	case 200:
		test := monitor.StockInfo(resp, body, asin)
		fmt.Println(test)
		return test
	case 404:
		monitor.PublishEvent(enums.UnableToFindProduct, enums.MonitorUpdate)
		return false
	case 503:
		fmt.Println("Dogs of Amazon")
		monitor.BecomeGuest()
		return false
	default:
		fmt.Printf("Unkown Code:%v", resp.StatusCode)
		return false
	}

}

// Scraping the info from the page, since we are rotating between two page types I have to check each value in two different places
func (monitor *Monitor) StockInfo(resp *http.Response, body, asin string) bool {
	if strings.Contains(body, "automated access") {
		fmt.Println("Captcha")
		return false
	}
	doc := soup.HTMLParse(body)
	ua := resp.Request.UserAgent()
	var err error
	var ofid string
	var merchantID string
	var priceStr string
	var itemName string

	if strings.Contains(resp.Request.URL.String(), "aod") {
		if doc.Find("input", "name", "offeringID.1").Error != nil {
			return false
		}
		ofid = doc.Find("input", "name", "offeringID.1").Attrs()["value"]
		merchantID = doc.Find("input", "id", "ftSelectMerchant").Attrs()["value"]
		priceStr = doc.Find("span", "class", "a-price-whole").Text()
		itemName = doc.Find("h5", "id", "aod-asin-title-text").Text()
	} else {
		if doc.Find("input", "name", "offerListingID").Error == nil {
			ofid = doc.Find("input", "name", "offerListingID").Attrs()["value"]
		} else {
			ofid, err = util.FindInString(body, `name="offerListingId" value="`, `"`)
			if err != nil {
				return false
			}
		}
		if doc.Find("input", "name", "merchantID").Error != nil {
			return false
		}
		merchantID = doc.Find("input", "name", "merchantID").Attrs()["value"]
		priceStr = doc.Find("span", "class", "a-price-whole").Text()
		if doc.Find("div", "id", "comparison_title1").Error == nil {
			title := doc.Find("div", "id", "comparison_title1").FindAll("span")
			for _, source := range title {
				itemName = source.Text()
			}
		} else {
			title := doc.Find("title").Text()
			itemName, err = util.FindInString(title, "Amazon.com:", ":")
			if err != nil {
				itemName, err = util.FindInString(title, "AmazonSmile:", ":")
				if err != nil {
					return false
				}
			}
		}
	}

	if ofid == "" {
		monitor.RunningMonitors = append(monitor.RunningMonitors, asin)
		return false
	}
	if merchantID != "ATVPDKIKX0DER" {
		monitor.RunningMonitors = append(monitor.RunningMonitors, asin)
		return false
	}

	price, _ := strconv.Atoi(priceStr)
	inBudget := monitor.ASINWithInfo[asin].MaxPrice > price
	if inBudget {
		monitor.EventInfo = events.AmazonSingleStockData{
			ASIN:        asin,
			OfferID:     ofid,
			Price:       price,
			ItemName:    itemName,
			UA:          ua,
			MonitorType: enums.SlowSKUMonitor,
		}
	}

	return inBudget
}

// Takes the task OfferID, ASIN, and SavedAddressID then tries adding that item to the cart,
// this is also known as OfferID mode.
func (monitor *Monitor) OFIDMonitor(asin string) bool {
	currentEndpoint := AmazonEndpoints[util.RandomNumberInt(0, 2)]
	form := url.Values{
		"isAsync":         {"1"},
		"addressID":       {monitor.AddressID},
		"asin.1":          {asin},
		"offerListing.1":  {monitor.ASINWithInfo[asin].OFID},
		"quantity.1":      {"1"},
		"forcePlaceOrder": {"Place+this+duplicate+order"},
	}
	ua := browser.Chrome()
	resp, body, err := util.MakeRequest(&util.Request{
		Client: monitor.AccountClient,
		Method: "POST",
		URL:    currentEndpoint + "/checkout/turbo-initiate?ref_=dp_start-bbf_1_glance_buyNow_2-1&referrer=detail&pipelineType=turbo&clientId=retailwebsite&weblab=RCX_CHECKOUT_TURBO_DESKTOP_PRIME_87783&temporaryAddToCart=1",
		RawHeaders: [][2]string{
			{"content-length", fmt.Sprint(len(form.Encode()))},
			{"sec-ch-ua", `" Not A;Brand";v="99", "Chromium";v="90", "Google Chrome";v="90"`},
			{"x-amz-checkout-entry-referer-url", currentEndpoint + fmt.Sprintf(MonitorEndpoints[util.RandomNumberInt(0, len(MonitorEndpoints))], asin) + util.Randomizer("&pldnSite=1")},
			{"x-amz-turbo-checkout-dp-url", currentEndpoint + fmt.Sprintf(MonitorEndpoints[util.RandomNumberInt(0, len(MonitorEndpoints))], asin) + util.Randomizer("&pldnSite=1")},
			{"rtt", "100"},
			{"sec-ch-ua-mobile", "?0"},
			{"user-agent", ua},
			{"content-type", "application/x-www-form-urlencoded"},
			{"x-amz-support-custom-signin", "1"},
			{"accept", "*/*"},
			{"x-requested-with", "XMLHttpRequest"},
			{"x-amz-checkout-csrf-token", monitor.SessionID},
			{"downlink", "10"},
			{"ect", "4g"},
			{"origin", BaseEndpoint},
			{"sec-fetch-site", "none"},
			{"sec-fetch-mode", "navigate"},
			{"sec-fetch-dest", "document"},
			{"referer", currentEndpoint + fmt.Sprintf(MonitorEndpoints[util.RandomNumberInt(0, len(MonitorEndpoints))], asin) + util.Randomizer("&pldnSite=1")},
			{"accept-encoding", "gzip, deflate, br"},
			{"accept-language", "en-US,en;q=0.9"},
		},
		Data: []byte(form.Encode()),
	})
	if err != nil {
		fmt.Println(err.Error())
	}

	// It is impossible to know if the OfferID actually exists so it's up to the user here when running OfferID mode/Fast mode
	monitor.RunningMonitors = append(monitor.RunningMonitors, asin)
	switch resp.StatusCode {
	case 200:
		var imageURL string
		doc := soup.HTMLParse(body)

		err = doc.Find("input", "name", "anti-csrftoken-a2z").Error
		if err != nil {
			return false
		}
		antiCSRF := doc.Find("input", "name", "anti-csrftoken-a2z").Attrs()["value"]

		pid, err := util.FindInString(body, `currentPurchaseId":"`, `"`)
		if err != nil {
			fmt.Println("Could not find PID")
			return false
		}
		rid, err := util.FindInString(body, `var ue_id = '`, `'`)
		if err != nil {
			fmt.Println("Could not find RID")
			return false
		}
		images := doc.FindAll("img")
		for _, source := range images {
			if strings.Contains(source.Attrs()["src"], "https://m.media-amazon.com") {
				imageURL = source.Attrs()["src"]
			}
		}

		monitor.EventInfo = events.AmazonSingleStockData{
			ASIN:        asin,
			OfferID:     monitor.ASINWithInfo[asin].OFID,
			AntiCsrf:    antiCSRF,
			PID:         pid,
			RID:         rid,
			ImageURL:    imageURL,
			UA:          ua,
			MonitorType: enums.FastSKUMonitor,
		}

		return true
	case 503:
		fmt.Println("Dogs of Amazon (503)")
		return false
	case 403:
		fmt.Println("SessionID expired")
		return false
	default:
		fmt.Printf("Unkown Code:%v", resp.StatusCode)
		return false

	}
}

// This becomes a guest basically to monitor amazon
func (monitor *Monitor) BecomeGuest() bool {
	resp, _, err := util.MakeRequest(&util.Request{
		Client: monitor.Monitor.Client,
		Method: "GET",
		URL:    BaseEndpoint,
		RawHeaders: [][2]string{
			{"upgrade-insecure-requests", "1"},
			{"user-agent", browser.Chrome()},
			{"accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"},
			{"sec-fetch-site", "none"},
			{"sec-fetch-mode", "navigate"},
			{"sec-fetch-user", "?1"},
			{"sec-fetch-dest", "document"},
			{"sec-ch-ua", `" Not A;Brand";v="99", "Chromium";v="90", "Google Chrome";v="90"`},
			{"sec-ch-ua-mobile", "?0"},
			{"accept-encoding", "gzip, deflate, br"},
			{"accept-language", "en-US,en;q=0.9"},
		},
	})
	if err != nil {
		return false
	}

	switch resp.StatusCode {
	case 200:
		return true
	case 503:
		return false
	default:
		return false
	}
}

// SendToTasks sends the product info to tasks
func (monitor *Monitor) SendToTasks() {
	data := events.AmazonStockData{
		InStock: []events.AmazonSingleStockData{monitor.EventInfo},
	}
	monitor.Monitor.EventBus.PublishProductEvent(enums.Amazon, data, monitor.Monitor.TaskGroup.GroupID)
}

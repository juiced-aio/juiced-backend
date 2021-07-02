package amazon

import (
	"fmt"
	"math/rand"
	"net/url"
	"strconv"
	"strings"
	"time"

	"backend.juicedbot.io/juiced.sitescripts/util"

	"backend.juicedbot.io/juiced.client/client"
	"backend.juicedbot.io/juiced.client/http"
	"backend.juicedbot.io/juiced.infrastructure/common"
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
func CreateAmazonMonitor(taskGroup *entities.TaskGroup, proxies []entities.Proxy, eventBus *events.EventBus, singleMonitors []entities.AmazonSingleMonitorInfo) (Monitor, error) {
	storedAmazonMonitors := make(map[string]entities.AmazonSingleMonitorInfo)
	amazonMonitor := Monitor{}
	asins := []string{}

	for _, monitor := range singleMonitors {
		client, err := util.CreateClient(proxies[rand.Intn(len(proxies))])
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
				Proxies:   proxies,
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
		go monitor.RunSingleMonitor(asin)
	}

}

func (monitor *Monitor) RunSingleMonitor(asin string) {
	needToStop := monitor.CheckForStop()
	if needToStop {
		return
	}

	if !common.InSlice(monitor.RunningMonitors, asin) {
		defer func() {
			recover()
			// TODO @silent: Re-run this specific monitor
		}()

		stockData := AmazonInStockData{}
		switch monitor.ASINWithInfo[asin].MonitorType {
		case enums.SlowSKUMonitor:
			stockData = monitor.TurboMonitor(asin)

		case enums.FastSKUMonitor:
			stockData = monitor.OFIDMonitor(asin)
		}

		if stockData.ASIN != "" {
			needToStop := monitor.CheckForStop()
			if needToStop {
				return
			}
			monitor.RunningMonitors = common.RemoveFromSlice(monitor.RunningMonitors, asin)
			var inSlice bool
			for _, monitorStock := range monitor.InStock {
				inSlice = monitorStock.ASIN == stockData.ASIN
			}
			if !inSlice {
				monitor.PublishEvent(enums.SendingProductInfoToTasks, enums.MonitorUpdate)
				monitor.InStock = append(monitor.InStock, stockData)
			}
		} else {
			if len(monitor.RunningMonitors) > 0 {
				if monitor.Monitor.TaskGroup.MonitorStatus != enums.WaitingForInStock {
					monitor.PublishEvent(enums.WaitingForInStock, enums.MonitorUpdate)
				}
			}
			for i, monitorStock := range monitor.InStock {
				if monitorStock.ASIN == stockData.ASIN {
					monitor.InStock = append(monitor.InStock[:i], monitor.InStock[i+1:]...)
					break
				}
			}
			time.Sleep(time.Duration(monitor.Monitor.TaskGroup.MonitorDelay) * time.Millisecond)
			monitor.RunSingleMonitor(asin)
		}
	}
}

// A lot of the stuff that I'm doing either seems useless or dumb but Cloudfront is Ai based and the more entropy/randomness you add to every request
// the better.
func (monitor *Monitor) TurboMonitor(asin string) AmazonInStockData {
	var currentClient http.Client
	stockData := AmazonInStockData{}
	currentEndpoint := AmazonEndpoints[util.RandomNumberInt(0, 2)]
	if currentEndpoint == "https://smile.amazon.com" {
		currentClient = monitor.AccountClient
	} else {
		currentClient = monitor.ASINWithInfo[asin].Client
	}
	client.UpdateProxy(&currentClient, common.ProxyCleaner(monitor.Monitor.Proxies[rand.Intn(len(monitor.Monitor.Proxies))]))
	resp, body, err := util.MakeRequest(&util.Request{
		Client: currentClient,
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
		stockData = monitor.StockInfo(resp, body, asin)
		return stockData
	case 404:
		monitor.PublishEvent(enums.UnableToFindProduct, enums.MonitorUpdate)
		return stockData
	case 503:
		fmt.Println("Dogs of Amazon")
		monitor.BecomeGuest()
		return stockData
	default:
		fmt.Printf("Unkown Code:%v", resp.StatusCode)
		return stockData
	}

}

// Scraping the info from the page, since we are rotating between two page types I have to check each value in two different places
func (monitor *Monitor) StockInfo(resp *http.Response, body, asin string) AmazonInStockData {
	stockData := AmazonInStockData{}
	if strings.Contains(body, "automated access") {
		fmt.Println("Captcha")
		return stockData
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
			return stockData
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
				return stockData
			}
		}
		if doc.Find("input", "name", "merchantID").Error != nil {
			return stockData
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
					return stockData
				}
			}
		}
	}

	if ofid == "" {
		monitor.RunningMonitors = append(monitor.RunningMonitors, asin)
		return stockData
	}
	if merchantID != "ATVPDKIKX0DER" {
		monitor.RunningMonitors = append(monitor.RunningMonitors, asin)
		return stockData
	}

	price, _ := strconv.Atoi(priceStr)
	inBudget := monitor.ASINWithInfo[asin].MaxPrice > price || monitor.ASINWithInfo[asin].MaxPrice == -1
	if inBudget {
		stockData = AmazonInStockData{
			ASIN:        asin,
			OfferID:     ofid,
			Price:       price,
			ItemName:    itemName,
			UA:          ua,
			MonitorType: enums.SlowSKUMonitor,
		}
	}

	return stockData
}

// Takes the task OfferID, ASIN, and SavedAddressID then tries adding that item to the cart,
// this is also known as OfferID mode.
func (monitor *Monitor) OFIDMonitor(asin string) AmazonInStockData {
	stockData := AmazonInStockData{}
	currentEndpoint := AmazonEndpoints[util.RandomNumberInt(0, 2)]
	form := url.Values{
		"isAsync":         {"1"},
		"addressID":       {monitor.AddressID},
		"asin.1":          {asin},
		"offerListing.1":  {monitor.ASINWithInfo[asin].OFID},
		"quantity.1":      {"1"},
		"forcePlaceOrder": {"Place+this+duplicate+order"},
	}
	client.UpdateProxy(&monitor.AccountClient, common.ProxyCleaner(monitor.Monitor.Proxies[rand.Intn(len(monitor.Monitor.Proxies))]))
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
			return stockData
		}
		antiCSRF := doc.Find("input", "name", "anti-csrftoken-a2z").Attrs()["value"]

		pid, err := util.FindInString(body, `currentPurchaseId":"`, `"`)
		if err != nil {
			fmt.Println("Could not find PID")
			return stockData
		}
		rid, err := util.FindInString(body, `var ue_id = '`, `'`)
		if err != nil {
			fmt.Println("Could not find RID")
			return stockData
		}
		images := doc.FindAll("img")
		for _, source := range images {
			if strings.Contains(source.Attrs()["src"], "https://m.media-amazon.com") {
				imageURL = source.Attrs()["src"]
			}
		}

		stockData = AmazonInStockData{
			ASIN:        asin,
			OfferID:     monitor.ASINWithInfo[asin].OFID,
			AntiCsrf:    antiCSRF,
			PID:         pid,
			RID:         rid,
			ImageURL:    imageURL,
			UA:          ua,
			MonitorType: enums.FastSKUMonitor,
		}

		return stockData
	case 503:
		fmt.Println("Dogs of Amazon (503)")
		return stockData
	case 403:
		fmt.Println("SessionID expired")
		return stockData
	default:
		fmt.Printf("Unkown Code:%v", resp.StatusCode)
		return stockData

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

package amazon

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"backend.juicedbot.io/juiced.client/client"
	"backend.juicedbot.io/juiced.client/http"
	"backend.juicedbot.io/juiced.infrastructure/common"
	"backend.juicedbot.io/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/juiced.infrastructure/common/enums"
	"backend.juicedbot.io/juiced.infrastructure/common/events"
	"backend.juicedbot.io/juiced.sitescripts/base"
	"backend.juicedbot.io/juiced.sitescripts/util"
	"github.com/anaskhan96/soup"
	browser "github.com/eddycjy/fake-useragent"
	cmap "github.com/orcaman/concurrent-map"
)

// Creating a pool to store amazon accounts that will be used to monitor
var AccountPool = cmap.New()

// CreateAmazonMonitor takes a TaskGroup entity and turns it into a Amazon Monitor
func CreateAmazonMonitor(taskGroup *entities.TaskGroup, proxyGroup *entities.ProxyGroup, eventBus *events.EventBus, singleMonitors []entities.AmazonSingleMonitorInfo) (Monitor, error) {
	storedAmazonMonitors := make(map[string]entities.AmazonSingleMonitorInfo)
	amazonMonitor := Monitor{}
	asins := []string{}

	for _, monitor := range singleMonitors {
		storedAmazonMonitors[monitor.ASIN] = monitor
		asins = append(asins, monitor.ASIN)
	}

	for created := false; !created; {
		// Making sure the accounts group id matches the monitors

		amazonMonitor = Monitor{
			Monitor: base.Monitor{
				TaskGroup:  taskGroup,
				ProxyGroup: proxyGroup,
				EventBus:   eventBus,
			},

			ASINs:        asins,
			ASINWithInfo: storedAmazonMonitors,
		}
		created = true
	}

	return amazonMonitor, nil
}

// PublishEvent wraps the EventBus's PublishMonitorEvent function
func (monitor *Monitor) PublishEvent(status enums.MonitorStatus, eventType enums.MonitorEventType, data interface{}) {
	monitor.Monitor.TaskGroup.SetMonitorStatus(status)
	monitor.Monitor.EventBus.PublishMonitorEvent(status, eventType, data, monitor.Monitor.TaskGroup.GroupID)
}

// CheckForStop checks the stop flag and stops the monitor if it's true
func (monitor *Monitor) CheckForStop() bool {
	if monitor.Monitor.StopFlag {
		monitor.PublishEvent(enums.MonitorIdle, enums.MonitorStop, nil)
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
		monitor.PublishEvent(enums.WaitingForProductData, enums.MonitorStart, nil)
	}
	needToStop := monitor.CheckForStop()
	if needToStop {
		return
	}

	wg := sync.WaitGroup{}
	wg.Add(len(monitor.ASINs))
	for _, asin := range monitor.ASINs {
		go func(x string) {
			monitor.RunSingleMonitor(x)
			wg.Done()
		}(asin)
	}
	wg.Wait()

}

func (monitor *Monitor) RunSingleMonitor(asin string) {

again:
	needToStop := monitor.CheckForStop()
	if needToStop {
		return
	}
	accounts, _ := AccountPool.Get(monitor.Monitor.TaskGroup.GroupID)
	if accounts != nil {
		if !(len(accounts.([]Acc)) > 0) {
			time.Sleep(common.MS_TO_WAIT)
			goto again
		}
	} else {
		time.Sleep(common.MS_TO_WAIT)
		goto again
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

		if stockData.OfferID != "" {
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
				monitor.PublishEvent(enums.SendingProductInfoToTasks, enums.MonitorUpdate, events.ProductInfo{
					Products: []events.Product{
						{ProductName: stockData.ItemName, ProductImageURL: stockData.ImageURL}},
				})
				monitor.InStock = append(monitor.InStock, stockData)
			}
		} else {
			if len(monitor.RunningMonitors) > 0 {
				if monitor.Monitor.TaskGroup.MonitorStatus != enums.WaitingForInStock {
					monitor.PublishEvent(enums.WaitingForInStock, enums.MonitorUpdate, events.ProductInfo{
						Products: []events.Product{
							{ProductName: stockData.ItemName, ProductImageURL: stockData.ImageURL}},
					})
				}
			}
			for i, monitorStock := range monitor.InStock {
				if monitorStock.ASIN == stockData.ASIN {
					monitor.InStock = append(monitor.InStock[:i], monitor.InStock[i+1:]...)
					break
				}
			}
			monitor.RunningMonitors = common.RemoveFromSlice(monitor.RunningMonitors, asin)
			time.Sleep(time.Duration(monitor.Monitor.TaskGroup.MonitorDelay) * time.Millisecond)
			monitor.RunSingleMonitor(asin)
		}
	}
}

// A lot of the stuff that I'm doing either seems useless or dumb but Cloudfront is Ai based and the more entropy/randomness you add to every request
// the better.
func (monitor *Monitor) TurboMonitor(asin string) AmazonInStockData {
	stockData := AmazonInStockData{}
	currentEndpoint := AmazonEndpoints[util.RandomNumberInt(0, 2)]
	pool, _ := AccountPool.Get(monitor.Monitor.TaskGroup.GroupID)
	account := pool.([]Acc)[rand.Intn(len(pool.([]Acc)))]
	currentClient := account.Client
	ua := browser.Chrome()

	if monitor.Monitor.ProxyGroup != nil {
		var proxy *entities.Proxy
		if len(monitor.Monitor.ProxyGroup.Proxies) > 0 {
			proxy = util.RandomLeastUsedProxy(monitor.Monitor.ProxyGroup.Proxies)
			proxy.AddCount()
			client.UpdateProxy(&currentClient, proxy)
			defer proxy.RemoveCount()
		}

	}

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
			{"user-agent", ua},
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
		stockData = monitor.StockInfo(ua, resp.Request.URL.String(), body, asin)
		return stockData
	case 404:
		monitor.PublishEvent(enums.UnableToFindProduct, enums.MonitorUpdate, nil)
		return stockData
	case 503:
		fmt.Println("Dogs of Amazon")
		return stockData
	default:
		fmt.Printf("Unkown Code:%v", resp.StatusCode)
		return stockData
	}

}

// Scraping the info from the page, since we are rotating between two page types I have to check each value in two different places
func (monitor *Monitor) StockInfo(ua, urL, body, asin string) AmazonInStockData {
	defer func() {
		monitor.RunningMonitors = append(monitor.RunningMonitors, asin)
	}()
	stockData := AmazonInStockData{}
	if strings.Contains(body, "automated access") {
		fmt.Println("Captcha")
		return stockData
	}
	doc := soup.HTMLParse(body)

	var err error
	var ofid string
	var merchantID string
	var priceStr string
	var itemName string
	var imageURL string

	if strings.Contains(urL, "aod") {
		item := doc.Find("span", "data-action", "aod-atc-action")
		if item.Error != nil {
			item := doc.Find("input", "name", "offeringID.1")
			if item.Error == nil {
				ofid = item.Attrs()["value"]
			}
		} else {
			jsonMap := make(map[string]string)
			err = json.Unmarshal([]byte(item.Attrs()["data-aod-atc-action"]), &jsonMap)
			if err == nil {
				ofid = jsonMap["oid"]
			}
		}

		item = doc.Find("input", "id", "ftSelectMerchant")
		if item.Error == nil {
			merchantID = item.Attrs()["value"]
		}

		price := doc.Find("span", "class", "a-price-whole")
		if price.Error == nil {
			priceStr = price.Text()
		}

		item = doc.Find("h5", "id", "aod-asin-title-text")
		if item.Error == nil {
			itemName = item.Text()
		}

		item = doc.Find("img", "id", "aod-asin-image-id")
		if item.Error == nil {
			imageURL = item.Attrs()["src"]
		}

	} else {
		item := doc.Find("input", "name", "offerListingID")
		if item.Error == nil {
			ofid = item.Attrs()["value"]
		} else {
			ofid, err = util.FindInString(body, `name="offerListingId" value="`, `"`)
			if err != nil {
				return stockData
			}
		}

		item = doc.Find("input", "name", "merchantID")
		if item.Error == nil {
			merchantID = item.Attrs()["value"]
		} else {

			item = doc.Find("input", "id", "ftSelectMerchant")
			if item.Error == nil {
				merchantID = item.Attrs()["value"]
			} else {
				return stockData
			}

		}

		item = doc.Find("div", "data-a-image-name", "immersiveViewMainImage")
		if item.Error == nil {
			imageURL = item.Attrs()["data-a-hires"]
		} else {
			item := doc.Find("img", "data-a-image-name", "landingImage")
			if item.Error == nil {
				imageURL = item.Attrs()["data-a-hires"]
			}
		}

		span := doc.Find("span", "id", "tp_price_block_total_price_ww")
		if span.Error == nil {
			price := span.Find("span", "class", "a-price-whole")
			if price.Error == nil {
				priceStr = price.Text()
			}
		}

		item = doc.Find("div", "id", "comparison_title1")
		if item.Error == nil {
			title := item.FindAll("span")
			for _, source := range title {
				itemName = source.Text()
			}
		} else {
			item := doc.Find("title")
			if item.Error == nil {
				title := item.Text()
				itemName, err = util.FindInString(title, "Amazon.com:", ":")
				if err != nil {
					itemName, err = util.FindInString(title, "AmazonSmile:", ":")
					if err != nil {
						return stockData
					}
				}
			}

		}
	}

	price, err := strconv.ParseFloat(priceStr, 64)
	stockData = AmazonInStockData{
		ASIN:        asin,
		OfferID:     ofid,
		Price:       price,
		ItemName:    itemName,
		ImageURL:    imageURL,
		UA:          ua,
		MonitorType: enums.SlowSKUMonitor,
	}

	if price == 0 || err != nil || merchantID != "ATVPDKIKX0DER" || !(float64(monitor.ASINWithInfo[asin].MaxPrice) >= price || monitor.ASINWithInfo[asin].MaxPrice == -1) {
		stockData.OfferID = ""
	}

	return stockData
}

// Takes the task OfferID, ASIN, and SavedAddressID then tries adding that item to the cart,
// this is also known as OfferID mode.
func (monitor *Monitor) OFIDMonitor(asin string) AmazonInStockData {
	stockData := AmazonInStockData{}
	pool, _ := AccountPool.Get(monitor.Monitor.TaskGroup.GroupID)
	account := pool.([]Acc)[rand.Intn(len(pool.([]Acc)))]
	currentEndpoint := AmazonEndpoints[util.RandomNumberInt(0, 2)]
	form := url.Values{
		"isAsync":         {"1"},
		"addressID":       {account.AccountInfo.SavedAddressID},
		"asin.1":          {asin},
		"offerListing.1":  {monitor.ASINWithInfo[asin].OFID},
		"quantity.1":      {"1"},
		"forcePlaceOrder": {"Place+this+duplicate+order"},
	}

	currentClient := account.Client

	if monitor.Monitor.ProxyGroup != nil {
		var proxy *entities.Proxy
		if len(monitor.Monitor.ProxyGroup.Proxies) > 0 {
			proxy = util.RandomLeastUsedProxy(monitor.Monitor.ProxyGroup.Proxies)
			client.UpdateProxy(&currentClient, proxy)
			defer proxy.RemoveCount()
		}

	}

	ua := browser.Chrome()
	resp, body, err := util.MakeRequest(&util.Request{
		Client: currentClient,
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
			{"x-amz-checkout-csrf-token", account.AccountInfo.SessionID},
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
		item := doc.Find("span", "class", "a-color-price")
		var price float64
		if item.Error == nil {
			priceStr := strings.ReplaceAll(item.Text(), " ", "")
			priceStr = strings.ReplaceAll(priceStr, "\n", "")
			priceStr = strings.ReplaceAll(priceStr, "$", "")
			price, _ = strconv.ParseFloat(priceStr, 64)
		}
		inBudget := float64(monitor.ASINWithInfo[asin].MaxPrice) >= price || monitor.ASINWithInfo[asin].MaxPrice == -1
		if inBudget {
			stockData = AmazonInStockData{
				ASIN:        asin,
				OfferID:     monitor.ASINWithInfo[asin].OFID,
				AntiCsrf:    antiCSRF,
				PID:         pid,
				RID:         rid,
				ImageURL:    imageURL,
				Price:       price,
				UA:          ua,
				Client:      currentClient,
				MonitorType: enums.FastSKUMonitor,
			}
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
func (monitor *Monitor) BecomeGuest(client http.Client) bool {
	resp, _, err := util.MakeRequest(&util.Request{
		Client: client,
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

package hottopic

import (
	"io/ioutil"
	"log"
	"net/url"
	"time"

	"backend.juicedbot.io/m/v2/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/m/v2/juiced.infrastructure/common/enums"
	"backend.juicedbot.io/m/v2/juiced.infrastructure/common/events"
	"backend.juicedbot.io/m/v2/juiced.sitescripts/base"
	"backend.juicedbot.io/m/v2/juiced.sitescripts/util"
)

// CreateWalmartTask takes a Task entity and turns it into a Walmart Task
func CreateHottopicTask(task *entities.Task, profile entities.Profile, proxy entities.Proxy, eventBus *events.EventBus) (Task, error) {
	walmartTask := Task{}
	client, err := util.CreateClient(proxy)
	if err != nil {
		return walmartTask, err
	}
	walmartTask = Task{
		Task: base.Task{
			Task:     task,
			Profile:  profile,
			Proxy:    proxy,
			EventBus: eventBus,
			Client:   client,
		},
	}
	return walmartTask, err
}

// PublishEvent wraps the EventBus's PublishTaskEvent function
func (task *Task) PublishEvent(status enums.TaskStatus, eventType enums.TaskEventType) {
	task.Task.Task.SetTaskStatus(status)
	task.Task.EventBus.PublishTaskEvent(status, eventType, nil, task.Task.Task.ID)
}

// CheckForStop checks the stop flag and stops the monitor if it's true
func (task *Task) CheckForStop() bool {
	if task.Task.StopFlag {
		task.PublishEvent(enums.TaskIdle, enums.TaskStop)
		return true
	}
	return false
}

//sSart tasks
func (task *Task) RunTask() {
	task.PublishEvent(enums.WaitingForMonitor, enums.TaskStart)
	// 1. WaitForMonitor
	needToStop := task.WaitForMonitor()
	if needToStop {
		return
	}

	startTime := time.Now()

	//AddTocart
	task.PublishEvent(enums.AddingToCart, enums.TaskUpdate)
	AddToCart := false
	for !AddToCart {
		needToStop := task.CheckForStop()
		if needToStop {
			return
		}
		AddToCart = task.AddToCart()
		if !AddToCart {
			time.Sleep(time.Duration(task.Task.Task.TaskDelay) * time.Millisecond)
		}
	}

	//GetCheckout
	task.PublishEvent(enums.GettingCartInfo, enums.TaskUpdate)
	GetCheckout := false
	for !GetCheckout {
		needToStop := task.CheckForStop()
		if needToStop {
			return
		}
		GetCheckout = task.GetCheckout()
		if !GetCheckout {
			time.Sleep(time.Duration(task.Task.Task.TaskDelay) * time.Millisecond)
		}
	}

	//ProceedToCheckout
	task.PublishEvent(enums.SettingCartInfo, enums.TaskUpdate)
	ProceedToCheckout := false
	for !ProceedToCheckout {
		needToStop := task.CheckForStop()
		if needToStop {
			return
		}
		ProceedToCheckout = task.ProceedToCheckout()
		if !ProceedToCheckout {
			time.Sleep(time.Duration(task.Task.Task.TaskDelay) * time.Millisecond)
		}
	}

	//GuestCheckout
	task.PublishEvent(enums.SettingShippingInfo, enums.TaskUpdate)
	GuestCheckout := false
	for !GuestCheckout {
		needToStop := task.CheckForStop()
		if needToStop {
			return
		}
		GuestCheckout = task.GuestCheckout()
		if !GuestCheckout {
			time.Sleep(time.Duration(task.Task.Task.TaskDelay) * time.Millisecond)
		}
	}

	//SubmitShipping
	task.PublishEvent(enums.SettingBillingInfo, enums.TaskUpdate)
	SubmitShipping := false
	for !SubmitShipping {
		needToStop := task.CheckForStop()
		if needToStop {
			return
		}
		SubmitShipping = task.SubmitShipping()
		if !SubmitShipping {
			time.Sleep(time.Duration(task.Task.Task.TaskDelay) * time.Millisecond)
		}
	}

	//UseOrigAddress
	task.PublishEvent(enums.CheckingOut, enums.TaskUpdate)
	UseOrigAddress := false
	for !UseOrigAddress {
		needToStop := task.CheckForStop()
		if needToStop {
			return
		}
		UseOrigAddress = task.UseOrigAddress()
		if !UseOrigAddress {
			time.Sleep(time.Duration(task.Task.Task.TaskDelay) * time.Millisecond)
		}
	}

	//SubmitPaymentInfo
	task.PublishEvent(enums.CheckingOut, enums.TaskUpdate)
	SubmitPaymentInfo := false
	for !SubmitPaymentInfo {
		needToStop := task.CheckForStop()
		if needToStop {
			return
		}
		SubmitPaymentInfo = task.SubmitPaymentInfo()
		if !SubmitPaymentInfo {
			time.Sleep(time.Duration(task.Task.Task.TaskDelay) * time.Millisecond)
		}
	}

	//SubmitOrder
	task.PublishEvent(enums.CheckingOut, enums.TaskUpdate)
	SubmitOrder := false
	for !SubmitOrder {
		needToStop := task.CheckForStop()
		if needToStop {
			return
		}
		SubmitOrder = task.SubmitOrder()
		if !SubmitOrder {
			time.Sleep(time.Duration(task.Task.Task.TaskDelay) * time.Millisecond)
		}
	}

	endTime := time.Now()

	log.Println("STARTED AT: " + startTime.String())
	log.Println("  ENDED AT: " + endTime.String())
	log.Println("TIME TO CHECK OUT: " + endTime.Sub(startTime).String())

	task.PublishEvent(enums.CheckedOut, enums.TaskComplete)
}

// WaitForMonitor waits until the Monitor has sent the info to the task to continue
func (task *Task) WaitForMonitor() bool {
	for {
		needToStop := task.CheckForStop()
		if needToStop {
			return true
		}
		if task.Pid != "" {
			return false
		}
	}
}

// AddToCart sends a post request to the AddToCartEndpoint with an AddToCartRequest body
func (task *Task) AddToCart() bool {
	//Ive added all the payload values here which are used from the browser however you can remove the marked values
	//and the ATC still seems to function fine. For this reason not sure if we need these values at all.
	//In theory could be a means to identify the bot and cancel orders but that could be an issue for another day.
	data := url.Values{
		"shippingMethod-13249991":  {"shipToHome"},
		"deliveryMsgHome-13249991": {"Backorder:Expected to ship by:05/18/21 - 05/29/21"}, //not needed
		"atc-13249991":             {"0.0"},                                               //not needed
		"storeId-13249991":         {"2536"},                                              //not needed
		"deliveryType-13249991":    {""},                                                  //not needed
		"deliveryMsg-13249991":     {"Unavailable for in-store pickup"},                   //not needed
		"cgid":                     {""},                                                  //not needed
		"pid":                      {task.Pid},
		"Quantity":                 {"1"},
		"hasColorSelected":         {"notRequired"}, //not needed
		"hasSizeSelected":          {"true"},        //not needed
		"hasInseamSelected":        {"notRequired"}, //not needed
		"cartAction":               {"add"},
		"productColor":             {""}, //not needed
	}

	resp, err := util.MakeRequest(&util.Request{
		Client:             task.Task.Client,
		Method:             "POST",
		URL:                AddToCartEndpoint,
		AddHeadersFunction: AddHottopicHeaders,                    //todo
		Referer:            AddToCartReferer + task.Pid + ".html", //todo
		RequestBodyStruct:  data,
	})
	if err != nil { //check the cart isnt empty somehow maybe
		return false
	}

	defer resp.Body.Close()

	return true
}

//GetCheckout is here to get the Dwcont which we need for the next request
//We can also maybe get some cart info here if we want about what items are in the cart
func (task *Task) GetCheckout() bool {
	resp, err := util.MakeRequest(&util.Request{
		Client:             task.Task.Client,
		Method:             "GET",
		URL:                GetCheckoutEndpoint, //setendpoint values
		AddHeadersFunction: AddHottopicHeaders,  //todo
		Referer:            GetCheckoutReferer,  //todo
	})
	if err != nil { //check the cart isnt empty somehow maybe
		return false
	}

	body, _ := ioutil.ReadAll(resp.Body)
	task.Dwcont = getDwCont(string(body))

	defer resp.Body.Close()

	return true
}
func (task *Task) ProceedToCheckout() bool {
	resp, err := util.MakeRequest(&util.Request{
		Client:             task.Task.Client,
		Method:             "GET",
		URL:                ProceedToCheckoutEndpoint + task.Dwcont, //setendpoint values
		AddHeadersFunction: AddHottopicHeaders,                      //todo
		Referer:            ProceedToCheckoutReferer,                //todo
	})
	if err != nil { //check the cart isnt empty somehow maybe
		return false
	}

	body, _ := ioutil.ReadAll(resp.Body)
	bodyText := string(body)
	task.OldDwcont = task.Dwcont
	task.Dwcont = getDwCont(bodyText)
	task.SecureKey = getSecureKey(bodyText)

	defer resp.Body.Close()

	return true
}
func (task *Task) GuestCheckout() bool {
	resp, err := util.MakeRequest(&util.Request{
		Client:             task.Task.Client,
		Method:             "GET",
		URL:                GuestCheckoutEndpoint + task.Dwcont,   //setendpoint values
		AddHeadersFunction: AddHottopicHeaders,                    //todo
		Referer:            GuestCheckoutReferer + task.OldDwcont, //todo
	})
	if err != nil { //check the cart isnt empty somehow maybe
		return false
	}

	body, _ := ioutil.ReadAll(resp.Body)
	bodyText := string(body)
	task.OldDwcont = task.Dwcont
	task.Dwcont = getDwCont(bodyText)
	task.SecureKey = getSecureKey(bodyText)

	defer resp.Body.Close()

	return true
}
func (task *Task) SubmitShipping() bool {
	resp, err := util.MakeRequest(&util.Request{
		Client:             task.Task.Client,
		Method:             "GET",
		URL:                SubmitShippingEndpoint + task.Dwcont,   //setendpoint values
		AddHeadersFunction: AddHottopicHeaders,                     //todo
		Referer:            SubmitShippingReferer + task.OldDwcont, //todo
	})
	if err != nil { //check the cart isnt empty somehow maybe
		return false
	}

	body, _ := ioutil.ReadAll(resp.Body)
	bodyText := string(body)
	task.OldDwcont = task.Dwcont
	task.Dwcont = getDwCont(bodyText)

	defer resp.Body.Close()

	return true
}
func (task *Task) UseOrigAddress() bool {
	resp, err := util.MakeRequest(&util.Request{
		Client:             task.Task.Client,
		Method:             "GET",
		URL:                UseOrigAddressEndpoint + task.Dwcont,   //setendpoint values
		AddHeadersFunction: AddHottopicHeaders,                     //todo
		Referer:            UseOrigAddressReferer + task.OldDwcont, //todo
	})
	if err != nil { //check the cart isnt empty somehow maybe
		return false
	}

	body, _ := ioutil.ReadAll(resp.Body)
	bodyText := string(body)
	task.OldDwcont = task.Dwcont
	task.Dwcont = getDwCont(bodyText)
	task.SecureKey = getSecureKey(bodyText)

	defer resp.Body.Close()

	return true
}
func (task *Task) SubmitPaymentInfo() bool {
	resp, err := util.MakeRequest(&util.Request{
		Client:             task.Task.Client,
		Method:             "GET",
		URL:                SubmitPaymentInfoEndpoint + task.Dwcont,   //setendpoint values
		AddHeadersFunction: AddHottopicHeaders,                        //todo
		Referer:            SubmitPaymentInfoReferer + task.OldDwcont, //todo
	})
	if err != nil { //check the cart isnt empty somehow maybe
		return false
	}

	defer resp.Body.Close()

	return true
}
func (task *Task) SubmitOrder() bool {
	resp, err := util.MakeRequest(&util.Request{
		Client:             task.Task.Client,
		Method:             "GET",
		URL:                SubmitOrderEndpoint,                 //setendpoint values
		AddHeadersFunction: AddHottopicHeaders,                  //todo
		Referer:            SubmitOrderReferer + task.OldDwcont, //todo
	})
	if err != nil { //check the cart isnt empty somehow maybe
		return false
	}

	defer resp.Body.Close()

	return true
}

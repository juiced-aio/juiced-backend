package hottopic

import (
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"backend.juicedbot.io/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/juiced.infrastructure/common/enums"
	"backend.juicedbot.io/juiced.infrastructure/common/events"
	"backend.juicedbot.io/juiced.sitescripts/base"
	"backend.juicedbot.io/juiced.sitescripts/util"
)

// CreateHottopicTask takes a Task entity and turns it into a Hottopic Task
func CreateHottopicTask(task *entities.Task, profile entities.Profile, proxy entities.Proxy, eventBus *events.EventBus) (Task, error) {
	hottopicTask := Task{}

	hottopicTask = Task{
		Task: base.Task{
			Task:     task,
			Profile:  profile,
			Proxy:    proxy,
			EventBus: eventBus,
		},
	}
	return hottopicTask, nil
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

// Start task
func (task *Task) RunTask() {
	defer func() {
		if recover() != nil {
			task.Task.StopFlag = true
			task.PublishEvent(enums.TaskIdle, enums.TaskFail)
		}
		task.PublishEvent(enums.TaskIdle, enums.TaskComplete)
	}()

	if task.Task.Task.TaskDelay == 0 {
		task.Task.Task.TaskDelay = 2000
	}

	client, err := util.CreateClient(task.Task.Proxy)
	if err != nil {
		return
	}
	task.Task.Client = client

	task.PublishEvent(enums.WaitingForMonitor, enums.TaskStart)
	// 1. WaitForMonitor
	needToStop := task.WaitForMonitor()
	if needToStop {
		return
	}

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

	startTime := time.Now()
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
	status := enums.OrderStatusFailed
	for !SubmitOrder {
		needToStop := task.CheckForStop()
		if needToStop {
			return
		}
		if status == enums.OrderStatusDeclined {
			break
		}
		SubmitOrder, status = task.SubmitOrder(startTime)
		if !SubmitOrder {
			time.Sleep(time.Duration(task.Task.Task.TaskDelay) * time.Millisecond)
		}
	}

	endTime := time.Now()

	log.Println("STARTED AT: " + startTime.String())
	log.Println("  ENDED AT: " + endTime.String())
	log.Println("TIME TO CHECK OUT: ", endTime.Sub(startTime).Milliseconds())

	task.PublishEvent(enums.CheckedOut, enums.TaskComplete)
}

// WaitForMonitor waits until the Monitor has sent the info to the task to continue
func (task *Task) WaitForMonitor() bool {
	for {
		needToStop := task.CheckForStop()
		if needToStop {
			return true
		}
		if task.StockData.PID != "" {
			return false
		}
		time.Sleep(1 * time.Millisecond)
	}
}

func (task *Task) AddToCart() bool {
	colorSelected := ""
	sizeSelected := ""
	inseamSelected := ""

	if len(task.StockData.Color) > 0 {
		colorSelected = "true"
	}
	if task.StockData.PID != task.StockData.SizePID {
		sizeSelected = "true"
		inseamSelected = "true"
	}

	data := url.Values{
		"shippingMethod-" + task.StockData.SizePID:  {"shipToHome"},
		"deliveryMsgHome-" + task.StockData.SizePID: {"In Stock"},
		"pid":               {task.StockData.SizePID},
		"Quantity":          {fmt.Sprint(task.Task.Task.TaskQty)},
		"hasColorSelected":  {colorSelected},
		"hasSizeSelected":   {sizeSelected},
		"hasInseamSelected": {inseamSelected},
		"cgid":              {""},
		"cartAction":        {"add"},
		"productColor":      {task.StockData.Color},
	}

	_, body, err := util.MakeRequest(&util.Request{
		Client:             task.Task.Client,
		Method:             "POST",
		URL:                AddToCartEndpoint,
		AddHeadersFunction: AddHottopicHeaders,
		Referer:            AddToCartReferer + task.StockData.PID + ".html",
		Data:               []byte(data.Encode()),
	})

	return err == nil && strings.Contains(body, fmt.Sprintf(`"productId":"%s"`, task.StockData.SizePID)) && strings.Contains(body, fmt.Sprintf(`"quantity":%d`, task.Task.Task.TaskQty))
}

func (task *Task) GetCheckout() bool {
	_, body, err := util.MakeRequest(&util.Request{
		Client:             task.Task.Client,
		Method:             "GET",
		URL:                GetCheckoutEndpoint,
		AddHeadersFunction: AddHottopicHeaders,
		Referer:            GetCheckoutReferer,
	})
	if err != nil {
		return false
	}

	task.Dwcont, err = getDwCont(body)
	return err == nil
}

func (task *Task) ProceedToCheckout() bool {
	data := url.Values{
		"dwfrm_cart_checkoutCart": {"Checkout"},
	}

	_, body, err := util.MakeRequest(&util.Request{
		Client:             task.Task.Client,
		Method:             "POST",
		URL:                ProceedToCheckoutEndpoint + task.Dwcont,
		AddHeadersFunction: AddHottopicHeaders,
		Referer:            ProceedToCheckoutReferer,
		Data:               []byte(data.Encode()),
	})
	if err != nil {
		return false
	}

	task.OldDwcont = task.Dwcont
	task.Dwcont, err = getDwCont(body)
	if err != nil {
		return false
	}
	task.SecureKey, err = getSecureKey(body)

	return err == nil
}

func (task *Task) GuestCheckout() bool {
	data := url.Values{
		"dwfrm_login_unregistered": {"Checkout As a Guest"},
		"dwfrm_login_securekey":    {task.SecureKey},
	}

	_, body, err := util.MakeRequest(&util.Request{
		Client:             task.Task.Client,
		Method:             "POST",
		URL:                GuestCheckoutEndpoint + task.Dwcont,
		AddHeadersFunction: AddHottopicHeaders,
		Referer:            GuestCheckoutReferer + task.OldDwcont,
		Data:               []byte(data.Encode()),
	})
	if err != nil {
		return false
	}

	task.OldDwcont = task.Dwcont
	task.Dwcont, err = getDwCont(body)
	if err != nil {
		return false
	}
	task.SecureKey, err = getSecureKey(body)

	// TODO
	return err == nil
}

func (task *Task) SubmitShipping() bool {
	data := url.Values{
		"dwfrm_singleshipping_shippingAddress_addressFields_phone":        {task.Task.Profile.PhoneNumber},
		"dwfrm_singleshipping_email_emailAddress":                         {task.Task.Profile.Email},
		"dwfrm_singleshipping_addToEmailList":                             {"false"},
		"dwfrm_singleshipping_shippingAddress_addressFields_firstName":    {task.Task.Profile.ShippingAddress.FirstName},
		"dwfrm_singleshipping_shippingAddress_addressFields_lastName":     {task.Task.Profile.ShippingAddress.LastName},
		"dwfrm_singleshipping_shippingAddress_addressFields_country":      {task.Task.Profile.ShippingAddress.CountryCode},
		"dwfrm_singleshipping_shippingAddress_addressFields_postal":       {task.Task.Profile.ShippingAddress.ZipCode},
		"dwfrm_singleshipping_shippingAddress_addressFields_address1":     {task.Task.Profile.ShippingAddress.Address1},
		"dwfrm_singleshipping_shippingAddress_addressFields_address2":     {task.Task.Profile.ShippingAddress.Address2},
		"dwfrm_singleshipping_shippingAddress_addressFields_city":         {task.Task.Profile.ShippingAddress.City},
		"dwfrm_singleshipping_shippingAddress_addressFields_states_state": {task.Task.Profile.ShippingAddress.StateCode},
		"dwfrm_singleshipping_shippingAddress_useAsBillingAddress":        {"false"}, //depends if they want to or not? Not sure what to do here.
		"dwfrm_singleshipping_shippingAddress_shippingMethodID":           {"7D"},    // multiple methods, should we default to 1?
		"dwfrm_singleshipping_shippingAddress_isGift":                     {"false"}, //assume always false?
		"dwfrm_singleshipping_shippingAddress_giftMessage":                {""},      //^
		"dwfrm_singleshipping_shippingAddress_save":                       {"Continue to Billing"},
		"dwfrm_singleshipping_securekey":                                  {task.SecureKey},
	}
	_, body, err := util.MakeRequest(&util.Request{
		Client:             task.Task.Client,
		Method:             "POST",
		URL:                SubmitShippingEndpoint + task.Dwcont,
		AddHeadersFunction: AddHottopicHeaders,
		Referer:            SubmitShippingReferer + task.OldDwcont,
		Data:               []byte(data.Encode()),
	})
	if err != nil {
		return false
	}

	task.OldDwcont = task.Dwcont
	task.Dwcont, err = getDwCont(body)

	// TODO
	return err == nil
}

func (task *Task) UseOrigAddress() bool {
	data := url.Values{
		"dwfrm_addForm_useOrig": {""},
	}
	_, body, err := util.MakeRequest(&util.Request{
		Client:             task.Task.Client,
		Method:             "POST",
		URL:                UseOrigAddressEndpoint + task.Dwcont,
		AddHeadersFunction: AddHottopicHeaders,
		Referer:            UseOrigAddressReferer + task.OldDwcont,
		Data:               []byte(data.Encode()),
	})
	if err != nil { //check the cart isnt empty somehow maybe
		return false
	}

	task.OldDwcont = task.Dwcont
	task.Dwcont, err = getDwCont(body)
	if err != nil {
		return false
	}
	task.SecureKey, err = getSecureKey(body)

	// TODO
	return err == nil
}

func (task *Task) SubmitPaymentInfo() bool {
	data := url.Values{
		"dwfrm_billing_addressChoice_addressChoices":              {"shipping"},
		"dwfrm_billing_billingAddress_addressFields_firstName":    {task.Task.Profile.BillingAddress.FirstName},
		"dwfrm_billing_billingAddress_addressFields_lastName":     {task.Task.Profile.BillingAddress.LastName},
		"dwfrm_billing_billingAddress_addressFields_country":      {task.Task.Profile.BillingAddress.CountryCode},
		"dwfrm_billing_billingAddress_addressFields_postal":       {task.Task.Profile.BillingAddress.ZipCode},
		"dwfrm_billing_billingAddress_addressFields_address1":     {task.Task.Profile.BillingAddress.Address1},
		"dwfrm_billing_billingAddress_addressFields_address2":     {task.Task.Profile.BillingAddress.Address2},
		"dwfrm_billing_billingAddress_addressFields_city":         {task.Task.Profile.BillingAddress.City},
		"dwfrm_billing_billingAddress_addressFields_states_state": {task.Task.Profile.BillingAddress.StateCode},
		"dwfrm_billing_billingAddress_addressFields_phone":        {task.Task.Profile.PhoneNumber},
		"dwfrm_billing_securekey":                                 {task.SecureKey},
		"dwfrm_billing_couponCode":                                {""}, //coupon
		"dwfrm_billing_giftCertCode":                              {""},
		"dwfrm_billing_paymentMethods_selectedPaymentMethodID":    {"CREDIT_CARD"},
		"dwfrm_billing_paymentMethods_creditCard_owner":           {task.Task.Profile.CreditCard.CardholderName},
		"dwfrm_billing_paymentMethods_creditCard_number":          {task.Task.Profile.CreditCard.CardNumber},
		"dwfrm_billing_paymentMethods_creditCard_type":            {task.Task.Profile.CreditCard.CardType},                                                  //Ex VISA
		"dwfrm_billing_paymentMethods_creditCard_month":           {strings.TrimPrefix(task.Task.Profile.CreditCard.ExpMonth, "0")},                         //should be month (no 0) Ex: 2
		"dwfrm_billing_paymentMethods_creditCard_year":            {task.Task.Profile.CreditCard.ExpYear},                                                   //should be full year Ex: 2026
		"dwfrm_billing_paymentMethods_creditCard_userexp":         {task.Task.Profile.CreditCard.ExpMonth + "/" + task.Task.Profile.CreditCard.ExpYear[2:]}, //should be smalldate Ex: 02/26
		"dwfrm_billing_paymentMethods_creditCard_cvn":             {task.Task.Profile.CreditCard.CVV},
		"cardToken":                              {""},                                           //is always empty
		"cardBin":                                {task.Task.Profile.CreditCard.CardNumber[0:6]}, //First 6 digits of card number
		"dwfrm_billing_paymentMethods_bml_year":  {""},                                           //always seems to be empty
		"dwfrm_billing_paymentMethods_bml_month": {""},                                           //always seems to be empty
		"dwfrm_billing_paymentMethods_bml_day":   {""},                                           //always seems to be empty
		"dwfrm_billing_paymentMethods_bml_ssn":   {""},                                           //always seems to be empty
		"dwfrm_billing_save":                     {"Continue to Review"},
	}
	_, _, err := util.MakeRequest(&util.Request{
		Client:             task.Task.Client,
		Method:             "POST",
		URL:                SubmitPaymentInfoEndpoint + task.Dwcont,
		AddHeadersFunction: AddHottopicHeaders,
		Referer:            SubmitPaymentInfoReferer + task.OldDwcont,
		Data:               []byte(data.Encode()),
	})

	// TODO
	return err == nil
}

func (task *Task) SubmitOrder(startTime time.Time) (bool, enums.OrderStatus) {
	status := enums.OrderStatusFailed
	data := url.Values{
		"cardBin":        {task.Task.Profile.CreditCard.CardNumber[0:6]}, //First 6 digits of card number
		"addToEmailList": {"false"},
	}
	_, body, err := util.MakeRequest(&util.Request{
		Client:             task.Task.Client,
		Method:             "POST",
		URL:                SubmitOrderEndpoint,
		AddHeadersFunction: AddHottopicHeaders,
		Referer:            SubmitOrderReferer + task.OldDwcont,
		Data:               []byte(data.Encode()),
	})
	if err != nil {
		return false, status
	}

	var success bool
	if !strings.Contains(body, "Your order could not be submitted") {
		status = enums.OrderStatusSuccess
		success = true
	} else {
		status = enums.OrderStatusDeclined
		success = false
	}

	go util.ProcessCheckout(util.ProcessCheckoutInfo{
		BaseTask:     task.Task,
		Success:      success,
		Content:      "",
		Embeds:       task.CreateHottopicEmbed(status, task.StockData.ImageURL),
		ItemName:     task.StockData.ProductName,
		Sku:          task.StockData.PID,
		Retailer:     enums.HotTopic,
		Price:        float64(task.StockData.Price),
		Quantity:     task.Task.Task.TaskQty,
		MsToCheckout: time.Since(startTime).Milliseconds(),
	})

	return success, status
}

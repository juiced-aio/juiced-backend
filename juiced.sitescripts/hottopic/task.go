package hottopic

import (
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"backend.juicedbot.io/juiced.infrastructure/common"
	"backend.juicedbot.io/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/juiced.infrastructure/common/enums"
	"backend.juicedbot.io/juiced.infrastructure/common/events"
	"backend.juicedbot.io/juiced.sitescripts/base"
	"backend.juicedbot.io/juiced.sitescripts/util"
)

// CreateHottopicTask takes a Task entity and turns it into a Hottopic Task
func CreateHottopicTask(task *entities.Task, profile entities.Profile, proxyGroup *entities.ProxyGroup, eventBus *events.EventBus) (Task, error) {
	hottopicTask := Task{
		Task: base.Task{
			Task:       task,
			Profile:    profile,
			ProxyGroup: proxyGroup,
			EventBus:   eventBus,
		},
	}
	if proxyGroup != nil {
		hottopicTask.Task.Proxy = util.RandomLeastUsedProxy(proxyGroup.Proxies)
	}
	return hottopicTask, nil
}

// PublishEvent wraps the EventBus's PublishTaskEvent function
func (task *Task) PublishEvent(status enums.TaskStatus, eventType enums.TaskEventType, statusPercentage int) {
	task.Task.Task.SetTaskStatus(status)
	task.Task.EventBus.PublishTaskEvent(status, statusPercentage, eventType, nil, task.Task.Task.ID)
}

// CheckForStop checks the stop flag and stops the monitor if it's true
func (task *Task) CheckForStop() bool {
	if task.Task.StopFlag {
		task.PublishEvent(enums.TaskIdle, enums.TaskStop, 0)
		return true
	}
	return false
}

// Start task
func (task *Task) RunTask() {
	defer func() {
		if recover() != nil {
			task.Task.StopFlag = true
			task.PublishEvent(enums.TaskIdle, enums.TaskFail, 0)
		}
		task.PublishEvent(enums.TaskIdle, enums.TaskComplete, 0)
	}()
	task.Task.HasStockData = false

	if task.Task.Task.TaskDelay == 0 {
		task.Task.Task.TaskDelay = 2000
	}
	if task.Task.Task.TaskQty <= 0 {
		task.Task.Task.TaskQty = 1
	}

	err := task.Task.CreateClient(task.Task.Proxy)
	if err != nil {
		return
	}

	task.PublishEvent(enums.WaitingForMonitor, enums.TaskStart, 20)
	// 1. WaitForMonitor
	needToStop := task.WaitForMonitor()
	if needToStop {
		return
	}

	// 2. AddTocart
	task.PublishEvent(enums.AddingToCart, enums.TaskUpdate, 30)
	addedToCart := false
	for !addedToCart {
		needToStop := task.CheckForStop()
		if needToStop {
			return
		}
		addedToCart = task.AddToCart()
		if !addedToCart {
			time.Sleep(time.Duration(task.Task.Task.TaskDelay) * time.Millisecond)
		}
	}

	startTime := time.Now()

	// 3. GetCheckout
	task.PublishEvent(enums.GettingCartInfo, enums.TaskUpdate, 50)
	gotCheckout := false
	for !gotCheckout {
		needToStop := task.CheckForStop()
		if needToStop {
			return
		}
		gotCheckout = task.GetCheckout()
		if !gotCheckout {
			time.Sleep(time.Duration(task.Task.Task.TaskDelay) * time.Millisecond)
		}
	}

	// 4. ProceedToCheckout
	task.PublishEvent(enums.SettingCartInfo, enums.TaskUpdate, 55)
	proceededToCheckout := false
	for !proceededToCheckout {
		needToStop := task.CheckForStop()
		if needToStop {
			return
		}
		proceededToCheckout = task.ProceedToCheckout()
		if !proceededToCheckout {
			time.Sleep(time.Duration(task.Task.Task.TaskDelay) * time.Millisecond)
		}
	}

	// 5. GuestCheckout
	task.PublishEvent(enums.SettingCartInfo, enums.TaskUpdate, 60)
	gotGuestCheckout := false
	for !gotGuestCheckout {
		needToStop := task.CheckForStop()
		if needToStop {
			return
		}
		gotGuestCheckout = task.GuestCheckout()
		if !gotGuestCheckout {
			time.Sleep(time.Duration(task.Task.Task.TaskDelay) * time.Millisecond)
		}
	}

	// 6. SubmitShipping
	task.PublishEvent(enums.SettingShippingInfo, enums.TaskUpdate, 65)
	submittedShipping := false
	for !submittedShipping {
		needToStop := task.CheckForStop()
		if needToStop {
			return
		}
		submittedShipping = task.SubmitShipping()
		if !submittedShipping {
			time.Sleep(time.Duration(task.Task.Task.TaskDelay) * time.Millisecond)
		}
	}

	// 7. UseOrigAddress
	task.PublishEvent(enums.SettingShippingInfo, enums.TaskUpdate, 70)
	usedOrigAddress := false
	for !usedOrigAddress {
		needToStop := task.CheckForStop()
		if needToStop {
			return
		}
		usedOrigAddress = task.UseOrigAddress()
		if !usedOrigAddress {
			time.Sleep(time.Duration(task.Task.Task.TaskDelay) * time.Millisecond)
		}
	}

	// 8. SubmitPaymentInfo
	task.PublishEvent(enums.SettingBillingInfo, enums.TaskUpdate, 80)
	submittedPayment := false
	doNotRetry := false
	for !submittedPayment {
		needToStop := task.CheckForStop()
		if needToStop || doNotRetry {
			return
		}
		submittedPayment, doNotRetry = task.SubmitPaymentInfo()
		if !submittedPayment {
			time.Sleep(time.Duration(task.Task.Task.TaskDelay) * time.Millisecond)
		}
	}

	// 9. SubmitOrder
	task.PublishEvent(enums.CheckingOut, enums.TaskUpdate, 90)
	submittedOrder := false
	status := enums.OrderStatusFailed
	for !submittedOrder {
		needToStop := task.CheckForStop()
		if needToStop {
			return
		}
		if status == enums.OrderStatusDeclined {
			break
		}
		submittedOrder, status = task.SubmitOrder(startTime)
		if !submittedOrder {
			time.Sleep(time.Duration(task.Task.Task.TaskDelay) * time.Millisecond)
		}
	}

	endTime := time.Now()

	log.Println("STARTED AT: " + startTime.String())
	log.Println("  ENDED AT: " + endTime.String())
	log.Println("TIME TO CHECK OUT: ", endTime.Sub(startTime).Milliseconds())

	if status == enums.OrderStatusSuccess {
		task.PublishEvent(enums.CheckedOut, enums.TaskComplete, 100)
	} else {
		task.PublishEvent(enums.CheckoutFailed, enums.TaskComplete, 100)
	}
}

// WaitForMonitor waits until the Monitor has sent the info to the task to continue
func (task *Task) WaitForMonitor() bool {

	for {
		needToStop := task.CheckForStop()
		if needToStop {
			return true
		}
		if task.StockData.PID != "" {
			task.Task.HasStockData = true
			return false
		}
		time.Sleep(common.MS_TO_WAIT)
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

	return err == nil && (strings.Contains(body, "Added to Bag") || (strings.Contains(body, fmt.Sprintf(`"productId":"%s"`, task.StockData.SizePID)) && strings.Contains(body, fmt.Sprintf(`"quantity":%d`, task.Task.Task.TaskQty))))
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
		"dwfrm_singleshipping_shippingAddress_useAsBillingAddress":        {"false"},
		"dwfrm_singleshipping_shippingAddress_shippingMethodID":           {"7D"},
		"dwfrm_singleshipping_shippingAddress_isGift":                     {"false"},
		"dwfrm_singleshipping_shippingAddress_giftMessage":                {""},
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

func (task *Task) SubmitPaymentInfo() (bool, bool) {
	if !common.ValidCardType([]byte(task.Task.Profile.CreditCard.CardNumber), task.Task.Task.TaskRetailer) {
		return false, true
	}

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
		"dwfrm_billing_couponCode":                                {""}, // TODO @Humphrey: Coupon code support
		"dwfrm_billing_giftCertCode":                              {""}, // TODO @Humphrey: Gift certificate support
		"dwfrm_billing_paymentMethods_selectedPaymentMethodID":    {"CREDIT_CARD"},
		"dwfrm_billing_paymentMethods_creditCard_owner":           {task.Task.Profile.CreditCard.CardholderName},
		"dwfrm_billing_paymentMethods_creditCard_number":          {task.Task.Profile.CreditCard.CardNumber},
		"dwfrm_billing_paymentMethods_creditCard_type":            {task.Task.Profile.CreditCard.CardType},
		"dwfrm_billing_paymentMethods_creditCard_month":           {strings.TrimPrefix(task.Task.Profile.CreditCard.ExpMonth, "0")},
		"dwfrm_billing_paymentMethods_creditCard_year":            {task.Task.Profile.CreditCard.ExpYear},
		"dwfrm_billing_paymentMethods_creditCard_userexp":         {task.Task.Profile.CreditCard.ExpMonth + "/" + task.Task.Profile.CreditCard.ExpYear[2:]},
		"dwfrm_billing_paymentMethods_creditCard_cvn":             {task.Task.Profile.CreditCard.CVV},
		"cardToken":                              {""},
		"cardBin":                                {task.Task.Profile.CreditCard.CardNumber[0:6]},
		"dwfrm_billing_paymentMethods_bml_year":  {""},
		"dwfrm_billing_paymentMethods_bml_month": {""},
		"dwfrm_billing_paymentMethods_bml_day":   {""},
		"dwfrm_billing_paymentMethods_bml_ssn":   {""},
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
	return err == nil, false
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

	if success || status == enums.OrderStatusDeclined {
		go util.ProcessCheckout(&util.ProcessCheckoutInfo{
			BaseTask:     task.Task,
			Success:      success,
			Status:       status,
			Content:      "",
			Embeds:       task.CreateHottopicEmbed(status, task.StockData.ImageURL),
			ItemName:     task.StockData.ProductName,
			ImageURL:     task.StockData.ImageURL,
			Sku:          task.StockData.PID,
			Retailer:     enums.HotTopic,
			Price:        float64(task.StockData.Price),
			Quantity:     task.Task.Task.TaskQty,
			MsToCheckout: time.Since(startTime).Milliseconds(),
		})
	}

	return success, status
}

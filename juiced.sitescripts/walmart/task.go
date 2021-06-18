package walmart

import (
	"fmt"
	"log"
	"time"

	"backend.juicedbot.io/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/juiced.infrastructure/common/enums"
	"backend.juicedbot.io/juiced.infrastructure/common/events"
	"backend.juicedbot.io/juiced.infrastructure/queries"
	"backend.juicedbot.io/juiced.sitescripts/base"
	"backend.juicedbot.io/juiced.sitescripts/util"
)

// CreateWalmartTask takes a Task entity and turns it into a Walmart Task
func CreateWalmartTask(task *entities.Task, profile entities.Profile, proxy entities.Proxy, eventBus *events.EventBus) (Task, error) {
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

// RunTask is the script driver that calls all the individual requests
// Function order:
// 		1. WaitForMonitor
//		2. AddToCart
//		3. GetCartInfo
// 		4. SetPCID
//		5. SetShippingInfo
//		6. SetPaymentInfo
//		7. PlaceOrder
func (task *Task) RunTask() {
	// If the function panics due to a runtime error, recover from it
	defer func() {
		recover()
		// TODO @silent: Let the UI know that a task failed
	}()

	task.PublishEvent(enums.WaitingForMonitor, enums.TaskStart)
	// 1. WaitForMonitor
	needToStop := task.WaitForMonitor()
	if needToStop {
		return
	}

	startTime := time.Now()

	// @Tehnic: The endpoint that you are monitoring with automatically adds it to the cart so you should somehow pass the
	// cookies/client to here and then completely cut out the AddToCart request, otherwise using a faster endpoint to monitor would be better.
	// 2. AddToCart
	task.PublishEvent(enums.AddingToCart, enums.TaskUpdate)
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

	// 3. GetCartInfo
	task.PublishEvent(enums.GettingCartInfo, enums.TaskUpdate)
	gotCartInfo := false
	for !gotCartInfo {
		needToStop := task.CheckForStop()
		if needToStop {
			return
		}
		gotCartInfo = task.GetCartInfo()
		if !gotCartInfo {
			time.Sleep(time.Duration(task.Task.Task.TaskDelay) * time.Millisecond)
		}
	}

	// 4. SetPCID
	task.PublishEvent(enums.SettingCartInfo, enums.TaskUpdate)
	setPCID := false
	for !setPCID {
		needToStop := task.CheckForStop()
		if needToStop {
			return
		}
		setPCID = task.SetPCID()
		if !setPCID {
			time.Sleep(time.Duration(task.Task.Task.TaskDelay) * time.Millisecond)
		}
	}

	// 5. SetShippingInfo
	task.PublishEvent(enums.SettingShippingInfo, enums.TaskUpdate)
	setShippingInfo := false
	for !setShippingInfo {
		needToStop := task.CheckForStop()
		if needToStop {
			return
		}
		setShippingInfo = task.SetShippingInfo()
		if !setShippingInfo {
			time.Sleep(time.Duration(task.Task.Task.TaskDelay) * time.Millisecond)
		}
	}

	// 6. SetPaymentInfo
	task.PublishEvent(enums.SettingBillingInfo, enums.TaskUpdate)
	setPaymentInfo := false
	for !setPaymentInfo {
		needToStop := task.CheckForStop()
		if needToStop {
			return
		}
		setPaymentInfo = task.SetPaymentInfo()
		if !setPaymentInfo {
			time.Sleep(time.Duration(task.Task.Task.TaskDelay) * time.Millisecond)
		}
	}

	// 7. PlaceOrder
	task.PublishEvent(enums.CheckingOut, enums.TaskUpdate)
	placedOrder := false
	for !placedOrder {
		needToStop := task.CheckForStop()
		if needToStop {
			return
		}
		placedOrder = task.PlaceOrder(startTime)
		if !placedOrder {
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
		if task.OfferID != "" && task.Sku != "" {
			return false
		}
	}
}

// AddToCart sends a post request to the AddToCartEndpoint with an AddToCartRequest body
func (task *Task) AddToCart() bool {
	data := AddToCartRequest{
		OfferID:               task.OfferID,
		Quantity:              1,
		ShipMethodDefaultRule: "SHIP_RULE_1",
	}

	addToCartResponse := AddToCartResponse{}

	_, _, err := util.MakeRequest(&util.Request{
		Client:             task.Task.Client,
		Method:             "POST",
		URL:                AddToCartEndpoint,
		AddHeadersFunction: AddWalmartHeaders,
		Referer:            AddToCartReferer + "ip/" + task.Sku,
		RequestBodyStruct:  data,
		ResponseBodyStruct: &addToCartResponse,
	})
	if err != nil || addToCartResponse.Cart.ItemCount == 0 {
		return false
	}

	return true
}

// GetCartInfo is required for setting the PCID cookie
func (task *Task) GetCartInfo() bool {
	data := GetCartInfoRequest{
		StoreListIds:  []StoreList{},
		ZipCode:       task.Task.Profile.ShippingAddress.ZipCode,
		City:          task.Task.Profile.ShippingAddress.City,
		State:         task.Task.Profile.ShippingAddress.StateCode,
		IsZipLocated:  true, //setting true as we are populating with values
		Crt:           "",
		CustomerId:    "",
		CustomerType:  "",
		AffiliateInfo: "",
	}

	_, _, err := util.MakeRequest(&util.Request{
		Client:             task.Task.Client,
		Method:             "POST",
		URL:                GetCartInfoEndpoint,
		AddHeadersFunction: AddWalmartHeaders,
		Referer:            GetCartInfoReferer,
		RequestBodyStruct:  data,
	})

	return err == nil
}

// SetPCID sets the PCID cookie
func (task *Task) SetPCID() bool {
	_, _, err := util.MakeRequest(&util.Request{
		Client:             task.Task.Client,
		Method:             "POST",
		URL:                SetPcidEndpoint,
		AddHeadersFunction: AddWalmartHeaders,
		Referer:            SetPcidReferer,
	})

	return err == nil
}

// SetShippingInfo sets the shipping address
func (task *Task) SetShippingInfo() bool {
	data := SetShippingInfoRequest{
		AddressLineOne:     task.Task.Profile.ShippingAddress.Address1,
		City:               task.Task.Profile.ShippingAddress.City,
		FirstName:          task.Task.Profile.ShippingAddress.FirstName,
		LastName:           task.Task.Profile.ShippingAddress.LastName,
		Phone:              task.Task.Profile.PhoneNumber,
		Email:              task.Task.Profile.Email,
		MarketingEmailPref: false,
		PostalCode:         task.Task.Profile.ShippingAddress.ZipCode,
		State:              task.Task.Profile.ShippingAddress.StateCode,
		CountryCode:        task.Task.Profile.ShippingAddress.CountryCode,
		AddressType:        "RESIDENTIAL",
		ChangedFields:      []string{""},
	}

	_, _, err := util.MakeRequest(&util.Request{
		Client:             task.Task.Client,
		Method:             "POST",
		URL:                SetShippingInfoEndpoint,
		AddHeadersFunction: AddWalmartHeaders,
		Referer:            SetShippingInfoReferer,
		RequestBodyStruct:  data,
	})

	return err == nil
}

// SetPaymentInfo sets the payment info to prepare for placing an order
func (task *Task) SetPaymentInfo() bool {
	data := PaymentsRequest{
		[]Payment{{
			PaymentType:    "CreditCard",
			CardType:       task.Task.Profile.CreditCard.CardType,
			FirstName:      task.Task.Profile.BillingAddress.FirstName,
			LastName:       task.Task.Profile.BillingAddress.LastName,
			AddressLineOne: task.Task.Profile.BillingAddress.Address1,
			AddressLineTwo: task.Task.Profile.BillingAddress.Address2,
			City:           task.Task.Profile.BillingAddress.City,
			State:          task.Task.Profile.BillingAddress.StateCode,
			PostalCode:     task.Task.Profile.BillingAddress.ZipCode,
			ExpiryMonth:    task.Task.Profile.CreditCard.ExpMonth,
			ExpiryYear:     task.Task.Profile.CreditCard.ExpYear,
			Email:          task.Task.Profile.Email,
			Phone:          task.Task.Profile.PhoneNumber,
			EncryptedPan:   task.CardInfo.EncryptedPan,
			EncryptedCvv:   task.CardInfo.EncryptedCvv,
			IntegrityCheck: task.CardInfo.IntegrityCheck,
			KeyId:          task.CardInfo.KeyId,
			Phase:          task.CardInfo.Phase,
			PiHash:         task.CardInfo.PiHash,
		}},
		true,
	}

	_, _, err := util.MakeRequest(&util.Request{
		Client:             task.Task.Client,
		Method:             "POST",
		URL:                SetPaymentInfoEndpoint,
		AddHeadersFunction: AddWalmartHeaders,
		Referer:            SetPaymentInfoReferer,
		RequestBodyStruct:  data,
	})

	return err == nil
}

// PlaceOrder completes the checkout process
func (task *Task) PlaceOrder(startTime time.Time) bool {
	data := PlaceOrderRequest{
		CvvInSession: true,
		VoltagePayment: []VoltagePayment{{
			PaymentType:    "CREDITCARD",
			EncryptedCvv:   task.CardInfo.EncryptedCvv,
			EncryptedPan:   task.CardInfo.EncryptedPan,
			IntegrityCheck: task.CardInfo.IntegrityCheck,
			KeyId:          task.CardInfo.KeyId,
			Phase:          task.CardInfo.Phase,
		}},
	}

	placeOrderResponse := PlaceOrderResponse{}

	resp, _, err := util.MakeRequest(&util.Request{
		Client:             task.Task.Client,
		Method:             "POST",
		URL:                PlaceOrderEndpoint,
		AddHeadersFunction: AddWalmartHeaders,
		Referer:            PlaceOrderReferer,
		RequestBodyStruct:  data,
		ResponseBodyStruct: &placeOrderResponse,
	})
	if err != nil {
		fmt.Println(err.Error())
	}

	var status enums.OrderStatus
	var success bool
	switch resp.StatusCode {
	case 200:
		//see if were on success page#
		task.PublishEvent(enums.CheckedOut, enums.TaskComplete)
		status = enums.OrderStatusSuccess
		success = true
	default:
		task.PublishEvent(enums.CheckoutFailed, enums.TaskComplete)
		status = enums.OrderStatusFailed
		success = false
	}

	if success {
		_, user, _ := queries.GetUserInfo()
		util.ProcessCheckout(util.ProcessCheckoutInfo{
			BaseTask:     task.Task,
			Success:      success,
			Content:      "",
			Embeds:       task.CreateWalmartEmbed(status, task.ImageUrl),
			UserInfo:     user,
			ItemName:     task.ItemName,
			Sku:          task.Sku,
			Price:        task.ItemPrice,
			Quantity:     1,
			MsToCheckout: time.Since(startTime).Milliseconds(),
		})
	}
	return success
}

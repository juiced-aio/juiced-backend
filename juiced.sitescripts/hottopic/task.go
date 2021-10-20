package hottopic

import (
	"fmt"
	"log"
	"net/url"
	"strings"

	"backend.juicedbot.io/juiced.infrastructure/entities"
	"backend.juicedbot.io/juiced.infrastructure/enums"
	"backend.juicedbot.io/juiced.sitescripts/util"
)

const MAX_RETRIES = 5

func CreateTask(input entities.TaskInput, baseTask *entities.BaseTask) (entities.RetailerTask, error) {
	// err := ValidateTaskInput(input.SiteSpecificInfo)
	// if err != nil {
	// 	return nil, err
	// }
	return &Task{
		Input:    input,
		BaseTask: baseTask,
	}, nil
}

func (task *Task) GetSetupFunctions() []entities.TaskFunction {
	setupTaskFunctions := []entities.TaskFunction{}
	return setupTaskFunctions
}

func (task *Task) GetMainFunctions() []entities.TaskFunction {
	mainTaskFunctions := []entities.TaskFunction{
		// 1. AddToCart
		{
			Function:         task.AddToCart,
			StatusBegin:      enums.AddingToCart,
			StatusPercentage: 50,
			MsBetweenRetries: task.Input.DelayMS,
		},
		// 2. GetCheckoutInfo
		{
			Function:         task.GetCheckoutInfo,
			StatusBegin:      enums.GettingCartInfo,
			StatusPercentage: 55,
			MaxRetries:       MAX_RETRIES,
		},
		// 3. PrepareCheckout
		{
			Function:         task.PrepareCheckout,
			StatusBegin:      enums.SettingCartInfo,
			StatusPercentage: 60,
			MaxRetries:       MAX_RETRIES,
		},
		// 4. ProceedToGuestCheckout
		{
			Function:         task.ProceedToGuestCheckout,
			StatusBegin:      enums.GettingOrderInfo,
			StatusPercentage: 65,
			MaxRetries:       MAX_RETRIES,
		},
		// 5. SubmitShippingDetails
		{
			Function:         task.SubmitShippingDetails,
			StatusBegin:      enums.SettingShippingInfo,
			StatusPercentage: 70,
			MaxRetries:       MAX_RETRIES,
		},
		// 6. UseOriginalAddress
		{
			Function:         task.UseOriginalAddress,
			StatusBegin:      enums.SettingBillingInfo,
			StatusPercentage: 75,
			MaxRetries:       MAX_RETRIES,
		},
		// 7. SubmitPaymentDetails
		{
			Function:         task.SubmitPaymentDetails,
			StatusBegin:      enums.SettingOrderInfo,
			StatusPercentage: 85,
			MaxRetries:       MAX_RETRIES,
		},
		// 8. SubmitOrder
		{
			Function:         task.SubmitOrder,
			StatusBegin:      enums.CheckingOut,
			MaxRetries:       MAX_RETRIES,
			StatusPercentage: 95,
			CheckoutFunction: true,
		},
	}
	return mainTaskFunctions
}

func (task *Task) AddToCart() (bool, string) {
	colorSelected := ""
	sizeSelected := ""
	inseamSelected := ""

	var color string
	var ok bool
	pid := task.BaseTask.ProductInfo.SKU

	if color, ok = task.BaseTask.ProductInfo.SiteSpecificInfo["Color"].(string); ok && color != "" {
		colorSelected = "true"
	}
	if vid, ok := task.BaseTask.ProductInfo.SiteSpecificInfo["VID"].(string); ok && vid != task.BaseTask.ProductInfo.SKU {
		sizeSelected = "true"
		inseamSelected = "true"
		pid = vid
	}

	data := url.Values{
		"shippingMethod-" + pid:  {"shipToHome"},
		"deliveryMsgHome-" + pid: {"In Stock"},
		"pid":                    {pid},
		"Quantity":               {fmt.Sprint(task.Input.Quantity)},
		"hasColorSelected":       {colorSelected},
		"hasSizeSelected":        {sizeSelected},
		"hasInseamSelected":      {inseamSelected},
		"cgid":                   {""},
		"cartAction":             {"add"},
		"productColor":           {color},
	}

	resp, body, err := util.MakeRequest(&util.Request{
		Client:             task.BaseTask.Client,
		Method:             "POST",
		URL:                AddToCartEndpoint,
		AddHeadersFunction: AddHottopicHeaders,
		Referer:            AddToCartReferer + pid + ".html",
		Data:               []byte(data.Encode()),
	})
	if err != nil {
		log.Println(err.Error())
		if resp == nil {
			return false, fmt.Sprintf(enums.AddingToCartFailure, err.Error())
		}
		// return false, fmt.Sprintf(enums.AddingToCartFailure, err.Error())
	}

	switch resp.StatusCode {
	case 200:
		if strings.Contains(body, "Added to Bag") || (strings.Contains(body, fmt.Sprintf(`"productId":"%s"`, pid)) && strings.Contains(body, fmt.Sprintf(`"quantity":%d`, task.Input.Quantity))) {
			return true, enums.AddingToCartSuccess
		}
	}

	return false, fmt.Sprintf(enums.AddingToCartFailure, fmt.Sprintf(enums.UnknownError, resp.StatusCode))
}

func (task *Task) GetCheckoutInfo() (bool, string) {
	resp, body, err := util.MakeRequest(&util.Request{
		Client:             task.BaseTask.Client,
		Method:             "GET",
		URL:                GetCheckoutEndpoint,
		AddHeadersFunction: AddHottopicHeaders,
		Referer:            GetCheckoutReferer,
	})
	if err != nil {
		log.Println(err.Error())
		if resp == nil {
			return false, fmt.Sprintf(enums.GettingCartInfoFailure, err.Error())
		}
		// return false, fmt.Sprintf(enums.GettingCartInfoFailure, err.Error())
	}

	switch resp.StatusCode {
	case 200:
		task.Dwcont, err = getDwCont(body)
		if err != nil {
			return false, fmt.Sprintf(enums.GettingCartInfoFailure, err.Error())
		}
		return true, enums.GettingCartInfoSuccess
	}

	return false, fmt.Sprintf(enums.GettingCartInfoFailure, fmt.Sprintf(enums.UnknownError, resp.StatusCode))
}

func (task *Task) PrepareCheckout() (bool, string) {
	data := url.Values{
		"dwfrm_cart_checkoutCart": {"Checkout"},
	}

	resp, body, err := util.MakeRequest(&util.Request{
		Client:             task.BaseTask.Client,
		Method:             "POST",
		URL:                ProceedToCheckoutEndpoint + task.Dwcont,
		AddHeadersFunction: AddHottopicHeaders,
		Referer:            ProceedToCheckoutReferer,
		Data:               []byte(data.Encode()),
	})
	if err != nil {
		log.Println(err.Error())
		if resp == nil {
			return false, fmt.Sprintf(enums.SettingCartInfoFailure, err.Error())
		}
		// return false, fmt.Sprintf(enums.SettingCartInfoFailure, err.Error())
	}

	switch resp.StatusCode {
	case 200:
		prevDwCont := task.Dwcont
		newDwCont, err := getDwCont(body)
		if err != nil {
			return false, fmt.Sprintf(enums.SettingCartInfoFailure, err.Error())
		}

		secureKey, err := getSecureKey(body)
		if err != nil {
			return false, fmt.Sprintf(enums.SettingCartInfoFailure, err.Error())
		}

		task.PrevDwcont = prevDwCont
		task.Dwcont = newDwCont
		task.SecureKey = secureKey
		return true, enums.GettingCartInfoSuccess
	}

	return false, fmt.Sprintf(enums.GettingCartInfoFailure, fmt.Sprintf(enums.UnknownError, resp.StatusCode))
}

func (task *Task) ProceedToGuestCheckout() (bool, string) {
	data := url.Values{
		"dwfrm_login_unregistered": {"Checkout As a Guest"},
		"dwfrm_login_securekey":    {task.SecureKey},
	}

	resp, body, err := util.MakeRequest(&util.Request{
		Client:             task.BaseTask.Client,
		Method:             "POST",
		URL:                GuestCheckoutEndpoint + task.Dwcont,
		AddHeadersFunction: AddHottopicHeaders,
		Referer:            GuestCheckoutReferer + task.PrevDwcont,
		Data:               []byte(data.Encode()),
	})
	if err != nil {
		log.Println(err.Error())
		if resp == nil {
			return false, fmt.Sprintf(enums.GettingOrderInfoFailure, err.Error())
		}
		// return false, fmt.Sprintf(enums.GettingOrderInfoFailure, err.Error())
	}

	switch resp.StatusCode {
	case 200:
		prevDwCont := task.Dwcont
		newDwCont, err := getDwCont(body)
		if err != nil {
			return false, fmt.Sprintf(enums.GettingOrderInfoFailure, err.Error())
		}

		secureKey, err := getSecureKey(body)
		if err != nil {
			return false, fmt.Sprintf(enums.GettingOrderInfoFailure, err.Error())
		}

		task.PrevDwcont = prevDwCont
		task.Dwcont = newDwCont
		task.SecureKey = secureKey
		return true, enums.GettingOrderInfoSuccess
	}

	return false, fmt.Sprintf(enums.GettingOrderInfoFailure, fmt.Sprintf(enums.UnknownError, resp.StatusCode))
}

func (task *Task) SubmitShippingDetails() (bool, string) {
	data := url.Values{
		"dwfrm_singleshipping_shippingAddress_addressFields_phone":        {task.BaseTask.Profile.PhoneNumber},
		"dwfrm_singleshipping_email_emailAddress":                         {task.BaseTask.Profile.Email},
		"dwfrm_singleshipping_addToEmailList":                             {"false"},
		"dwfrm_singleshipping_shippingAddress_addressFields_firstName":    {task.BaseTask.Profile.ShippingAddress.FirstName},
		"dwfrm_singleshipping_shippingAddress_addressFields_lastName":     {task.BaseTask.Profile.ShippingAddress.LastName},
		"dwfrm_singleshipping_shippingAddress_addressFields_country":      {task.BaseTask.Profile.ShippingAddress.CountryCode},
		"dwfrm_singleshipping_shippingAddress_addressFields_postal":       {task.BaseTask.Profile.ShippingAddress.ZipCode},
		"dwfrm_singleshipping_shippingAddress_addressFields_address1":     {task.BaseTask.Profile.ShippingAddress.Address1},
		"dwfrm_singleshipping_shippingAddress_addressFields_address2":     {task.BaseTask.Profile.ShippingAddress.Address2},
		"dwfrm_singleshipping_shippingAddress_addressFields_city":         {task.BaseTask.Profile.ShippingAddress.City},
		"dwfrm_singleshipping_shippingAddress_addressFields_states_state": {task.BaseTask.Profile.ShippingAddress.StateCode},
		"dwfrm_singleshipping_shippingAddress_useAsBillingAddress":        {"false"},
		"dwfrm_singleshipping_shippingAddress_shippingMethodID":           {"7D"},
		"dwfrm_singleshipping_shippingAddress_isGift":                     {"false"},
		"dwfrm_singleshipping_shippingAddress_giftMessage":                {""},
		"dwfrm_singleshipping_shippingAddress_save":                       {"Continue to Billing"},
		"dwfrm_singleshipping_securekey":                                  {task.SecureKey},
	}

	resp, body, err := util.MakeRequest(&util.Request{
		Client:             task.BaseTask.Client,
		Method:             "POST",
		URL:                SubmitShippingEndpoint + task.Dwcont,
		AddHeadersFunction: AddHottopicHeaders,
		Referer:            SubmitShippingReferer + task.PrevDwcont,
		Data:               []byte(data.Encode()),
	})
	if err != nil {
		log.Println(err.Error())
		if resp == nil {
			return false, fmt.Sprintf(enums.SettingShippingInfoFailure, err.Error())
		}
		// return false, fmt.Sprintf(enums.SettingShippingInfoFailure, err.Error())
	}

	switch resp.StatusCode {
	case 200:
		prevDwCont := task.Dwcont
		newDwCont, err := getDwCont(body)
		if err != nil {
			return false, fmt.Sprintf(enums.SettingShippingInfoFailure, err.Error())
		}

		task.PrevDwcont = prevDwCont
		task.Dwcont = newDwCont
		return true, enums.SettingShippingInfoSuccess
	}

	return false, fmt.Sprintf(enums.SettingShippingInfoFailure, fmt.Sprintf(enums.UnknownError, resp.StatusCode))
}

func (task *Task) UseOriginalAddress() (bool, string) {
	data := url.Values{
		"dwfrm_addForm_useOrig": {""},
	}
	resp, body, err := util.MakeRequest(&util.Request{
		Client:             task.BaseTask.Client,
		Method:             "POST",
		URL:                UseOrigAddressEndpoint + task.Dwcont,
		AddHeadersFunction: AddHottopicHeaders,
		Referer:            UseOrigAddressReferer + task.PrevDwcont,
		Data:               []byte(data.Encode()),
	})
	if err != nil {
		log.Println(err.Error())
		if resp == nil {
			return false, fmt.Sprintf(enums.SettingBillingInfoFailure, err.Error())
		}
		// return false, fmt.Sprintf(enums.SettingBillingInfoFailure, err.Error())
	}

	switch resp.StatusCode {
	case 200:
		prevDwCont := task.Dwcont
		newDwCont, err := getDwCont(body)
		if err != nil {
			return false, fmt.Sprintf(enums.SettingBillingInfoFailure, err.Error())
		}

		secureKey, err := getSecureKey(body)
		if err != nil {
			return false, fmt.Sprintf(enums.SettingBillingInfoFailure, err.Error())
		}

		task.PrevDwcont = prevDwCont
		task.Dwcont = newDwCont
		task.SecureKey = secureKey
		return true, enums.SettingBillingInfoSuccess
	}

	return false, fmt.Sprintf(enums.SettingBillingInfoFailure, fmt.Sprintf(enums.UnknownError, resp.StatusCode))
}

func (task *Task) SubmitPaymentDetails() (bool, string) {
	data := url.Values{
		"dwfrm_billing_addressChoice_addressChoices":              {"shipping"},
		"dwfrm_billing_billingAddress_addressFields_firstName":    {task.BaseTask.Profile.BillingAddress.FirstName},
		"dwfrm_billing_billingAddress_addressFields_lastName":     {task.BaseTask.Profile.BillingAddress.LastName},
		"dwfrm_billing_billingAddress_addressFields_country":      {task.BaseTask.Profile.BillingAddress.CountryCode},
		"dwfrm_billing_billingAddress_addressFields_postal":       {task.BaseTask.Profile.BillingAddress.ZipCode},
		"dwfrm_billing_billingAddress_addressFields_address1":     {task.BaseTask.Profile.BillingAddress.Address1},
		"dwfrm_billing_billingAddress_addressFields_address2":     {task.BaseTask.Profile.BillingAddress.Address2},
		"dwfrm_billing_billingAddress_addressFields_city":         {task.BaseTask.Profile.BillingAddress.City},
		"dwfrm_billing_billingAddress_addressFields_states_state": {task.BaseTask.Profile.BillingAddress.StateCode},
		"dwfrm_billing_billingAddress_addressFields_phone":        {task.BaseTask.Profile.PhoneNumber},
		"dwfrm_billing_securekey":                                 {task.SecureKey},
		"dwfrm_billing_couponCode":                                {""}, // TODO: Coupon code support
		"dwfrm_billing_giftCertCode":                              {""}, // TODO: Gift certificate support
		"dwfrm_billing_paymentMethods_selectedPaymentMethodID":    {"CREDIT_CARD"},
		"dwfrm_billing_paymentMethods_creditCard_owner":           {task.BaseTask.Profile.CreditCard.CardholderName},
		"dwfrm_billing_paymentMethods_creditCard_number":          {task.BaseTask.Profile.CreditCard.CardNumber},
		"dwfrm_billing_paymentMethods_creditCard_type":            {task.BaseTask.Profile.CreditCard.CardType},
		"dwfrm_billing_paymentMethods_creditCard_month":           {strings.TrimPrefix(task.BaseTask.Profile.CreditCard.ExpMonth, "0")},
		"dwfrm_billing_paymentMethods_creditCard_year":            {task.BaseTask.Profile.CreditCard.ExpYear},
		"dwfrm_billing_paymentMethods_creditCard_userexp":         {task.BaseTask.Profile.CreditCard.ExpMonth + "/" + task.BaseTask.Profile.CreditCard.ExpYear[2:]},
		"dwfrm_billing_paymentMethods_creditCard_cvn":             {task.BaseTask.Profile.CreditCard.CVV},
		"cardToken":                              {""},
		"cardBin":                                {task.BaseTask.Profile.CreditCard.CardNumber[0:6]},
		"dwfrm_billing_paymentMethods_bml_year":  {""},
		"dwfrm_billing_paymentMethods_bml_month": {""},
		"dwfrm_billing_paymentMethods_bml_day":   {""},
		"dwfrm_billing_paymentMethods_bml_ssn":   {""},
		"dwfrm_billing_save":                     {"Continue to Review"},
	}
	resp, _, err := util.MakeRequest(&util.Request{
		Client:             task.BaseTask.Client,
		Method:             "POST",
		URL:                SubmitPaymentInfoEndpoint + task.Dwcont,
		AddHeadersFunction: AddHottopicHeaders,
		Referer:            SubmitPaymentInfoReferer + task.PrevDwcont,
		Data:               []byte(data.Encode()),
	})
	if err != nil {
		log.Println(err.Error())
		if resp == nil {
			return false, fmt.Sprintf(enums.SettingOrderInfoFailure, err.Error())
		}
		// return false, fmt.Sprintf(enums.SettingOrderInfoFailure, err.Error())
	}

	switch resp.StatusCode {
	case 200:
		return true, enums.SettingOrderInfoSuccess
	}

	return false, fmt.Sprintf(enums.SettingOrderInfoFailure, fmt.Sprintf(enums.UnknownError, resp.StatusCode))
}

func (task *Task) SubmitOrder() (bool, string) {
	data := url.Values{
		"cardBin":        {task.BaseTask.Profile.CreditCard.CardNumber[0:6]},
		"addToEmailList": {"false"},
	}
	resp, body, err := util.MakeRequest(&util.Request{
		Client:             task.BaseTask.Client,
		Method:             "POST",
		URL:                SubmitOrderEndpoint,
		AddHeadersFunction: AddHottopicHeaders,
		Referer:            SubmitOrderReferer + task.PrevDwcont,
		Data:               []byte(data.Encode()),
	})
	if err != nil {
		log.Println(err.Error())
		if resp == nil {
			return false, fmt.Sprintf(enums.CheckingOutFailure, err.Error())
		}
		// return false, fmt.Sprintf(enums.CheckingOutFailure, err.Error())
	}

	switch resp.StatusCode {
	case 200:
		if !strings.Contains(body, "Your order could not be submitted") {
			return true, enums.CheckingOutSuccess
		} else {
			return false, enums.CardDeclined
		}
	}

	return false, fmt.Sprintf(enums.CheckingOutFailure, fmt.Sprintf(enums.UnknownError, resp.StatusCode))
}

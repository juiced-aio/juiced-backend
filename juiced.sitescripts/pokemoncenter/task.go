package pokemoncenter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"backend.juicedbot.io/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/juiced.infrastructure/common/enums"
	"backend.juicedbot.io/juiced.infrastructure/common/events"
	"backend.juicedbot.io/juiced.sitescripts/base"
	"backend.juicedbot.io/juiced.sitescripts/util"
)

// CreatePokemonCenterTask takes a Task entity and turns it into a PokemonCenter Task
func CreatePokemonCenterTask(task *entities.Task, profile entities.Profile, proxy entities.Proxy, eventBus *events.EventBus) (Task, error) {
	pokemonCenterTask := Task{}

	pokemonCenterTask = Task{
		Task: base.Task{
			Task:     task,
			Profile:  profile,
			Proxy:    proxy,
			EventBus: eventBus,
		},
	}
	return pokemonCenterTask, nil
}

// RefreshPX3 refreshes the px3 cookie every 4 minutes since it expires every 5 minutes
func (task *Task) RefreshPX3() {
	defer func() {
		recover()
		task.RefreshPX3()
	}()
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
		task.Task.StopFlag = true
		task.PublishEvent(enums.TaskIdle, enums.TaskFail)
	}()

	task.PublishEvent(enums.WaitingForMonitor, enums.TaskStart)
	go task.RefreshPX3()

	// 1. WaitForMonitor
	needToStop := task.WaitForMonitor()
	if needToStop {
		return
	}

	startTime := time.Now()

	// 2. AddToCart
	task.PublishEvent(enums.AddingToCart, enums.TaskUpdate)
	addedToCart := false
	for !addedToCart {
		needToStop := task.CheckForStop()
		if needToStop {
			return
		}
		/* = task.AddToCart()
		if !addedToCart {
			time.Sleep(time.Duration(task.Task.Task.TaskDelay) * time.Millisecond)
		}*/
	}

	// 3. GetCartInfo
	task.PublishEvent(enums.GettingCartInfo, enums.TaskUpdate)
	gotCartInfo := false
	for !gotCartInfo {
		needToStop := task.CheckForStop()
		if needToStop {
			return
		}
		/*gotCartInfo = task.GetCartInfo()
		if !gotCartInfo {
			time.Sleep(time.Duration(task.Task.Task.TaskDelay) * time.Millisecond)
		}*/
	}

	// 4. SetPCID
	task.PublishEvent(enums.SettingCartInfo, enums.TaskUpdate)
	setPCID := false
	for !setPCID {
		needToStop := task.CheckForStop()
		if needToStop {
			return
		}
		/*setPCID = task.SetPCID()
		if !setPCID {
			time.Sleep(time.Duration(task.Task.Task.TaskDelay) * time.Millisecond)
		}*/
	}

	// 5. SetShippingInfo
	task.PublishEvent(enums.SettingShippingInfo, enums.TaskUpdate)
	setShippingInfo := false
	for !setShippingInfo {
		needToStop := task.CheckForStop()
		if needToStop {
			return
		}
		/*setShippingInfo = task.GetCartInfo()
		if !setShippingInfo {
			time.Sleep(time.Duration(task.Task.Task.TaskDelay) * time.Millisecond)
		}*/
	}

	// 6. SetPaymentInfo
	task.PublishEvent(enums.SettingBillingInfo, enums.TaskUpdate)
	setPaymentInfo := false
	for !setPaymentInfo {
		needToStop := task.CheckForStop()
		if needToStop {
			return
		}
		/*setPaymentInfo = task.GetCartInfo()
		if !setPaymentInfo {
			time.Sleep(time.Duration(task.Task.Task.TaskDelay) * time.Millisecond)
		}*/
	}

	// 7. PlaceOrder
	task.PublishEvent(enums.CheckingOut, enums.TaskUpdate)
	placedOrder := false
	for !placedOrder {
		needToStop := task.CheckForStop()
		if needToStop {
			return
		}
		/*placedOrder = task.GetCartInfo()
		if !placedOrder {
			time.Sleep(time.Duration(task.Task.Task.TaskDelay) * time.Millisecond)
		}*/
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
	}
}

func (task *Task) Login() bool {
	return true
}

func (task *Task) AddToCart() bool {
	//Setup request using data passed from 'Instock' data to the tasks 'Checkout data' (Done in monitor-store)
	addToCartRequest := AddToCartRequest{ProductUri: task.CheckoutInfo.AddToCartForm, Quantity: 1, Configuration: ""}
	//Empty Response for the response
	addToCartResponse := AddToCartResponse{}

	//json marshal this for content length.
	addToCartRequestBytes, err := json.Marshal(addToCartRequest)
	if err != nil {
		log.Fatal("Marshal payload failed with error " + err.Error())
	}

	//setup request
	resp, _, err := util.MakeRequest(&util.Request{
		Client: task.Task.Client,
		Method: "POST",
		URL:    fmt.Sprintf(AddToCartEndpoint),
		RawHeaders: [][2]string{
			{"content-length", fmt.Sprint(bytes.NewReader(addToCartRequestBytes).Size())},
			{"sec-ch-ua", "\" Not A;Brand\";v=\"99\", \"Chromium\";v=\"90\", \"Google Chrome\";v=\"90\""},
			{"accept", "*/*"},
			{"x-requested-with", "XMLHttpRequest"},
			{"sec-ch-ua-mobile", "?0"},
			{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.212 Safari/537.36"},
			{"content-type", "application/x-www-form-urlencoded; charset=UTF-8"},
			{"origin", BaseEndpoint},
			{"sec-fetch-site", "same-origin"},
			{"sec-fetch-mode", "cors"},
			{"sec-fetch-dest", "empty"},
			{"referer", fmt.Sprintf(AddToCartRefererEndpoint, task.CheckoutInfo.SKU)},
			{"accept-encoding", "gzip, deflate, br"},
			{"accept-language", "en-US,en;q=0.9"},
		},
		ResponseBodyStruct: &addToCartResponse,
		RequestBodyStruct:  &addToCartRequest,
	})
	if err != nil {
		fmt.Println(err.Error())
	}

	switch resp.StatusCode {
	case 200:
		switch addToCartResponse.Type {
		case "carts.line-item":
			//instock
			return true
		default:
			//anything else is out of stock but we can get specific if need be, For example captcha
			return false
		}
	default:
		return false
	}
}

func (task *Task) SubmitEmailAddress() bool {
	return true
}

func (task *Task) SubmitAddressDetailsValidate() bool {
	return true
}

func (task *Task) SubmitAddressDetails() bool {
	return true
}

func (task *Task) GetPaymentKeyId() bool {
	return true
}

func (task *Task) SubmitPaymentDetails() bool {
	return true
}

func (task *Task) Checkout() bool {
	return true
}

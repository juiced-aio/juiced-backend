package hottopic

import (
	"backend.juicedbot.io/m/v2/juiced.infrastructure/common/enums"
	"backend.juicedbot.io/m/v2/juiced.sitescripts/base"
)

// Endpoints
const (
	BaseEndpoint      = "https://www.hottopic.com/"
	AddToCartEndpoint = "https://www.hottopic.com/on/demandware.store/Sites-hottopic-Site/default/Cart-AddProduct?format=ajax"
	AddToCartReferer  = ""

	GetCheckoutEndpoint = "https://www.hottopic.com/cart"
	SuffixEndpoint      = "https://www.hottopic.com/cart?"
	SubmitOrderEndpoint = "https://www.hottopic.com/orderconfirmation"

	MonitorEndpoint = ""
)

// Monitor info
type Monitor struct {
	Monitor     base.Monitor
	MonitorType enums.MonitorType
	Pids        []string
}

// Task info
type Task struct {
	Task      base.Task
	Pid       string
	Dwcont    string
	SecureKey string
}

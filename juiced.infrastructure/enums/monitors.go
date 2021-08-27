package enums

type MonitorStatus = string

const (
	MonitorIdle               MonitorStatus = "Idle"
	SettingUpMonitor          MonitorStatus = "Setting up"
	BypassingPXMonitor        MonitorStatus = "Bypassing PX"
	WaitingForProductData     MonitorStatus = "Searching"
	ProxyBanned               MonitorStatus = "Proxy is banned"
	UnableToFindProduct       MonitorStatus = "Product not found"
	WaitingForInStock         MonitorStatus = "Out of stock"
	OutOfPriceRange           MonitorStatus = "Out of price range"
	SendingProductInfoToTasks MonitorStatus = "Sending to tasks"
	SentProductInfoToTasks    MonitorStatus = "Tasks in progress"
)

type MonitorEventType = string

const (
	MonitorStart    MonitorEventType = "MonitorStart"
	MonitorUpdate   MonitorEventType = "MonitorUpdate"
	MonitorFail     MonitorEventType = "MonitorFail"
	MonitorStop     MonitorEventType = "MonitorStop"
	MonitorComplete MonitorEventType = "MonitorComplete"
)

type MonitorType = string

const (
	SKUMonitor     MonitorType = "SKU_MONITOR"
	FastSKUMonitor MonitorType = "FAST_SKU_MONITOR"
	SlowSKUMonitor MonitorType = "SLOW_SKU_MONITOR"
	URLMonitor     MonitorType = "URL_MONITOR"
	KeywordMonitor MonitorType = "KEYWORD_MONITOR"
)

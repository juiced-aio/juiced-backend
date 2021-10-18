package enums

import "fmt"

type MonitorStatus = string

const (
	MonitorIdle   MonitorStatus = "Idle"
	MonitorFailed MonitorStatus = "Fatal error: %s"

	SettingUpMonitor MonitorStatus = "Setting up"
	Searching        MonitorStatus = "Searching"

	WaitingForCaptchaMonitor  MonitorStatus = "Waiting for Captcha"
	BypassingPXMonitor        MonitorStatus = "Bypassing PX"
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

var validMonitorTypes = []MonitorType{
	SKUMonitor,
	FastSKUMonitor,
	SlowSKUMonitor,
	URLMonitor,
	KeywordMonitor,
}

var supportedMonitorTypes = map[Retailer][]MonitorType{
	Amazon:        {SlowSKUMonitor, FastSKUMonitor},
	BestBuy:       {SKUMonitor},
	BoxLunch:      {SKUMonitor},
	Disney:        {SKUMonitor},
	GameStop:      {SKUMonitor},
	HotTopic:      {SKUMonitor},
	Newegg:        {SKUMonitor},
	PokemonCenter: {SKUMonitor},
	Target:        {SKUMonitor},
	Topps:         {URLMonitor},
	Walmart:       {SKUMonitor, FastSKUMonitor},

	GenericShopify:  {SKUMonitor, URLMonitor},
	MattelCreations: {SKUMonitor, URLMonitor},
}

type InvalidMonitorTypeError struct {
	MonitorType string
}

func (e *InvalidMonitorTypeError) Error() string {
	return fmt.Sprintf("invalid monitor type: %s", e.MonitorType)
}

type UnsupportedMonitorTypeError struct {
	Retailer    string
	MonitorType string
}

func (e *UnsupportedMonitorTypeError) Error() string {
	return fmt.Sprintf("unsupported monitor type for retailer %s: %s", e.Retailer, e.MonitorType)
}

func IsValidMonitorType(monitorType, retailer string) error {
	valid := false
	for _, validMonitorType := range validMonitorTypes {
		if monitorType == validMonitorType {
			valid = true
			break
		}
	}
	if !valid {
		return &InvalidMonitorTypeError{monitorType}
	}

	supported := false
	for _, supportedMonitorType := range supportedMonitorTypes[retailer] {
		if monitorType == supportedMonitorType {
			supported = true
			break
		}
	}
	if !supported {
		return &UnsupportedMonitorTypeError{retailer, monitorType}
	}

	return nil
}

type EmptyInputError struct{}

func (e *EmptyInputError) Error() string {
	return "monitor input cannot be empty"
}

type InputIsNotURLError struct {
	Retailer string
}

func (e *InputIsNotURLError) Error() string {
	return fmt.Sprintf("monitor input is not a valid URL for retailer %s", e.Retailer)
}

type InvalidSKUError struct {
	Retailer string
	Format   string
}

func (e *InvalidSKUError) Error() string {
	return fmt.Sprintf("monitor input is not a valid SKU for retailer %s (format: %s)", e.Retailer, e.Format)
}

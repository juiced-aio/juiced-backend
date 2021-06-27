package pokemoncenter

import (
	"backend.juicedbot.io/juiced.infrastructure/common/enums"
	"backend.juicedbot.io/juiced.sitescripts/base"
)

// Endpoints
const (
	BaseEndpoint = "https://www.pokemoncenter.com/"

	MonitorEndpoint = ""
)

// Monitor info
type Monitor struct {
	Monitor     base.Monitor
	MonitorType enums.MonitorType
}

// Task info
type Task struct {
	Task base.Task
}

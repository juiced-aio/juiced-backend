package logging

import (
	"time"

	"backend.juicedbot.io/juiced.client/http"
	"backend.juicedbot.io/juiced.sitescripts/base"
)

type Log struct {
	Request      *http.Request
	RequestBody  string
	Response     *http.Response
	ResponseBody string
	Task         base.Task
	Monitor      base.Monitor
	Time         time.Time
}

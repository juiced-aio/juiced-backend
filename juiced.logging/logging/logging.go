package logging

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kirsle/configdir"
)

var loggingChannel = make(chan Log)
var writingChannel = make(chan Log)

func LoggingServer() {
	go func() {
		cachedLogs := make(map[string][]Log)

		for {
			log := <-loggingChannel

			uri := log.Request.URL.String()

			var shouldSaveLog bool
			if len(cachedLogs[uri]) != 0 {
				cachedLog := cachedLogs[uri][len(cachedLogs[uri])-1]
				if cachedLog.ResponseBody != "" {
					if log.ResponseBody != cachedLog.ResponseBody {
						shouldSaveLog = true
					}
				}
				if cachedLog.RequestBody != "" {
					if log.RequestBody != cachedLog.RequestBody {
						shouldSaveLog = true
					}
				}
				if log.Response.StatusCode != cachedLog.Response.StatusCode {
					shouldSaveLog = true
				}
			} else {
				shouldSaveLog = true
			}

			_, ok := log.Response.Header["Content-Type"]
			if ok {
				if strings.Contains(log.Response.Header["Content-Type"][0], "text/html") {
					// @silent: We can just do a max size here instead if you want
					log.ResponseBody = "HTML RESPONSE"
				}
			}

			_, ok = log.Response.Header["content-type"]
			if ok {
				if strings.Contains(log.Response.Header["content-type"][0], "text/html") {
					log.ResponseBody = "HTML RESPONSE"
				}
			}

			_, ok = cachedLogs[uri]
			if ok {
				log.Request.RawHeader = compareRawHeaders(log.Request.RawHeader, cachedLogs[uri][0].Request.RawHeader)
				log.Response.Header = compareHeaders(log.Response.Header, cachedLogs[uri][0].Response.Header)
			}

			if shouldSaveLog {
				cachedLogs[uri] = append(cachedLogs[uri], log)
				writingChannel <- log
			}

		}
	}()
	go func() {
		configPath := configdir.LocalConfig("juiced", "logs")
		err := configdir.MakePath(configPath)
		if err != nil {
			fmt.Println(err)
			return
		}

		for {
			log := <-writingChannel

			var logType string
			if log.Monitor.TaskGroup != nil {
				if log.Monitor.TaskGroup.Name != "" {
					logType = "MONITOR"
				}
			} else if log.Task.Task != nil {
				if log.Task.Task.ID != "" {
					logType = "TASK"
				}
			}

			var filename string
			var otherInfo string
			switch logType {
			case "TASK":
				filename = filepath.Join(configPath, log.Task.Task.ID)
				data, _ := json.Marshal(log.Task.Task)
				otherInfo = string(data)
			case "MONITOR":
				filename = filepath.Join(configPath, log.Monitor.TaskGroup.GroupID)
				data, _ := json.Marshal(log.Task)
				otherInfo = string(data)
			default:
				filename = filepath.Join(configPath, "OTHER")
			}

			f, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
			if err != nil {
				fmt.Println(err)
				continue
			}

			rawHeadersString := ""
			for _, header := range log.Request.RawHeader {
				rawHeadersString += header[0] + ": " + header[1] + "\n"
			}
			headersString := ""
			for header, values := range log.Response.Header {
				headersString += header + ": " + strings.Join(values, ",") + "\n"
			}
			logTemplate := "\n================\n%v\nSTART %v %v\nREQUEST BODY: %v\nREQUEST HEADERS:\n%v\nRESPONSE STATUSCODE: %v\nRESPONSE BODY: %v\nRESPONSE HEADERS:\n%v\nOTHER INFO:%v\nEND %v\n================\n"

			f.WriteString(fmt.Sprintf(logTemplate, log.Time.String(), log.Request.Method, log.Request.URL.String(), log.RequestBody, rawHeadersString, log.Response.StatusCode, log.ResponseBody, headersString, otherInfo, log.Request.URL.String()))
			f.Close()
		}
	}()

}

func SendLog(l Log) {
	loggingChannel <- l
}

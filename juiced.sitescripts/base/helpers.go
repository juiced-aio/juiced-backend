package base

import (
	"backend.juicedbot.io/juiced.client/http/cookiejar"

	"backend.juicedbot.io/juiced.client/client"

	"backend.juicedbot.io/juiced.client/utls"
	"backend.juicedbot.io/juiced.infrastructure/common"
	"backend.juicedbot.io/juiced.infrastructure/common/entities"
)

func (task *Task) UpdateProxy(proxy *entities.Proxy) error {
	task.Proxy.RemoveCount()
	if proxy != nil {
		err := client.UpdateProxy(&task.Client, proxy)
		if err != nil {
			return err
		}
		task.Proxy = proxy
	}

	return nil
}

func (monitor *Monitor) UpdateProxy(proxy *entities.Proxy) error {
	monitor.Proxy.RemoveCount()
	if proxy != nil {
		err := client.UpdateProxy(&monitor.Client, proxy)
		if err != nil {
			return err
		}
		monitor.Proxy = proxy
	}

	return nil
}

// CreateClient creates an HTTP client
func (task *Task) CreateClient(proxy ...*entities.Proxy) error {
	var err error
	task.Proxy.RemoveCount()
	if len(proxy) > 0 {
		if proxy[0] != nil {
			proxy[0].AddCount()
			task.Client, err = client.NewClient(utls.HelloChrome_90, common.ProxyCleaner(*proxy[0]))
			if err != nil {
				return err
			}
		} else {
			task.Client, _ = client.NewClient(utls.HelloChrome_90)
		}
	} else {
		task.Client, _ = client.NewClient(utls.HelloChrome_90)
	}
	cookieJar, err := cookiejar.New(nil)
	if err != nil {
		return err
	}
	task.Client.Jar = cookieJar
	return err
}

func (monitor *Monitor) CreateClient(proxy ...*entities.Proxy) error {
	var err error
	monitor.Proxy.RemoveCount()
	if len(proxy) > 0 {
		if proxy[0] != nil {
			proxy[0].AddCount()
			monitor.Client, err = client.NewClient(utls.HelloChrome_90, common.ProxyCleaner(*proxy[0]))
			if err != nil {
				return err
			}
		} else {
			monitor.Client, _ = client.NewClient(utls.HelloChrome_90)
		}
	} else {
		monitor.Client, _ = client.NewClient(utls.HelloChrome_90)
	}
	cookieJar, err := cookiejar.New(nil)
	if err != nil {
		return err
	}
	monitor.Client.Jar = cookieJar
	return err
}

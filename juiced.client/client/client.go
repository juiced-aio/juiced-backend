package client

import (
	"log"

	utls "backend.juicedbot.io/juiced.client/utls"
	"backend.juicedbot.io/juiced.infrastructure/common"
	"backend.juicedbot.io/juiced.infrastructure/common/entities"
	"golang.org/x/net/proxy"

	"backend.juicedbot.io/juiced.client/http"
)

func UpdateProxy(client *http.Client, newProxy *entities.Proxy) error {
	if newProxy == nil || newProxy.Host == "" {
		client.Transport = newRoundTripper(utls.HelloChrome_90, proxy.Direct)
		return nil
	}
	newProxy.AddCount()
	dialer, err := newConnectDialer(common.ProxyCleaner(*newProxy))
	if err != nil {
		return err
	}
	client.Transport = newRoundTripper(utls.HelloChrome_90, dialer)
	return nil

}

func NewClient(clientHello utls.ClientHelloID, proxyUrl ...string) (http.Client, error) {
	if len(proxyUrl) > 0 {
		if len(proxyUrl[0]) > 0 {
			dialer, err := newConnectDialer(proxyUrl[0])
			if err != nil {
				return http.Client{}, err
			}
			return http.Client{
				Transport: newRoundTripper(clientHello, dialer),
			}, nil
		} else {
			return http.Client{
				Transport: newRoundTripper(clientHello, proxy.Direct),
			}, nil
		}
	} else {
		log.Println(2)
		return http.Client{
			Transport: newRoundTripper(clientHello, proxy.Direct),
		}, nil
	}
}

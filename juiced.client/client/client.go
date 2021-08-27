package client

import (
	utls "backend.juicedbot.io/juiced.client/utls"
	"backend.juicedbot.io/juiced.infrastructure/entities"
	"backend.juicedbot.io/juiced.infrastructure/util"
	"golang.org/x/net/proxy"

	"backend.juicedbot.io/juiced.client/http"
)

func UpdateProxy(client *http.Client, proxy *entities.Proxy) error {
	if proxy == nil || proxy.Host == "" {
		return nil
	}
	proxy.AddCount()
	dialer, err := newConnectDialer(util.ProxyCleaner(*proxy))
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
		return http.Client{
			Transport: newRoundTripper(clientHello, proxy.Direct),
		}, nil
	}
}

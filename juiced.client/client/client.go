package client

import (
	utls "github.com/Titanium-ctrl/utls"
	"golang.org/x/net/proxy"

	"backend.juicedbot.io/juiced.client/http"
)

func UpdateProxy(client *http.Client, proxyurl string) error {
	dialer, err := newConnectDialer(proxyurl)
	if err != nil {
		return err
	}
	client.Transport = newRoundTripper(utls.HelloChrome_83, dialer)
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

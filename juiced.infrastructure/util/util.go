package util

import (
	"fmt"

	"backend.juicedbot.io/juiced.infrastructure/entities"
)

func ProxyCleaner(proxyDirty entities.Proxy) string {
	if proxyDirty.Host == "" {
		return ""
	}
	if proxyDirty.Username == "" && proxyDirty.Password == "" {
		return fmt.Sprintf("http://%s:%s", proxyDirty.Host, proxyDirty.Port)
	} else {
		return fmt.Sprintf("http://%s:%s@%s:%s", proxyDirty.Username, proxyDirty.Password, proxyDirty.Host, proxyDirty.Port)
	}
}

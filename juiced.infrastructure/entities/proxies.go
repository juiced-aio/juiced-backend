package entities

import (
	"encoding/json"

	"github.com/google/uuid"
)

func (proxy *Proxy) UnmarshalJSON(data []byte) error {
	type ProxyAlias Proxy

	temp := &ProxyAlias{ID: uuid.New().String()}

	err := json.Unmarshal(data, temp)
	if err != nil {
		return err
	}
	*proxy = Proxy(*temp)
	return nil
}

type Proxy struct {
	ID           string `json:"ID" db:"ID"`
	ProxyGroupID string `json:"proxyGroupID" db:"proxyGroupID"`
	Host         string `json:"host" db:"host"`
	Port         string `json:"port" db:"port"`
	Username     string `json:"username" db:"username"`
	Password     string `json:"password" db:"password"`
	Count        int
}

func (proxy *Proxy) RemoveCount() {
	if proxy != nil {
		proxy.Count--
	}
}

func (proxy *Proxy) AddCount() {
	if proxy != nil {
		proxy.Count++
	}
}

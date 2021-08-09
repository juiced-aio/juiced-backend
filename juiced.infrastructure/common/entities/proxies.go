package entities

import (
	"encoding/json"

	"github.com/google/uuid"
)

// UnmarshalJSON is a Proxy's redefinition of the default UnmarshalJSON function
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

// Proxy is a class that holds details about a single proxy
type Proxy struct {
	ID           string `json:"ID" db:"ID"`
	ProxyGroupID string `json:"proxyGroupID" db:"proxyGroupID"`
	Host         string `json:"host" db:"host"`
	Port         string `json:"port" db:"port"`
	Username     string `json:"username" db:"username"`
	Password     string `json:"password" db:"password"`
	Count        int
}

// SetID updates the Proxy's ID
func (proxy *Proxy) SetID(ID string) {
	proxy.ID = ID
}

// RemoveCount subtracts one from count
func (proxy *Proxy) RemoveCount() {
	if proxy != nil {
		proxy.Count--
	}
}

// AddCount adds one to count
func (proxy *Proxy) AddCount() {
	if proxy != nil {
		proxy.AddCount()
	}
}

// ProxyGroup is a class that holds a list of proxies
type ProxyGroup struct {
	GroupID      string   `json:"groupID" db:"groupID"`
	Name         string   `json:"name" db:"name"`
	Proxies      []*Proxy `json:"proxies"`
	CreationDate int64    `json:"creationDate" db:"creationDate"`
}

// SetName updates the ProxyGroup's Name
func (proxyGroup *ProxyGroup) SetName(name string) {
	proxyGroup.Name = name
}

// SetGroupID updates the ProxyGroup's GroupID
func (proxyGroup *ProxyGroup) SetGroupID(GroupID string) {
	proxyGroup.GroupID = GroupID
}

// ParseProxyGroup returns a ProxyGroup object parsed from a JSON bytes array
func ParseProxyGroup(proxyGroup *ProxyGroup, data []byte) error {
	err := json.Unmarshal(data, &proxyGroup)
	return err
}

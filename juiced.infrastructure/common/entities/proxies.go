package entities

import (
	"encoding/json"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UnmarshalJSON is a Proxy's redefinition of the default UnmarshalJSON function
func (proxy *Proxy) UnmarshalJSON(data []byte) error {
	type ProxyAlias Proxy
	temp := &ProxyAlias{ID: primitive.NewObjectID()}

	err := json.Unmarshal(data, temp)
	if err != nil {
		return err
	}
	*proxy = Proxy(*temp)
	return nil
}

// Proxy is a class that holds details about a single proxy
type Proxy struct {
	ID       primitive.ObjectID `json:"ID"`
	Host     string             `json:"host"`
	Port     string             `json:"port"`
	Username string             `json:"username"`
	Password string             `json:"password"`
}

// SetID updates the Proxy's ID
func (proxy *Proxy) SetID(ID primitive.ObjectID) {
	proxy.ID = ID
}

// ProxyGroup is a class that holds a list of proxies
type ProxyGroup struct {
	GroupID primitive.ObjectID `json:"groupID"`
	Name    string             `json:"name"`
	Proxies []Proxy            `json:"proxies"`
}

// ProxyGroups is a class that holdes multiple ProxyGroups
type ProxyGroups struct {
	ProxyGroups []ProxyGroup `json:"proxyGroups"`
}

// SetName updates the ProxyGroup's Name
func (proxyGroup *ProxyGroup) SetName(name string) {
	proxyGroup.Name = name
}

// SetGroupID updates the ProxyGroup's GroupID
func (proxyGroup *ProxyGroup) SetGroupID(GroupID primitive.ObjectID) {
	proxyGroup.GroupID = GroupID
}

// ParseProxyGroup returns a ProxyGroup object parsed from a JSON bytes array
func ParseProxyGroup(proxyGroup *ProxyGroup, data []byte) error {
	err := json.Unmarshal(data, &proxyGroup)
	return err
}

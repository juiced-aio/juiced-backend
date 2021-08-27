package entities

type ProxyGroup struct {
	GroupID      string   `json:"groupID" db:"groupID"`
	Name         string   `json:"name" db:"name"`
	Proxies      []*Proxy `json:"proxies"`
	CreationDate int64    `json:"creationDate" db:"creationDate"`
}

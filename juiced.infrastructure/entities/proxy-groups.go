package entities

import (
	"math/rand"
	"sort"
)

type ProxyGroup struct {
	GroupID      string   `json:"groupID" db:"groupID"`
	Name         string   `json:"name" db:"name"`
	Proxies      []*Proxy `json:"proxies"`
	CreationDate int64    `json:"creationDate" db:"creationDate"`
}

func (proxyGroup *ProxyGroup) GetRandomLeastUsedProxy() *Proxy {
	if len(proxyGroup.Proxies) == 0 {
		return nil
	}

	countMap := make(map[int][]*Proxy)
	for _, proxy := range proxyGroup.Proxies {
		countMap[proxy.Count] = append(countMap[proxy.Count], proxy)
	}

	var proxyCounts []int
	for key := range countMap {
		proxyCounts = append(proxyCounts, key)
	}
	sort.Ints(proxyCounts)

	return countMap[proxyCounts[0]][rand.Intn(len(countMap[proxyCounts[0]]))]
}

package stores

import (
	"backend.juicedbot.io/juiced.infrastructure/common/entities"
)

// ProxyStore stores information about loaded proxies
type ProxyStore struct {
	ProxyGroups map[string]*entities.ProxyGroup
}

func (proxyStore *ProxyStore) AddProxyGroup(proxyGroup *entities.ProxyGroup) {
	proxyStore.ProxyGroups[proxyGroup.GroupID] = proxyGroup
}

func (proxyStore *ProxyStore) UpdateProxyGroup(groupID string, proxyGroup *entities.ProxyGroup) {
	// @silent: This will strictly update the group, want to also add it if it doesn't exist?
	if _, ok := proxyStore.ProxyGroups[groupID]; ok {
		proxyStore.ProxyGroups[groupID] = proxyGroup
	}
}

func (proxyStore *ProxyStore) RemoveProxyGroup(groupID string) {
	delete(proxyStore.ProxyGroups, groupID)
}

func (proxyStore *ProxyStore) GetProxyGroup(groupID string) (*entities.ProxyGroup, bool) {
	proxyGroup, ok := proxyStore.ProxyGroups[groupID]
	return proxyGroup, ok
}

var proxyStore *ProxyStore

// InitProxyStore initializes the singleton instance of the ProxyStore
func InitProxyStore() {
	proxyStore = &ProxyStore{
		ProxyGroups: make(map[string]*entities.ProxyGroup),
	}
}

// GetProxyStore returns the singleton instance of the ProxyStore
func GetProxyStore() *ProxyStore {
	return proxyStore
}

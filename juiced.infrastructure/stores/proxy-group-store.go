package stores

import (
	"fmt"
	"time"

	"backend.juicedbot.io/juiced.infrastructure/database"
	"backend.juicedbot.io/juiced.infrastructure/entities"
	"github.com/google/uuid"
)

type ProxyGroupStore struct {
	ProxyGroups map[string]*entities.ProxyGroup
}

var proxyGroupStore ProxyGroupStore

func (store *ProxyGroupStore) Init() error {
	proxyGroups, err := database.GetAllProxyGroups()
	if err != nil {
		return err
	}

	for _, proxyGroup := range proxyGroups {
		store.ProxyGroups[proxyGroup.GroupID] = &proxyGroup
	}

	return nil
}

type ProxyGroupNotFoundError struct {
	ID string
}

func (e *ProxyGroupNotFoundError) Error() string {
	return fmt.Sprintf("ProxyGroup with ID %s not found", e.ID)
}

func GetAllProxyGroups() []*entities.ProxyGroup {
	proxyGroups := []*entities.ProxyGroup{}

	for _, proxyGroup := range proxyGroupStore.ProxyGroups {
		proxyGroups = append(proxyGroups, proxyGroup)
	}

	return proxyGroups
}

func GetProxyGroup(groupID string) (*entities.ProxyGroup, error) {
	proxyGroup, ok := proxyGroupStore.ProxyGroups[groupID]
	if !ok {
		return nil, &ProxyGroupNotFoundError{groupID}
	}

	return proxyGroup, nil
}

func CreateProxyGroup(proxyGroup entities.ProxyGroup) (*entities.ProxyGroup, error) {
	if proxyGroup.GroupID == "" {
		proxyGroup.GroupID = uuid.New().String()
	}
	if proxyGroup.CreationDate == 0 {
		proxyGroup.CreationDate = time.Now().Unix()
	}

	err := database.CreateProxyGroup(proxyGroup)
	if err != nil {
		return nil, err
	}

	proxyGroupPtr := &proxyGroup
	proxyGroupStore.ProxyGroups[proxyGroup.GroupID] = proxyGroupPtr

	return proxyGroupPtr, nil
}

func UpdateProxyGroup(groupID string, newProxyGroup entities.ProxyGroup) (*entities.ProxyGroup, error) {
	proxyGroup, err := GetProxyGroup(groupID)
	if err != nil {
		return nil, err
	}

	proxyGroup.Name = newProxyGroup.Name
	proxyGroup.Proxies = newProxyGroup.Proxies

	return proxyGroup, database.UpdateProxyGroup(groupID, *proxyGroup)
}

func RemoveProxyGroup(groupID string) (entities.ProxyGroup, error) {
	proxyGroup, err := GetProxyGroup(groupID)
	if err != nil {
		return entities.ProxyGroup{}, err
	}

	delete(proxyGroupStore.ProxyGroups, groupID)
	return *proxyGroup, database.RemoveProxyGroup(groupID)
}

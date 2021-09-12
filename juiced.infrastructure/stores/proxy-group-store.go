package stores

import (
	"fmt"
	"sort"
	"time"

	"backend.juicedbot.io/juiced.infrastructure/database"
	"backend.juicedbot.io/juiced.infrastructure/entities"
	"backend.juicedbot.io/juiced.infrastructure/util"
	"github.com/google/uuid"
)

type ProxyGroupStore struct {
	ProxyGroups map[string]*entities.ProxyGroup
}

var proxyGroupStore ProxyGroupStore

func InitProxyGroupStore() error {
	proxyGroupStore = ProxyGroupStore{
		ProxyGroups: make(map[string]*entities.ProxyGroup),
	}

	proxyGroups, err := database.GetAllProxyGroups()
	if err != nil {
		return err
	}

	for _, proxyGroup := range proxyGroups {
		proxyGroup := proxyGroup
		proxyGroupStore.ProxyGroups[proxyGroup.GroupID] = &proxyGroup
	}

	return nil
}

type ProxyGroupNotFoundError struct {
	ID string
}

func (e *ProxyGroupNotFoundError) Error() string {
	return fmt.Sprintf("ProxyGroup with ID %s not found", e.ID)
}

type ProxyGroupNotFoundByNameError struct {
	Name string
}

func (e *ProxyGroupNotFoundByNameError) Error() string {
	return fmt.Sprintf("ProxyGroup with name %s not found", e.Name)
}

type ProxyGroupAlreadyExistsError struct {
	Name string
}

func (e *ProxyGroupAlreadyExistsError) Error() string {
	return fmt.Sprintf("ProxyGroup with name %s already exists", e.Name)
}

func GetAllProxyGroups() []*entities.ProxyGroup {
	proxyGroups := []*entities.ProxyGroup{}

	for _, proxyGroup := range proxyGroupStore.ProxyGroups {
		proxyGroups = append(proxyGroups, proxyGroup)
	}

	sort.SliceStable(proxyGroups, func(i, j int) bool {
		return proxyGroups[i].CreationDate < proxyGroups[j].CreationDate
	})

	return proxyGroups
}

func GetProxyGroup(groupID string) (*entities.ProxyGroup, error) {
	proxyGroup, ok := proxyGroupStore.ProxyGroups[groupID]
	if !ok {
		return nil, &ProxyGroupNotFoundError{groupID}
	}

	return proxyGroup, nil
}

func GetProxyGroupByName(name string) (*entities.ProxyGroup, error) {
	proxyGroups := GetAllProxyGroups()
	for _, proxyGroup := range proxyGroups {
		if proxyGroup.Name == name {
			return proxyGroup, nil
		}
	}
	return nil, &ProxyGroupNotFoundByNameError{name}
}

func CreateProxyGroup(proxyGroup entities.ProxyGroup) (*entities.ProxyGroup, error) {
	_, err := GetProxyGroupByName(proxyGroup.Name)
	if err == nil {
		return nil, &ProxyGroupAlreadyExistsError{proxyGroup.Name}
	}

	if proxyGroup.GroupID == "" {
		proxyGroup.GroupID = uuid.New().String()
	}
	if proxyGroup.CreationDate == 0 {
		proxyGroup.CreationDate = time.Now().Unix()
	}

	for _, proxy := range proxyGroup.Proxies {
		proxy.ID = uuid.New().String()
		proxy.ProxyGroupID = proxyGroup.GroupID
		proxy.CreationDate = time.Now().Unix()
	}

	err = database.CreateProxyGroup(proxyGroup)
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

	// TODO: Update Tasks/TaskGroups with this ProxyGroup to change proxies and restart

	return proxyGroup, database.UpdateProxyGroup(groupID, *proxyGroup)
}

func RemoveProxyGroup(groupID string) (entities.ProxyGroup, error) {
	proxyGroup, err := GetProxyGroup(groupID)
	if err != nil {
		return entities.ProxyGroup{}, err
	}

	delete(proxyGroupStore.ProxyGroups, groupID)

	// TODO: Update Tasks/TaskGroups with this ProxyGroup to use no ProxyGroup

	return *proxyGroup, database.RemoveProxyGroup(groupID)
}

func CloneProxyGroup(groupID string) (*entities.ProxyGroup, error) {
	proxyGroupPtr, err := GetProxyGroup(groupID)
	if err != nil {
		return nil, err
	}

	newProxyGroup := *proxyGroupPtr
	newProxyGroup.Name += " " + util.RandString(3)
	newProxyGroup.GroupID = ""
	newProxyGroup.CreationDate = 0

	newProxies := []*entities.Proxy{}
	for _, proxy := range newProxyGroup.Proxies {
		newProxy := *proxy
		newProxies = append(newProxies, &newProxy)
	}
	newProxyGroup.Proxies = newProxies

	return CreateProxyGroup(newProxyGroup)
}

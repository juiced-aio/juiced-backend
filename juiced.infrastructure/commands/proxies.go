package commands

import (
	"errors"

	"backend.juicedbot.io/juiced.infrastructure/common"
	"backend.juicedbot.io/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/juiced.infrastructure/queries"
	_ "github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

// CreateProxyGroup adds the ProxyGroup object to the database
func CreateProxyGroup(proxyGroup entities.ProxyGroup) error {
	database := common.GetDatabase()
	if database == nil {
		return errors.New("database not initialized")
	}

	statement, err := database.Preparex(`INSERT INTO proxyGroups (groupID, name) VALUES (?, ?)`)
	if err != nil {
		return err
	}

	_, err = statement.Exec(proxyGroup.GroupID, proxyGroup.Name)
	if err != nil {
		return err
	}

	for _, proxy := range proxyGroup.Proxies {
		statement, err := database.Preparex(`INSERT INTO proxys (ID, proxyGroupID, host, port, username, password) VALUES (?, ?, ?, ?, ?, ?)`)
		if err != nil {
			return err
		}
		_, err = statement.Exec(proxy.ID, proxyGroup.GroupID, proxy.Host, proxy.Port, proxy.Username, proxy.Password)
		if err != nil {
			return err
		}
	}

	return err
}

// RemoveProxyGroup removes the ProxyGroup from the database with the given groupID and returns it (if it exists)
func RemoveProxyGroup(groupID string) (entities.ProxyGroup, error) {
	proxyGroup := entities.ProxyGroup{}
	database := common.GetDatabase()
	if database == nil {
		return proxyGroup, errors.New("database not initialized")
	}

	proxyGroup, err := queries.GetProxyGroup(groupID)
	if err != nil {
		return proxyGroup, err
	}

	statement, err := database.Preparex(`DELETE FROM proxyGroups WHERE groupID = @p1`)
	if err != nil {
		return proxyGroup, err
	}
	_, err = statement.Exec(groupID)
	if err != nil {
		return proxyGroup, err
	}

	statement, err = database.Preparex(`DELETE FROM proxys WHERE proxyGroupID = @p1`)
	if err != nil {
		return proxyGroup, err
	}
	_, err = statement.Exec(groupID)

	return proxyGroup, err
}

// UpdateProxyGroup updates the ProxyGroup from the database with the given groupID and returns it (if it exists)
func UpdateProxyGroup(groupID string, newProxyGroup entities.ProxyGroup) (entities.ProxyGroup, error) {
	proxyGroup, err := RemoveProxyGroup(groupID)
	if err != nil {
		return proxyGroup, err
	}

	err = CreateProxyGroup(newProxyGroup)
	if err != nil {
		return proxyGroup, err
	}

	return queries.GetProxyGroup(groupID)
}

package queries

import (
	"errors"

	"backend.juicedbot.io/juiced.infrastructure/common"
	"backend.juicedbot.io/juiced.infrastructure/common/entities"
)

// GetAllProxyGroups returns all ProxyGroup objects from the database
func GetAllProxyGroups() ([]entities.ProxyGroup, error) {
	proxyGroups := []entities.ProxyGroup{}
	database := common.GetDatabase()
	if database == nil {
		return proxyGroups, errors.New("database not initialized")
	}

	rows, err := database.Queryx("SELECT * FROM proxyGroups")
	if err != nil {
		return proxyGroups, err
	}

	defer rows.Close()
	for rows.Next() {
		tempProxyGroup := entities.ProxyGroup{}
		err = rows.StructScan(&tempProxyGroup)
		if err != nil {
			return proxyGroups, err
		}
		tempProxyGroup, err = GetProxies(tempProxyGroup)
		if err != nil {
			return proxyGroups, err
		}
		proxyGroups = append(proxyGroups, tempProxyGroup)

	}
	return proxyGroups, err
}

// GetProxyGroup returns the ProxyGroup object from the database with the given groupID (if it exists)
func GetProxyGroup(groupID string) (entities.ProxyGroup, error) {
	proxyGroup := entities.ProxyGroup{}
	database := common.GetDatabase()
	if database == nil {
		return proxyGroup, errors.New("database not initialized")
	}

	statement, err := database.Preparex("SELECT * FROM proxyGroups WHERE groupID = @p1")
	if err != nil {
		return proxyGroup, err
	}

	rows, err := statement.Queryx(groupID)
	if err != nil {
		return proxyGroup, err
	}

	defer rows.Close()
	for rows.Next() {
		err = rows.StructScan(&proxyGroup)
		if err != nil {
			return proxyGroup, err
		}
	}

	return GetProxies(proxyGroup)
}

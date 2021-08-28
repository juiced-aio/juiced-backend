package database

import (
	"sort"

	"backend.juicedbot.io/juiced.infrastructure/entities"
)

func GetAllProxyGroups() ([]entities.ProxyGroup, error) {
	proxyGroups := []entities.ProxyGroup{}
	if database == nil {
		return proxyGroups, &DatabaseNotInitializedError{}
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

	sort.SliceStable(proxyGroups, func(i, j int) bool {
		return proxyGroups[i].CreationDate < proxyGroups[j].CreationDate
	})

	return proxyGroups, err
}

func GetProxyGroup(groupID string) (entities.ProxyGroup, error) {
	proxyGroup := entities.ProxyGroup{}
	if database == nil {
		return proxyGroup, &DatabaseNotInitializedError{}
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

func CreateProxyGroup(proxyGroup entities.ProxyGroup) error {
	if database == nil {
		return &DatabaseNotInitializedError{}
	}

	statement, err := database.Preparex(`INSERT INTO proxyGroups (groupID, name, creationDate) VALUES (?, ?, ?)`)
	if err != nil {
		return err
	}

	_, err = statement.Exec(proxyGroup.GroupID, proxyGroup.Name, proxyGroup.CreationDate)
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

func RemoveProxyGroup(groupID string) error {
	if database == nil {
		return &DatabaseNotInitializedError{}
	}

	statement, err := database.Preparex(`DELETE FROM proxyGroups WHERE groupID = @p1`)
	if err != nil {
		return err
	}
	_, err = statement.Exec(groupID)
	if err != nil {
		return err
	}

	statement, err = database.Preparex(`DELETE FROM proxys WHERE proxyGroupID = @p1`)
	if err != nil {
		return err
	}
	_, err = statement.Exec(groupID)

	return err
}

func UpdateProxyGroup(groupID string, newProxyGroup entities.ProxyGroup) error {
	err := RemoveProxyGroup(groupID)
	if err != nil {
		return err
	}
	return CreateProxyGroup(newProxyGroup)
}

func GetProxies(proxyGroup entities.ProxyGroup) (entities.ProxyGroup, error) {
	if database == nil {
		return proxyGroup, &DatabaseNotInitializedError{}
	}

	statement, err := database.Preparex(`SELECT * FROM proxys WHERE proxyGroupID = @p1`)
	if err != nil {
		return proxyGroup, err
	}
	rows, err := statement.Queryx(proxyGroup.GroupID)
	if err != nil {
		return proxyGroup, err
	}

	defer rows.Close()
	for rows.Next() {
		tempProxy := entities.Proxy{}
		err = rows.StructScan(&tempProxy)
		if err != nil {
			return proxyGroup, err
		}
		proxyGroup.Proxies = append(proxyGroup.Proxies, &tempProxy)
	}

	return proxyGroup, err
}

package database

import (
	"sort"

	"backend.juicedbot.io/juiced.infrastructure/entities"
)

func GetProxies(proxyGroup entities.ProxyGroup) (entities.ProxyGroup, error) {
	if database == nil {
		return proxyGroup, &DatabaseNotInitializedError{}
	}

	statement, err := database.Preparex(`SELECT * FROM proxies WHERE proxyGroupID = @p1`)
	if err != nil {
		return proxyGroup, err
	}
	rows, err := statement.Queryx(proxyGroup.GroupID)
	if err != nil {
		return proxyGroup, err
	}

	proxies := []*entities.Proxy{}
	defer rows.Close()
	for rows.Next() {
		tempProxy := entities.Proxy{}
		err = rows.StructScan(&tempProxy)
		if err != nil {
			return proxyGroup, err
		}
		proxies = append(proxies, &tempProxy)
	}

	sort.SliceStable(proxies, func(i, j int) bool {
		return proxies[i].CreationDate < proxies[j].CreationDate
	})

	proxyGroup.Proxies = proxies

	return proxyGroup, err
}

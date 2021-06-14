package queries

import (
	"errors"

	"backend.juicedbot.io/juiced.infrastructure/common"
	"backend.juicedbot.io/juiced.infrastructure/common/entities"
)

// GetSettings returns the settings object from the database
func GetSettings() (entities.Settings, error) {
	settings := entities.Settings{}
	database := common.GetDatabase()
	if database == nil {
		return settings, errors.New("database not initialized")
	}

	// Might want to add "WHERE id = 0" to the query
	rows, err := database.Queryx("SELECT * FROM settings")
	if err != nil {
		return settings, err
	}

	defer rows.Close()
	for rows.Next() {
		err = rows.StructScan(&settings)
		if err != nil {
			return settings, err
		}
	}
	return settings, err
}

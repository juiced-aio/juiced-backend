package database

import (
	"errors"
	"sort"
	"strings"

	"backend.juicedbot.io/juiced.infrastructure/entities"
	"backend.juicedbot.io/juiced.infrastructure/helpers"
)

// GetAllProfileGroups returns all ProfileGroup objects from the database
func GetAllProfileGroups() ([]entities.ProfileGroup, error) {
	profileGroups := []entities.ProfileGroup{}
	if database == nil {
		return profileGroups, errors.New("database not initialized")
	}

	rows, err := database.Queryx("SELECT * FROM profileGroups")
	if err != nil {
		return profileGroups, err
	}

	defer rows.Close()
	for rows.Next() {
		tempProfileGroup := entities.ProfileGroup{}
		err = rows.StructScan(&tempProfileGroup)
		if err != nil {
			return profileGroups, err
		}

		if tempProfileGroup.ProfileIDsJoined != "" {
			tempProfileGroup.ProfileIDs = strings.Split(tempProfileGroup.ProfileIDsJoined, ",")
		}

		profileGroups = append(profileGroups, tempProfileGroup)
	}
	sort.SliceStable(profileGroups, func(i, j int) bool {
		return profileGroups[i].CreationDate < profileGroups[j].CreationDate
	})
	return profileGroups, err
}

// GetProfileGroup returns the ProfileGroup object from the database with the given groupID (if it exists)
func GetProfileGroup(groupID string) (entities.ProfileGroup, error) {
	profileGroup := entities.ProfileGroup{}
	if database == nil {
		return profileGroup, &DatabaseNotInitializedError{}
	}

	statement, err := database.Preparex("SELECT * FROM profileGroups WHERE groupID = @p1")
	if err != nil {
		return profileGroup, err
	}

	rows, err := statement.Queryx(groupID)
	if err != nil {
		return profileGroup, err
	}

	defer rows.Close()
	for rows.Next() {
		err = rows.StructScan(&profileGroup)
		if err != nil {
			return profileGroup, err
		}
	}

	if profileGroup.ProfileIDsJoined != "" {
		profileGroup.ProfileIDs = strings.Split(profileGroup.ProfileIDsJoined, ",")
	}

	return profileGroup, err
}

func CreateProfileGroup(profileGroup entities.ProfileGroup) error {
	if database == nil {
		return &DatabaseNotInitializedError{}
	}

	statement, err := database.Preparex(`INSERT INTO profileGroups (groupID, name, profileIDsJoined, creationDate) VALUES (?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	profileIDsJoined := strings.Join(profileGroup.ProfileIDs, ",")

	_, err = statement.Exec(profileGroup.GroupID, profileGroup.Name, profileIDsJoined, profileGroup.CreationDate)
	if err != nil {
		return err
	}

	for _, profileID := range profileGroup.ProfileIDs {
		profile, err := GetProfile(profileID)
		if err != nil {
			return err
		}
		profile.ProfileGroupIDs = append(profile.ProfileGroupIDs, profileGroup.GroupID)
		err = UpdateProfile(profile.ID, profile)
		if err != nil {
			return err
		}
	}

	return err
}

func RemoveProfileGroup(groupID string) error {
	if database == nil {
		return &DatabaseNotInitializedError{}
	}

	profileGroup, err := GetProfileGroup(groupID)
	if err != nil {
		return err
	}

	statement, err := database.Preparex(`DELETE FROM profileGroups WHERE groupID = @p1`)
	if err != nil {
		return err
	}
	_, err = statement.Exec(groupID)

	if profileGroup.ProfileIDsJoined != "" {
		profileGroup.ProfileIDs = strings.Split(profileGroup.ProfileIDsJoined, ",")
	}

	for _, profileID := range profileGroup.ProfileIDs {
		profile, err := GetProfile(profileID)
		if err != nil {
			return err
		}

		profile.ProfileGroupIDs = helpers.RemoveFromSlice(profile.ProfileGroupIDs, groupID)

		err = UpdateProfile(profile.ID, profile)
		if err != nil {
			return err
		}
	}

	return err
}

func UpdateProfileGroup(groupID string, newProfileGroup entities.ProfileGroup) error {
	err := RemoveProfileGroup(groupID)
	if err != nil {
		return err
	}
	err = CreateProfileGroup(newProfileGroup)
	return err
}

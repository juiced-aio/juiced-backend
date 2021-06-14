package commands

import (
	"errors"
	"strings"

	"backend.juicedbot.io/juiced.infrastructure/common"
	"backend.juicedbot.io/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/juiced.infrastructure/queries"
	_ "github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

// CreateProfileGroup adds the ProfileGroup object to the database
func CreateProfileGroup(profileGroup entities.ProfileGroup) error {
	database := common.GetDatabase()
	if database == nil {
		return errors.New("database not initialized")
	}

	statement, err := database.Preparex(`INSERT INTO profileGroups (groupID, name, profileIDsJoined) VALUES (?, ?, ?)`)
	if err != nil {
		return err
	}
	profileIDsJoined := strings.Join(profileGroup.ProfileIDs, ",")

	_, err = statement.Exec(profileGroup.GroupID, profileGroup.Name, profileIDsJoined)
	if err != nil {
		return err
	}

	for _, profileID := range profileGroup.ProfileIDs {
		profile, err := queries.GetProfile(profileID)
		if err != nil {
			return err
		}
		profile.ProfileGroupID = profileGroup.GroupID
		_, err = UpdateProfile(profile.ID, profile)
		if err != nil {
			return err
		}
	}

	return err
}

// RemoveProfileGroup removes the ProfileGroup from the database with the given groupID and returns it (if it exists)
func RemoveProfileGroup(groupID string) (entities.ProfileGroup, error) {
	profileGroup := entities.ProfileGroup{}
	database := common.GetDatabase()
	if database == nil {
		return profileGroup, errors.New("database not initialized")
	}

	profileGroup, err := queries.GetProfileGroup(groupID)
	if err != nil {
		return profileGroup, err
	}

	statement, err := database.Preparex(`DELETE FROM profileGroups WHERE groupID = @p1`)
	if err != nil {
		return profileGroup, err
	}
	_, err = statement.Exec(groupID)

	profileGroup.ProfileIDs = strings.Split(profileGroup.ProfileIDsJoined, ",")

	for _, profileID := range profileGroup.ProfileIDs {
		profile, err := queries.GetProfile(profileID)
		if err != nil {
			return profileGroup, err
		}
		profile.ProfileGroupID = ""
		_, err = UpdateProfile(profile.ID, profile)
		if err != nil {
			return profileGroup, err
		}
	}

	return profileGroup, err

}

// UpdateProfileGroup updates the ProfileGroup from the database with the given groupID and returns it (if it exists)
func UpdateProfileGroup(groupID string, newProfileGroup entities.ProfileGroup) (entities.ProfileGroup, error) {
	profileGroup := entities.ProfileGroup{}
	_, err := RemoveProfileGroup(groupID)
	if err != nil {
		return profileGroup, err
	}
	err = CreateProfileGroup(newProfileGroup)
	if err != nil {
		return profileGroup, err
	}

	return queries.GetProfileGroup(groupID)
}

// CreateProfile adds the Profile object to the database
func CreateProfile(profile entities.Profile) error {
	database := common.GetDatabase()
	if database == nil {
		return errors.New("database not initialized")
	}

	statement, err := database.Preparex(`INSERT INTO profiles (ID, profileGroupID, name, email, phoneNumber) VALUES (?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}

	_, err = statement.Exec(profile.ID, profile.ProfileGroupID, profile.Name, profile.Email, profile.PhoneNumber)
	if err != nil {
		return err
	}

	return CreateProfileInfos(profile)
}

// RemoveProfile removes the Profile from the database with the given ID and returns it (if it exists)
func RemoveProfile(ID string) (entities.Profile, error) {
	profile := entities.Profile{}
	database := common.GetDatabase()
	if database == nil {
		return profile, errors.New("database not initialized")
	}

	profile, err := queries.GetProfile(ID)
	if err != nil {
		return profile, err
	}

	statement, err := database.Preparex(`DELETE FROM profiles WHERE ID = @p1`)
	if err != nil {
		return profile, err
	}
	_, err = statement.Exec(ID)
	if err != nil {
		return profile, err
	}

	err = DeleteProfileInfos(ID)
	return profile, err

}

// UpdateProfile updates the Profile from the database with the given ID and returns it (if it exists)
func UpdateProfile(ID string, newProfile entities.Profile) (entities.Profile, error) {
	profile := entities.Profile{}
	_, err := RemoveProfile(ID)
	if err != nil {
		return profile, err
	}

	err = CreateProfile(newProfile)
	if err != nil {
		return profile, err
	}

	return queries.GetProfile(ID)
}

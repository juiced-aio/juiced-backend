package queries

import (
	"errors"
	"sort"
	"strings"

	"backend.juicedbot.io/juiced.infrastructure/common"
	"backend.juicedbot.io/juiced.infrastructure/common/entities"
)

// GetAllProfileGroups returns all ProfileGroup objects from the database
func GetAllProfileGroups() ([]entities.ProfileGroup, error) {
	profileGroups := []entities.ProfileGroup{}
	database := common.GetDatabase()
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
	}
	sort.SliceStable(profileGroups, func(i, j int) bool {
		return profileGroups[i].CreationDate > profileGroups[j].CreationDate
	})
	return profileGroups, err
}

// GetProfileGroup returns the ProfileGroup object from the database with the given groupID (if it exists)
func GetProfileGroup(groupID string) (entities.ProfileGroup, error) {
	profileGroup := entities.ProfileGroup{}
	database := common.GetDatabase()
	if database == nil {
		return profileGroup, errors.New("database not initialized")
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

// GetAllProfiles returns all Profile objects from the database
func GetAllProfiles() ([]entities.Profile, error) {
	profiles := []entities.Profile{}
	database := common.GetDatabase()
	if database == nil {
		return profiles, errors.New("database not initialized")
	}

	rows, err := database.Queryx("SELECT * FROM profiles")
	if err != nil {
		return profiles, err
	}

	defer rows.Close()
	for rows.Next() {
		tempProfile := entities.Profile{}
		err = rows.StructScan(&tempProfile)
		if err != nil {
			return profiles, err
		}
		tempProfile, err = GetProfileInfo(tempProfile)
		profiles = append(profiles, tempProfile)
	}

	sort.SliceStable(profiles, func(i, j int) bool {
		return profiles[i].CreationDate > profiles[j].CreationDate
	})

	return profiles, err
}

// GetProfile returns the Profile object from the database with the given ID (if it exists)
func GetProfile(ID string) (entities.Profile, error) {
	profile := entities.Profile{}
	database := common.GetDatabase()
	if database == nil {
		return profile, errors.New("database not initialized")
	}

	statement, err := database.Preparex("SELECT * FROM profiles WHERE ID = @p1")
	if err != nil {
		return profile, err
	}

	rows, err := statement.Queryx(ID)
	if err != nil {
		return profile, err
	}

	defer rows.Close()
	for rows.Next() {
		err = rows.StructScan(&profile)
		if err != nil {
			return profile, err
		}
	}

	return GetProfileInfo(profile)
}

// ConvertProfileIDsToProfiles returns a ProfileGroupWithProfiles object from a ProfileGroup object
func ConvertProfileIDsToProfiles(profileGroup *entities.ProfileGroup) (entities.ProfileGroupWithProfiles, error) {
	profileGroupWithProfiles := entities.ProfileGroupWithProfiles{GroupID: profileGroup.GroupID, Name: profileGroup.Name, Profiles: []entities.Profile{}}
	profiles := []entities.Profile{}
	for i := 0; i < len(profileGroup.ProfileIDs); i++ {
		profile, err := GetProfile(profileGroup.ProfileIDs[i])
		if err != nil {
			if profile.ID == "" {
				return profileGroupWithProfiles, err
			}
		} else {
			profiles = append(profiles, profile)
		}
	}
	profileGroupWithProfiles.SetProfiles(profiles)
	return profileGroupWithProfiles, nil
}

// ConvertProfilesToProfileIDs returns a ProfileGroup object from a ProfileGroupWithProfiles object
func ConvertProfilesToProfileIDs(profileGroupWithProfiles *entities.ProfileGroupWithProfiles) (entities.ProfileGroup, error) {
	profileGroup := entities.ProfileGroup{GroupID: profileGroupWithProfiles.GroupID, Name: profileGroupWithProfiles.Name, ProfileIDs: []string{}}
	profileIDs := []string{}
	for i := 0; i < len(profileGroupWithProfiles.Profiles); i++ {
		profileID := profileGroupWithProfiles.Profiles[i].ID
		profileIDs = append(profileIDs, profileID)
	}
	profileGroup.SetProfileIDs(profileIDs)
	return profileGroup, nil
}

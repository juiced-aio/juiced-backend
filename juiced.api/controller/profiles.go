package controller

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"backend.juicedbot.io/juiced.api/requests"
	"backend.juicedbot.io/juiced.api/responses"
	"backend.juicedbot.io/juiced.infrastructure/entities"
	"backend.juicedbot.io/juiced.infrastructure/stores"
	"backend.juicedbot.io/juiced.sitescripts/util"
	"github.com/gofiber/fiber/v2"
)

func GetProfiles(c *fiber.Ctx) error {
	return c.Status(200).JSON(stores.GetAllProfiles())
}

func CreateProfile(c *fiber.Ctx) error {
	var profile entities.Profile

	if err := c.BodyParser(&profile); err != nil {
		return responses.ReturnResponse(c, responses.CreateProfileParseErrorResponse, err)
	}

	cardType := util.DetectCardType([]byte(profile.CreditCard.CardNumber))
	if cardType == "" {
		return responses.ReturnResponse(c, responses.CreateProfileUnsupportedPaymentMethodErrorResponse, nil)
	}

	profilePtr, err := stores.CreateProfile(profile)
	if err != nil {
		return responses.ReturnResponse(c, responses.CreateProfileCreateErrorResponse, err)
	}

	return c.Status(200).JSON(*profilePtr)
}

func UpdateProfile(c *fiber.Ctx) error {
	profileID := c.Params("id")
	if profileID == "" {
		return responses.ReturnResponse(c, responses.UpdateProfileEmptyIDErrorResponse, nil)
	}

	var newProfile entities.Profile

	if err := c.BodyParser(&newProfile); err != nil {
		return responses.ReturnResponse(c, responses.UpdateProfileParseErrorResponse, err)
	}

	_, err := stores.GetProfile(profileID)
	if err != nil {
		return responses.ReturnResponse(c, responses.UpdateProfileNotFoundErrorResponse, err)
	}

	profilePtr, err := stores.UpdateProfile(profileID, newProfile)
	if err != nil {
		return responses.ReturnResponse(c, responses.UpdateProfileUpdateErrorResponse, err)
	}

	return c.Status(200).JSON(*profilePtr)
}

func DeleteProfiles(c *fiber.Ctx) error {
	var input requests.ProfilesRequest
	var err error

	if err = c.BodyParser(&input); err != nil {
		return responses.ReturnResponse(c, responses.DeleteProfilesParseErrorResponse, err)
	}

	if len(input.ProfileIDs) == 0 {
		return responses.ReturnResponse(c, responses.DeleteProfilesEmptyInputErrorResponse, nil)
	}

	response := responses.ProfilesSuccessResponse{}
	for _, profileID := range input.ProfileIDs {
		_, err_ := stores.RemoveProfile(profileID)
		if err_ == nil {
			response.SuccessProfileIDs = append(response.SuccessProfileIDs, profileID)
		} else {
			if err == nil {
				err = err_
			}
			response.FailureProfileIDs = append(response.FailureProfileIDs, profileID)
		}
	}

	if len(response.SuccessProfileIDs) == 0 {
		return responses.ReturnResponse(c, responses.DeleteProfilesDeleteErrorResponse, err)
	}

	return c.Status(200).JSON(response)
}

func CloneProfiles(c *fiber.Ctx) error {
	var input requests.ProfilesRequest
	var err error

	if err = c.BodyParser(&input); err != nil {
		return responses.ReturnResponse(c, responses.DeleteProfilesParseErrorResponse, err)
	}

	if len(input.ProfileIDs) == 0 {
		return responses.ReturnResponse(c, responses.DeleteProfilesEmptyInputErrorResponse, nil)
	}

	profiles := []entities.Profile{}
	success := false
	for _, profileID := range input.ProfileIDs {
		newProfilePtr, err_ := stores.CloneProfile(profileID)
		if err_ == nil {
			success = true
			profiles = append(profiles, *newProfilePtr)
		} else {
			if err == nil {
				err = err_
			}
		}
	}

	if !success {
		return responses.ReturnResponse(c, responses.DeleteProfilesDeleteErrorResponse, err)
	}

	return c.Status(200).JSON(profiles)
}

func ImportProfiles(c *fiber.Ctx) error {
	var input requests.ImportProfilesRequest
	var err error

	if err = c.BodyParser(&input); err != nil {
		return responses.ReturnResponse(c, responses.ImportProfilesParseErrorResponse, err)
	}

	if input.FilePath == "" {
		return responses.ReturnResponse(c, responses.ImportProfilesEmptyInputErrorResponse, err)
	}

	file, err := os.Open(input.FilePath)
	if err != nil {
		return responses.ReturnResponse(c, responses.ImportProfilesImportErrorResponse, err)
	}
	defer file.Close()
	byteValue, err := ioutil.ReadAll(file)
	if err != nil {
		return responses.ReturnResponse(c, responses.ImportProfilesImportErrorResponse, err)
	}
	parsedProfiles := requests.ImportProfilesFileFormat{}
	if err = json.Unmarshal(byteValue, &parsedProfiles); err != nil {
		return responses.ReturnResponse(c, responses.ImportProfilesImportErrorResponse, err)
	}

	if len(parsedProfiles.Profiles) == 0 {
		return responses.ReturnResponse(c, responses.ImportProfilesEmptyProfilesErrorResponse, err)
	}

	skippedGroups := 0
	validGroupIDs := []string{}
	for _, groupID := range input.GroupIDs {
		group, err := stores.GetProfileGroup(groupID)
		if err == nil && group.GroupID != "" {
			validGroupIDs = append(validGroupIDs, groupID)
		} else {
			skippedGroups++
		}
	}

	skippedProfiles := 0
	newProfiles := []entities.Profile{}
	newProfileIDs := []string{}
	for _, parsedProfile := range parsedProfiles.Profiles {
		existingProfile, err := stores.GetProfileByName(parsedProfile.Name)
		if err != nil || existingProfile.ID == "" || existingProfile.Name == "" {
			cardType := util.DetectCardType([]byte(parsedProfile.CreditCard.CardNumber))
			if cardType != "" {
				parsedProfile.CreditCard.CardType = cardType

				newProfilePtr, err := stores.CreateProfile(parsedProfile)
				if err == nil {
					parsedProfile = *newProfilePtr
					newProfiles = append(newProfiles, *newProfilePtr)
					newProfileIDs = append(newProfileIDs, newProfilePtr.ID)
				} else {
					skippedProfiles++
				}
			} else {
				skippedProfiles++
			}
		} else {
			skippedProfiles++
		}
	}

	newProfileGroups := []entities.ProfileGroup{}
	for _, groupID := range validGroupIDs {
		if newProfileGroupPtr, err := stores.AddProfilesToGroup(groupID, newProfileIDs); err != nil {
			skippedGroups++
		} else {
			newProfileGroups = append(newProfileGroups, *newProfileGroupPtr)
		}
	}

	importProfilesResponse := responses.ImportProfilesSuccessResponse{
		NewProfiles:      newProfiles,
		SkippedProfiles:  skippedProfiles,
		NewProfileGroups: newProfileGroups,
		SkippedGroups:    skippedGroups,
	}

	return c.Status(200).JSON(importProfilesResponse)
}

func ExportProfiles(c *fiber.Ctx) error {
	var input requests.ExportProfilesRequest
	var err error

	if err = c.BodyParser(&input); err != nil {
		return responses.ReturnResponse(c, responses.ExportProfilesParseErrorResponse, err)
	}

	if input.FilePath == "" || len(input.ProfileIDs) == 0 {
		return responses.ReturnResponse(c, responses.ExportProfilesEmptyInputErrorResponse, err)
	}

	if !strings.Contains(input.FilePath, ".json") ||
		len(strings.Split(input.FilePath, ".json")) == 1 ||
		len(strings.Split(input.FilePath, ".json")[1]) > 0 {
		return responses.ReturnResponse(c, responses.ExportProfilesFilePathNotJSONErrorResponse, nil)
	}

	profilesToExport := []requests.ExportProfile{}
	for _, profileID := range input.ProfileIDs {
		if profilePtr, err := stores.GetProfile(profileID); err == nil {
			exportProfile := requests.ExportProfile{}
			exportProfile.Name = profilePtr.Name
			exportProfile.Email = profilePtr.Email
			exportProfile.PhoneNumber = profilePtr.PhoneNumber
			exportProfile.ShippingAddress.FirstName = profilePtr.ShippingAddress.FirstName
			exportProfile.ShippingAddress.LastName = profilePtr.ShippingAddress.LastName
			exportProfile.ShippingAddress.Address1 = profilePtr.ShippingAddress.Address1
			exportProfile.ShippingAddress.Address2 = profilePtr.ShippingAddress.Address2
			exportProfile.ShippingAddress.City = profilePtr.ShippingAddress.City
			exportProfile.ShippingAddress.ZipCode = profilePtr.ShippingAddress.ZipCode
			exportProfile.ShippingAddress.StateCode = profilePtr.ShippingAddress.StateCode
			exportProfile.ShippingAddress.CountryCode = profilePtr.ShippingAddress.CountryCode
			exportProfile.BillingAddress.FirstName = profilePtr.BillingAddress.FirstName
			exportProfile.BillingAddress.LastName = profilePtr.BillingAddress.LastName
			exportProfile.BillingAddress.Address1 = profilePtr.BillingAddress.Address1
			exportProfile.BillingAddress.Address2 = profilePtr.BillingAddress.Address2
			exportProfile.BillingAddress.City = profilePtr.BillingAddress.City
			exportProfile.BillingAddress.ZipCode = profilePtr.BillingAddress.ZipCode
			exportProfile.BillingAddress.StateCode = profilePtr.BillingAddress.StateCode
			exportProfile.BillingAddress.CountryCode = profilePtr.BillingAddress.CountryCode
			exportProfile.CreditCard.CardholderName = profilePtr.CreditCard.CardholderName
			exportProfile.CreditCard.CardNumber = profilePtr.CreditCard.CardNumber
			exportProfile.CreditCard.ExpMonth = profilePtr.CreditCard.ExpMonth
			exportProfile.CreditCard.ExpYear = profilePtr.CreditCard.ExpYear
			exportProfile.CreditCard.CVV = profilePtr.CreditCard.CVV
			exportProfile.CreditCard.CardType = profilePtr.CreditCard.CardType
			profilesToExport = append(profilesToExport, exportProfile)
		}
	}

	if len(profilesToExport) == 0 {
		return responses.ReturnResponse(c, responses.ExportProfilesEmptyProfilesErrorResponse, nil)
	}

	u, err := user.Current()
	if err != nil {
		return responses.ReturnResponse(c, responses.ExportProfilesExportErrorResponse, err)
	}
	path := filepath.Join(u.HomeDir, "Desktop", input.FilePath)

	if _, err := os.Stat(path); !(os.IsNotExist(err)) {
		return responses.ReturnResponse(c, responses.ExportProfilesFileAlreadyExistsErrorResponse, nil)
	}

	profiles := requests.ExportProfilesFileFormat{
		Profiles: profilesToExport,
	}

	data, err := json.MarshalIndent(profiles, "", "\t")

	if err != nil {
		return responses.ReturnResponse(c, responses.ExportProfilesExportErrorResponse, err)
	}

	if err := ioutil.WriteFile(path, data, 0644); err != nil {
		return responses.ReturnResponse(c, responses.ExportProfilesExportErrorResponse, err)
	}

	return c.SendStatus(200)
}

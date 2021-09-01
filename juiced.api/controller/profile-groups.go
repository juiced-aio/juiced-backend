package controller

import (
	"backend.juicedbot.io/juiced.api/requests"
	"backend.juicedbot.io/juiced.api/responses"
	"backend.juicedbot.io/juiced.infrastructure/entities"
	"backend.juicedbot.io/juiced.infrastructure/stores"
	"github.com/gofiber/fiber/v2"
)

func GetProfileGroups(c *fiber.Ctx) error {
	return c.Status(200).JSON(stores.GetAllProfileGroups())
}

func CreateProfileGroup(c *fiber.Ctx) error {
	var profileGroup entities.ProfileGroup

	if err := c.BodyParser(&profileGroup); err != nil {
		return responses.ReturnResponse(c, responses.CreateProfileGroupParseErrorResponse, err)
	}

	profileGroupPtr, err := stores.CreateProfileGroup(profileGroup)
	if err != nil {
		return responses.ReturnResponse(c, responses.CreateProfileGroupCreateErrorResponse, err)
	}

	return c.Status(200).JSON(*profileGroupPtr)
}

func UpdateProfileGroup(c *fiber.Ctx) error {
	profileGroupID := c.Params("id")
	if profileGroupID == "" {
		return responses.ReturnResponse(c, responses.UpdateProfileGroupEmptyIDErrorResponse, nil)
	}

	var newProfileGroup entities.ProfileGroup

	if err := c.BodyParser(&newProfileGroup); err != nil {
		return responses.ReturnResponse(c, responses.UpdateProfileGroupParseErrorResponse, err)
	}

	_, err := stores.GetProfileGroup(profileGroupID)
	if err != nil {
		return responses.ReturnResponse(c, responses.UpdateProfileGroupNotFoundErrorResponse, err)
	}

	profileGroupPtr, err := stores.UpdateProfileGroup(profileGroupID, newProfileGroup)
	if err != nil {
		return responses.ReturnResponse(c, responses.UpdateProfileGroupUpdateErrorResponse, err)
	}

	return c.Status(200).JSON(*profileGroupPtr)
}

func DeleteProfileGroups(c *fiber.Ctx) error {
	var input requests.ProfileGroupsRequest
	var err error

	if err = c.BodyParser(&input); err != nil {
		return responses.ReturnResponse(c, responses.DeleteProfileGroupsParseErrorResponse, err)
	}

	if len(input.ProfileGroupIDs) == 0 {
		return responses.ReturnResponse(c, responses.DeleteProfileGroupsEmptyInputErrorResponse, nil)
	}

	response := responses.ProfileGroupsSuccessResponse{}
	for _, profileGroupID := range input.ProfileGroupIDs {
		_, err_ := stores.RemoveProfileGroup(profileGroupID)
		if err_ == nil {
			response.SuccessProfileGroupIDs = append(response.SuccessProfileGroupIDs, profileGroupID)
		} else {
			if err == nil {
				err = err_
			}
			response.FailureProfileGroupIDs = append(response.FailureProfileGroupIDs, profileGroupID)
		}
	}

	if len(response.SuccessProfileGroupIDs) == 0 {
		return responses.ReturnResponse(c, responses.DeleteProfileGroupsDeleteErrorResponse, err)
	}

	return c.Status(200).JSON(response)
}

func CloneProfileGroups(c *fiber.Ctx) error {
	var input requests.ProfileGroupsRequest
	var err error

	if err = c.BodyParser(&input); err != nil {
		return responses.ReturnResponse(c, responses.DeleteProfileGroupsParseErrorResponse, err)
	}

	if len(input.ProfileGroupIDs) == 0 {
		return responses.ReturnResponse(c, responses.DeleteProfileGroupsEmptyInputErrorResponse, nil)
	}

	profileGroups := []entities.ProfileGroup{}
	success := false
	for _, profileGroupID := range input.ProfileGroupIDs {
		newProfileGroupPtr, err_ := stores.CloneProfileGroup(profileGroupID)
		if err_ == nil {
			success = true
			profileGroups = append(profileGroups, *newProfileGroupPtr)
		} else {
			if err == nil {
				err = err_
			}
		}
	}

	if !success {
		return responses.ReturnResponse(c, responses.DeleteProfileGroupsDeleteErrorResponse, err)
	}

	return c.Status(200).JSON(profileGroups)
}

func AddProfilesToGroups(c *fiber.Ctx) error {
	var input requests.ProfileGroupsRequest
	var err error

	if err = c.BodyParser(&input); err != nil {
		return responses.ReturnResponse(c, responses.AddProfilesToGroupsParseErrorResponse, err)
	}

	if len(input.ProfileGroupIDs) == 0 || len(input.ProfileIDs) == 0 {
		return responses.ReturnResponse(c, responses.AddProfilesToGroupsEmptyInputErrorResponse, nil)
	}

	profileGroups := []entities.ProfileGroup{}
	success := false
	for _, profileGroupID := range input.ProfileGroupIDs {
		profileGroupPtr, err_ := stores.AddProfilesToGroup(profileGroupID, input.ProfileIDs)
		if err_ == nil {
			success = true
			profileGroups = append(profileGroups, *profileGroupPtr)
		} else {
			if err == nil {
				err = err_
			}
		}
	}

	if !success {
		return responses.ReturnResponse(c, responses.DeleteProfileGroupsDeleteErrorResponse, err)
	}

	return c.Status(200).JSON(profileGroups)
}

func RemoveProfilesFromGroups(c *fiber.Ctx) error {
	var input requests.ProfileGroupsRequest
	var err error

	if err = c.BodyParser(&input); err != nil {
		return responses.ReturnResponse(c, responses.RemoveProfilesFromGroupsParseErrorResponse, err)
	}

	if len(input.ProfileGroupIDs) == 0 || len(input.ProfileIDs) == 0 {
		return responses.ReturnResponse(c, responses.RemoveProfilesFromGroupsEmptyInputErrorResponse, nil)
	}

	profileGroups := []entities.ProfileGroup{}
	success := false
	for _, profileGroupID := range input.ProfileGroupIDs {
		profileGroupPtr, err_ := stores.RemoveProfilesFromGroup(profileGroupID, input.ProfileIDs)
		if err_ == nil {
			success = true
			profileGroups = append(profileGroups, *profileGroupPtr)
		} else {
			if err == nil {
				err = err_
			}
		}
	}

	if !success {
		return responses.ReturnResponse(c, responses.DeleteProfileGroupsDeleteErrorResponse, err)
	}

	return c.Status(200).JSON(profileGroups)
}

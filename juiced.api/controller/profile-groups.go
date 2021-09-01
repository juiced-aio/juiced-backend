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

	response := responses.ProfileGroupsSuccessResponse{}
	for _, profileGroupID := range input.ProfileGroupIDs {
		newProfileGroup, err_ := stores.CloneProfileGroup(profileGroupID)
		if err_ == nil {
			response.SuccessProfileGroupIDs = append(response.SuccessProfileGroupIDs, newProfileGroup.GroupID)
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

func AddProfilesToGroups(c *fiber.Ctx) error {

}

func RemoveProfilesFromGroups(c *fiber.Ctx) error {

}

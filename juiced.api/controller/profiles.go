package controller

import (
	"backend.juicedbot.io/juiced.api/requests"
	"backend.juicedbot.io/juiced.api/responses"
	"backend.juicedbot.io/juiced.infrastructure/entities"
	"backend.juicedbot.io/juiced.infrastructure/stores"
	"github.com/gofiber/fiber/v2"
)

func CreateProfile(c *fiber.Ctx) error {
	var profile entities.Profile

	if err := c.BodyParser(&profile); err != nil {
		return responses.ReturnResponse(c, responses.CreateProfileParseErrorResponse, err)
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

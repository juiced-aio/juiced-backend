package controller

import (
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

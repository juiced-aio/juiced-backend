package controller

import (
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

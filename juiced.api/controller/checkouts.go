package controller

import (
	"strconv"

	"backend.juicedbot.io/juiced.api/responses"
	"backend.juicedbot.io/juiced.infrastructure/staticstores"
	"github.com/gofiber/fiber/v2"
)

func GetCheckouts(c *fiber.Ctx) error {
	retailer := c.Query("retailer")
	daysStr := c.Query("days")

	if daysStr == "" {
		daysStr = "-1"
	}

	days, err := strconv.Atoi(daysStr)
	if err != nil {
		return responses.ReturnResponse(c, responses.GetCheckoutsParseErrorResponse, err)
	}

	if days != -1 && days <= 0 {
		return responses.ReturnResponse(c, responses.GetCheckoutsNegativeDaysErrorResponse, nil)
	}

	return c.Status(200).JSON(staticstores.GetCheckouts(retailer, days))
}

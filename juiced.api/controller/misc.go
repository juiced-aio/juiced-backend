package controller

import (
	"backend.juicedbot.io/juiced.api/requests"
	"backend.juicedbot.io/juiced.api/responses"
	rpc "backend.juicedbot.io/juiced.rpc"

	"github.com/gofiber/fiber/v2"
)

func SetVersion(c *fiber.Ctx) error {
	var input requests.SetVersionRequest

	if err := c.BodyParser(&input); err != nil {
		return responses.ReturnResponse(c, responses.SetVersionParseErrorResponse, err)
	}

	if input.Version == "" {
		return responses.ReturnResponse(c, responses.SetVersionEmptyInputErrorResponse, nil)
	}

	if err := rpc.SetActivity(input.Version); err != nil {
		return responses.ReturnResponse(c, responses.SetVersionStartRPCWarningResponse, nil)
	}

	return responses.ReturnSuccessResponse(c)
}

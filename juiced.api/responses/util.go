package responses

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
)

type Response struct {
	StatusCode int    `json:"-"`
	ErrorCode  int    `json:"errorCode"`
	Message    string `json:"message"`
}

func ReturnSuccessResponse(c *fiber.Ctx) error {
	return c.SendStatus(200)
}

func ReturnErrorResponse(c *fiber.Ctx, info Response, err error) error {
	message := info.Message
	if err != nil {
		message = fmt.Sprintf("%s: %s", info.Message, err.Error())
	}
	return c.Status(info.StatusCode).JSON(Response{
		ErrorCode: info.ErrorCode,
		Message:   message,
	})
}

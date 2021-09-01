package controller

import (
	"backend.juicedbot.io/juiced.api/responses"
	"backend.juicedbot.io/juiced.infrastructure/entities"
	"backend.juicedbot.io/juiced.infrastructure/stores"
	"github.com/gofiber/fiber/v2"
)

func GetTaskGroups(c *fiber.Ctx) error {
	return c.Status(200).JSON(stores.GetAllTaskGroups())
}

func CreateTaskGroup(c *fiber.Ctx) error {
	var taskGroup entities.TaskGroup

	if err := c.BodyParser(&taskGroup); err != nil {
		return responses.ReturnResponse(c, responses.CreateTaskGroupParseErrorResponse, err)
	}

	taskGroupPtr, err := stores.CreateTaskGroup(taskGroup)
	if err != nil {
		return responses.ReturnResponse(c, responses.CreateTaskGroupCreateErrorResponse, err)
	}

	return c.Status(200).JSON(*taskGroupPtr)
}

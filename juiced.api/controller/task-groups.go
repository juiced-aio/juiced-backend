package controller

import (
	"backend.juicedbot.io/juiced.api/requests"
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

func DeleteTaskGroups(c *fiber.Ctx) error {
	var input requests.TaskGroupsRequest
	var err error

	if err = c.BodyParser(&input); err != nil {
		return responses.ReturnResponse(c, responses.DeleteTaskGroupsParseErrorResponse, err)
	}

	if len(input.TaskGroupIDs) == 0 {
		return responses.ReturnResponse(c, responses.DeleteTaskGroupsEmptyInputErrorResponse, nil)
	}

	response := responses.TaskGroupsSuccessResponse{}
	for _, taskGroupID := range input.TaskGroupIDs {
		_, err_ := stores.RemoveTaskGroup(taskGroupID)
		if err_ == nil {
			response.SuccessTaskGroupIDs = append(response.SuccessTaskGroupIDs, taskGroupID)
		} else {
			if err == nil {
				err = err_
			}
			response.FailureTaskGroupIDs = append(response.FailureTaskGroupIDs, taskGroupID)
		}
	}

	if len(response.SuccessTaskGroupIDs) == 0 {
		return responses.ReturnResponse(c, responses.DeleteTaskGroupsDeleteErrorResponse, err)
	}

	return c.Status(200).JSON(response)
}

func CloneTaskGroups(c *fiber.Ctx) error {
	var input requests.TaskGroupsRequest
	var err error

	if err = c.BodyParser(&input); err != nil {
		return responses.ReturnResponse(c, responses.DeleteTaskGroupsParseErrorResponse, err)
	}

	if len(input.TaskGroupIDs) == 0 {
		return responses.ReturnResponse(c, responses.DeleteTaskGroupsEmptyInputErrorResponse, nil)
	}

	taskGroups := []entities.TaskGroup{}
	success := false
	for _, taskGroupID := range input.TaskGroupIDs {
		newTaskGroupPtr, err_ := stores.CloneTaskGroup(taskGroupID)
		if err_ == nil {
			success = true
			taskGroups = append(taskGroups, *newTaskGroupPtr)
		} else {
			if err == nil {
				err = err_
			}
		}
	}

	if !success {
		return responses.ReturnResponse(c, responses.DeleteTaskGroupsDeleteErrorResponse, err)
	}

	return c.Status(200).JSON(taskGroups)
}

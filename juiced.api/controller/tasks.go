package controller

import (
	"backend.juicedbot.io/juiced.api/requests"
	"backend.juicedbot.io/juiced.api/responses"
	"backend.juicedbot.io/juiced.infrastructure/entities"
	"backend.juicedbot.io/juiced.infrastructure/stores"
	"github.com/gofiber/fiber/v2"
)

func CreateTasks(c *fiber.Ctx) error {
	var input requests.CreateTasksRequest
	var err error

	if err := c.BodyParser(&input); err != nil {
		return responses.ReturnResponse(c, responses.CreateTaskParseErrorResponse, err)
	}

	if input.TaskGroupID == "" ||
		input.NumTasksPerProfile == 0 ||
		(input.ProfileGroupID == "" && len(input.ProfileIDs) == 0) ||
		input.Retailer == "" {
		return responses.ReturnResponse(c, responses.CreateTaskEmptyInputErrorResponse, nil)
	}

	profileIDs := input.ProfileIDs
	if input.ProfileGroupID != "" {
		if profileGroupPtr, err := stores.GetProfileGroup(input.ProfileGroupID); err != nil {
			return responses.ReturnResponse(c, responses.CreateTaskInvalidProfileGroupErrorResponse, err)
		} else {
			profileIDs = profileGroupPtr.ProfileIDs
		}
	}

	newTasks := []entities.Task{}
	newTaskPtrs := []*entities.Task{}
	for _, profileID := range profileIDs {
		for i := 0; i < input.NumTasksPerProfile; i++ {
			newTask := entities.Task{
				TaskGroupID: input.TaskGroupID,
				Retailer:    input.Retailer,
				Task: &entities.BaseTask{
					TaskInput: entities.TaskInput{
						ProxyGroupID:     input.ProxyGroupID,
						Quantity:         input.Quantity,
						ProfileID:        profileID,
						DelayMS:          input.DelayMS,
						SiteSpecificInfo: input.SiteSpecificInfo,
					},
				},
			}
			var taskPtr *entities.Task
			taskPtr, err = stores.CreateTask(newTask)
			if err != nil {
				break
			}
			newTaskPtrs = append(newTaskPtrs, taskPtr)
			newTasks = append(newTasks, *taskPtr)
		}
	}

	if _, err := stores.AddTasksToGroup(input.TaskGroupID, newTaskPtrs); err != nil {
		return responses.ReturnResponse(c, responses.CreateTaskGroupCreateErrorResponse, err)
	}

	if err != nil {
		return responses.ReturnResponse(c, responses.CreateTaskGroupCreateErrorResponse, err)
	}

	return c.Status(200).JSON(newTasks)
}

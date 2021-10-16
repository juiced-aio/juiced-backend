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
		return responses.ReturnResponse(c, responses.CreateTaskCreateErrorResponse, err)
	}

	if err != nil {
		return responses.ReturnResponse(c, responses.CreateTaskCreateErrorResponse, err)
	}

	return c.Status(200).JSON(newTasks)
}

func DeleteTasks(c *fiber.Ctx) error {
	var input requests.TasksRequest
	var err error

	if err = c.BodyParser(&input); err != nil {
		return responses.ReturnResponse(c, responses.DeleteTasksParseErrorResponse, err)
	}

	if len(input.TaskIDs) == 0 {
		return responses.ReturnResponse(c, responses.DeleteTasksEmptyInputErrorResponse, nil)
	}

	response := responses.TasksSuccessResponse{}
	for _, taskID := range input.TaskIDs {
		_, err_ := stores.RemoveTask(taskID)
		if err_ == nil {
			response.SuccessTaskIDs = append(response.SuccessTaskIDs, taskID)
		} else {
			if err == nil {
				err = err_
			}
			response.FailureTaskIDs = append(response.FailureTaskIDs, taskID)
		}
	}

	if len(response.SuccessTaskIDs) == 0 {
		return responses.ReturnResponse(c, responses.DeleteTasksDeleteErrorResponse, err)
	}

	return c.Status(200).JSON(response)
}

func CloneTasks(c *fiber.Ctx) error {
	var input requests.TasksRequest
	var err error

	if err = c.BodyParser(&input); err != nil {
		return responses.ReturnResponse(c, responses.CloneTasksParseErrorResponse, err)
	}

	if len(input.TaskIDs) == 0 {
		return responses.ReturnResponse(c, responses.CloneTasksEmptyInputErrorResponse, nil)
	}

	tasks := []entities.Task{}
	success := false
	for _, taskID := range input.TaskIDs {
		newTaskPtr, err_ := stores.CloneTask(taskID, "")
		if err_ == nil {
			success = true
			tasks = append(tasks, *newTaskPtr)
		} else {
			if err == nil {
				err = err_
			}
		}
	}

	if !success {
		return responses.ReturnResponse(c, responses.CloneTasksCloneErrorResponse, err)
	}

	return c.Status(200).JSON(tasks)
}

func StartTasks(c *fiber.Ctx) error {
	var input requests.TasksRequest
	var err error

	if err = c.BodyParser(&input); err != nil {
		return responses.ReturnResponse(c, responses.StartTasksParseErrorResponse, err)
	}

	if len(input.TaskIDs) == 0 {
		return responses.ReturnResponse(c, responses.StartTasksEmptyInputErrorResponse, nil)
	}

	response := responses.TasksSuccessResponse{}
	for _, taskID := range input.TaskIDs {
		err_ := stores.StartTask(taskID)
		if err_ == nil {
			response.SuccessTaskIDs = append(response.SuccessTaskIDs, taskID)
		} else {
			if err == nil {
				err = err_
			}
			response.FailureTaskIDs = append(response.FailureTaskIDs, taskID)
		}
	}

	if len(response.SuccessTaskIDs) == 0 {
		return responses.ReturnResponse(c, responses.StartTasksStartErrorResponse, err)
	}

	return c.Status(200).JSON(response)
}

func StopTasks(c *fiber.Ctx) error {
	var input requests.TasksRequest
	var err error

	if err = c.BodyParser(&input); err != nil {
		return responses.ReturnResponse(c, responses.StopTasksParseErrorResponse, err)
	}

	if len(input.TaskIDs) == 0 {
		return responses.ReturnResponse(c, responses.StopTasksEmptyInputErrorResponse, nil)
	}

	response := responses.TasksSuccessResponse{}
	for _, taskID := range input.TaskIDs {
		err_ := stores.StopTask(taskID)
		if err_ == nil {
			response.SuccessTaskIDs = append(response.SuccessTaskIDs, taskID)
		} else {
			if err == nil {
				err = err_
			}
			response.FailureTaskIDs = append(response.FailureTaskIDs, taskID)
		}
	}

	if len(response.SuccessTaskIDs) == 0 {
		return responses.ReturnResponse(c, responses.StopTasksStopErrorResponse, err)
	}

	return c.Status(200).JSON(response)
}

func UpdateTasks(c *fiber.Ctx) error {
	var input requests.UpdateTasksRequest
	var err error

	if err = c.BodyParser(&input); err != nil {
		return responses.ReturnResponse(c, responses.UpdateTasksParseErrorResponse, err)
	}

	if len(input.TaskIDs) == 0 {
		return responses.ReturnResponse(c, responses.UpdateTasksEmptyInputErrorResponse, nil)
	}

	response := responses.TasksResponse{}
	for _, taskID := range input.TaskIDs {
		newTask, err_ := stores.UpdateTask(taskID, input.TaskInput)
		if err_ == nil {
			response.Data = append(response.Data, *newTask)
		} else {
			if err == nil {
				err = err_
			}
		}
	}

	if len(response.Data) == 0 {
		return responses.ReturnResponse(c, responses.UpdateTasksUpdateErrorResponse, err)
	}

	return c.Status(200).JSON(response)
}

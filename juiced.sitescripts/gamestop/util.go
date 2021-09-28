package gamestop

import (
	"strings"

	"backend.juicedbot.io/juiced.infrastructure/enums"
)

const SKU_FORMAT = "XXXXXXXX"

func ValidateMonitorInput(input string, monitorType enums.MonitorType, info map[string]interface{}) (MonitorInput, error) {
	gameStopMonitorInput := MonitorInput{}
	switch monitorType {
	case enums.SKUMonitor:
		if len(input) != 8 {
			return gameStopMonitorInput, &enums.InvalidSKUError{Retailer: enums.GameStop, Format: SKU_FORMAT}
		}

	default:
		return gameStopMonitorInput, &enums.UnsupportedMonitorTypeError{Retailer: enums.GameStop, MonitorType: monitorType}
	}

	sizeInterfaces, ok := info["sizes"].([]interface{})
	if !ok {
		return gameStopMonitorInput, &enums.InvalidInputTypeError{Field: "sizes", ShouldBe: "[]string"}
	}
	sizes := []string{}
	for _, sizeInterface := range sizeInterfaces {
		if size, ok := sizeInterface.(string); !ok {
			return gameStopMonitorInput, &enums.InvalidInputTypeError{Field: "sizes", ShouldBe: "[]string"}
		} else {
			sizes = append(sizes, size)
		}
	}

	colorInterfaces, ok := info["colors"].([]interface{})
	colors := []string{}
	if !ok {
		return gameStopMonitorInput, &enums.InvalidInputTypeError{Field: "colors", ShouldBe: "[]string"}
	}
	for _, colorInterface := range colorInterfaces {
		if color, ok := colorInterface.(string); !ok {
			return gameStopMonitorInput, &enums.InvalidInputTypeError{Field: "colors", ShouldBe: "[]string"}
		} else {
			colors = append(colors, color)
		}
	}

	conditionInterfaces, ok := info["conditions"].([]interface{})
	conditions := []string{}
	if !ok {
		return gameStopMonitorInput, &enums.InvalidInputTypeError{Field: "conditions", ShouldBe: "[]string"}
	}
	for _, conditionInterface := range conditionInterfaces {
		if condition, ok := conditionInterface.(string); !ok {
			return gameStopMonitorInput, &enums.InvalidInputTypeError{Field: "conditions", ShouldBe: "[]string"}
		} else {
			conditions = append(conditions, condition)
		}
	}

	gameStopMonitorInput.Size = strings.Join(sizes, ",")
	gameStopMonitorInput.Color = strings.Join(colors, ",")
	gameStopMonitorInput.Condition = strings.Join(conditions, ",")

	return gameStopMonitorInput, nil
}

func ValidateTaskInput(info map[string]interface{}) (TaskInput, error) {
	gameStopTaskInput := TaskInput{}
	if info["taskType"] == enums.TaskTypeAccount {
		gameStopTaskInput.TaskType = enums.TaskTypeAccount
		if email, ok := info["email"].(string); !ok {
			return gameStopTaskInput, &enums.InvalidInputTypeError{Field: "email", ShouldBe: "string"}
		} else {
			if email == "" {
				return gameStopTaskInput, &enums.EmptyInputFieldError{Field: "email"}
			}
			gameStopTaskInput.Email = email
		}
		if password, ok := info["password"].(string); !ok {
			return gameStopTaskInput, &enums.InvalidInputTypeError{Field: "password", ShouldBe: "string"}
		} else {
			if password == "" {
				return gameStopTaskInput, &enums.EmptyInputFieldError{Field: "password"}
			}
			gameStopTaskInput.Password = password
		}
	} else {
		gameStopTaskInput.TaskType = enums.TaskTypeGuest
	}

	return gameStopTaskInput, nil
}

package disney

import (
	"strings"

	"backend.juicedbot.io/juiced.infrastructure/enums"
)

const SKU_FORMAT = "XXXXXXXXXXXX"

func ValidateMonitorInput(input string, monitorType enums.MonitorType, info map[string]interface{}) (MonitorInput, error) {
	disneyMonitorInput := MonitorInput{}
	switch monitorType {
	case enums.SKUMonitor:
		if len(input) < len(SKU_FORMAT) {
			return disneyMonitorInput, &enums.InvalidSKUError{Retailer: enums.Disney, Format: SKU_FORMAT}
		}

	default:
		return disneyMonitorInput, &enums.UnsupportedMonitorTypeError{Retailer: enums.Disney, MonitorType: monitorType}
	}

	sizeInterfaces, ok := info["sizes"].([]interface{})
	if !ok {
		return disneyMonitorInput, &enums.InvalidInputTypeError{Field: "sizes", ShouldBe: "[]string"}
	}
	sizes := []string{}
	for _, sizeInterface := range sizeInterfaces {
		if size, ok := sizeInterface.(string); !ok {
			return disneyMonitorInput, &enums.InvalidInputTypeError{Field: "sizes", ShouldBe: "[]string"}
		} else {
			sizes = append(sizes, size)
		}
	}

	colorInterfaces, ok := info["colors"].([]interface{})
	colors := []string{}
	if !ok {
		return disneyMonitorInput, &enums.InvalidInputTypeError{Field: "colors", ShouldBe: "[]string"}
	}
	for _, colorInterface := range colorInterfaces {
		if color, ok := colorInterface.(string); !ok {
			return disneyMonitorInput, &enums.InvalidInputTypeError{Field: "colors", ShouldBe: "[]string"}
		} else {
			colors = append(colors, color)
		}
	}

	disneyMonitorInput.Size = strings.Join(sizes, ",")
	disneyMonitorInput.Color = strings.Join(colors, ",")

	return disneyMonitorInput, nil
}

func ValidateTaskInput(info map[string]interface{}) (TaskInput, error) {
	disneyTaskInput := TaskInput{}
	if info["taskType"] == enums.TaskTypeAccount {
		disneyTaskInput.TaskType = enums.TaskTypeAccount
		if email, ok := info["email"].(string); !ok {
			return disneyTaskInput, &enums.InvalidInputTypeError{Field: "email", ShouldBe: "string"}
		} else {
			if email == "" {
				return disneyTaskInput, &enums.EmptyInputFieldError{Field: "email"}
			}
			disneyTaskInput.Email = email
		}
		if password, ok := info["password"].(string); !ok {
			return disneyTaskInput, &enums.InvalidInputTypeError{Field: "password", ShouldBe: "string"}
		} else {
			if password == "" {
				return disneyTaskInput, &enums.EmptyInputFieldError{Field: "password"}
			}
			disneyTaskInput.Password = password
		}
	} else {
		disneyTaskInput.TaskType = enums.TaskTypeGuest
	}

	return disneyTaskInput, nil
}

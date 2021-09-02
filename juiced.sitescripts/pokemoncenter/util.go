package pokemoncenter

import (
	"strings"

	"backend.juicedbot.io/juiced.infrastructure/enums"
)

const SKU_FORMAT = "XXX-XXXXX"

func ValidateMonitorInput(input string, monitorType enums.MonitorType, info map[string]interface{}) error {
	switch monitorType {
	case enums.SKUMonitor:
		if !strings.Contains(input, "-") {
			return &enums.InvalidSKUError{Retailer: enums.PokemonCenter, Format: SKU_FORMAT}
		}
		split := strings.Split(input, "-")
		if len(split) != 2 || len(split[0]) != 3 || len(split[1]) != 5 {
			return &enums.InvalidSKUError{Retailer: enums.PokemonCenter, Format: SKU_FORMAT}
		}

	default:
		return &enums.UnsupportedMonitorTypeError{Retailer: enums.PokemonCenter, MonitorType: monitorType}
	}

	return nil
}

func ValidateTaskInput(info map[string]interface{}) (TaskInput, error) {
	pokemonCenterTaskInput := TaskInput{}
	if info["taskType"] == enums.TaskTypeAccount {
		pokemonCenterTaskInput.TaskType = enums.TaskTypeAccount
		if email, ok := info["email"].(string); !ok {
			return pokemonCenterTaskInput, &enums.InvalidInputTypeError{Field: "email", ShouldBe: "string"}
		} else {
			if email == "" {
				return pokemonCenterTaskInput, &enums.EmptyInputFieldError{Field: "email"}
			}
			pokemonCenterTaskInput.Email = email
		}
		if password, ok := info["password"].(string); !ok {
			return pokemonCenterTaskInput, &enums.InvalidInputTypeError{Field: "password", ShouldBe: "string"}
		} else {
			if password == "" {
				return pokemonCenterTaskInput, &enums.EmptyInputFieldError{Field: "password"}
			}
			pokemonCenterTaskInput.Password = password
		}
	} else {
		pokemonCenterTaskInput.TaskType = enums.TaskTypeGuest
	}

	return pokemonCenterTaskInput, nil
}

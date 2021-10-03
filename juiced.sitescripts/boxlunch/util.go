package boxlunch

import (
	"strings"

	"backend.juicedbot.io/juiced.infrastructure/enums"
)

const SKU_FORMAT = "XXXXXXXX"

func ValidateMonitorInput(input string, monitorType enums.MonitorType, info map[string]interface{}) (MonitorInput, error) {
	boxlunchMonitorInput := MonitorInput{}
	switch monitorType {
	case enums.SKUMonitor:
		if len(input) != 8 {
			return boxlunchMonitorInput, &enums.InvalidSKUError{Retailer: enums.BoxLunch, Format: SKU_FORMAT}
		}

	default:
		return boxlunchMonitorInput, &enums.UnsupportedMonitorTypeError{Retailer: enums.BoxLunch, MonitorType: monitorType}
	}

	sizeInterfaces, ok := info["sizes"].([]interface{})
	if !ok {
		return boxlunchMonitorInput, &enums.InvalidInputTypeError{Field: "sizes", ShouldBe: "[]string"}
	}
	sizes := []string{}
	for _, sizeInterface := range sizeInterfaces {
		if size, ok := sizeInterface.(string); !ok {
			return boxlunchMonitorInput, &enums.InvalidInputTypeError{Field: "sizes", ShouldBe: "[]string"}
		} else {
			sizes = append(sizes, size)
		}
	}

	colorInterfaces, ok := info["colors"].([]interface{})
	colors := []string{}
	if !ok {
		return boxlunchMonitorInput, &enums.InvalidInputTypeError{Field: "colors", ShouldBe: "[]string"}
	}
	for _, colorInterface := range colorInterfaces {
		if color, ok := colorInterface.(string); !ok {
			return boxlunchMonitorInput, &enums.InvalidInputTypeError{Field: "colors", ShouldBe: "[]string"}
		} else {
			colors = append(colors, color)
		}
	}

	boxlunchMonitorInput.Size = strings.Join(sizes, ",")
	boxlunchMonitorInput.Color = strings.Join(colors, ",")

	return boxlunchMonitorInput, nil
}

func ValidateTaskInput(info map[string]interface{}) (TaskInput, error) {
	boxlunchTaskInput := TaskInput{}
	return boxlunchTaskInput, nil
}

package hottopic

import (
	"strings"

	"backend.juicedbot.io/juiced.infrastructure/enums"
)

const SKU_FORMAT = "XXXXXXXX"

func ValidateMonitorInput(input string, monitorType enums.MonitorType, info map[string]interface{}) (MonitorInput, error) {
	hotTopicMonitorInput := MonitorInput{}
	switch monitorType {
	case enums.SKUMonitor:
		if len(input) != len(SKU_FORMAT) {
			return hotTopicMonitorInput, &enums.InvalidSKUError{Retailer: enums.HotTopic, Format: SKU_FORMAT}
		}

	default:
		return hotTopicMonitorInput, &enums.UnsupportedMonitorTypeError{Retailer: enums.HotTopic, MonitorType: monitorType}
	}

	sizeInterfaces, ok := info["sizes"].([]interface{})
	if !ok {
		return hotTopicMonitorInput, &enums.InvalidInputTypeError{Field: "sizes", ShouldBe: "[]string"}
	}
	sizes := []string{}
	for _, sizeInterface := range sizeInterfaces {
		if size, ok := sizeInterface.(string); !ok {
			return hotTopicMonitorInput, &enums.InvalidInputTypeError{Field: "sizes", ShouldBe: "[]string"}
		} else {
			sizes = append(sizes, size)
		}
	}

	colorInterfaces, ok := info["colors"].([]interface{})
	colors := []string{}
	if !ok {
		return hotTopicMonitorInput, &enums.InvalidInputTypeError{Field: "colors", ShouldBe: "[]string"}
	}
	for _, colorInterface := range colorInterfaces {
		if color, ok := colorInterface.(string); !ok {
			return hotTopicMonitorInput, &enums.InvalidInputTypeError{Field: "colors", ShouldBe: "[]string"}
		} else {
			colors = append(colors, color)
		}
	}

	hotTopicMonitorInput.Size = strings.Join(sizes, ",")
	hotTopicMonitorInput.Color = strings.Join(colors, ",")

	return hotTopicMonitorInput, nil
}

func ValidateTaskInput(info map[string]interface{}) (TaskInput, error) {
	hotTopicTaskInput := TaskInput{}
	return hotTopicTaskInput, nil
}

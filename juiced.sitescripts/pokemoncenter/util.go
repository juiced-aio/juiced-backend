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

package base

import (
	"backend.juicedbot.io/juiced.infrastructure/common/enums"
	"backend.juicedbot.io/juiced.sitescripts/pokemoncenter"
)

type Monitor struct {
	Retailer enums.Retailer

	PokemonCenterMonitor *pokemoncenter.Monitor
}

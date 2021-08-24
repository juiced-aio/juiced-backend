package base

import (
	"backend.juicedbot.io/juiced.infrastructure/common/enums"

	"backend.juicedbot.io/juiced.sitescripts/pokemoncenter"
)

type Task struct {
	Retailer          enums.Retailer
	PokemonCenterTask *pokemoncenter.Task
}

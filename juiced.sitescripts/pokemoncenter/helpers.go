package pokemoncenter

import (
	"backend.juicedbot.io/juiced.antibot/datadome"
	"backend.juicedbot.io/juiced.client/http"
	"backend.juicedbot.io/juiced.infrastructure/entities"
	"backend.juicedbot.io/juiced.infrastructure/enums"
	"backend.juicedbot.io/juiced.sitescripts/util"
)

// AddPokemonCenterHeaders adds PokemonCenter headers to the request
func AddPokemonCenterHeaders(request *http.Request, referer ...string) {
	util.AddBaseHeaders(request)
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Origin", "https://www.pokemoncenter.com")
	// omitcsrfjwt: true
	// omitcorrelationid: true
	// credentials: include
	// TODO: Header order
	if len(referer) != 0 {
		request.Header.Set("Referer", referer[0])
	}
}

func HandleDatadomeMonitor(monitor *entities.BaseMonitor, body string) error {
	return datadome.HandleDatadomeMonitor(monitor, enums.PokemonCenter, BaseURL, "https://www.pokemoncenter.com/", "https://www.pokemoncenter.com", ".pokemoncenter.com", body)
}

func HandleDatadomeTask(task *entities.BaseTask, body string) error {
	return datadome.HandleDatadomeTask(task, enums.PokemonCenter, BaseURL, "https://www.pokemoncenter.com/", "https://www.pokemoncenter.com", ".pokemoncenter.com", body)
}

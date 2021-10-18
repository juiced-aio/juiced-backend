package gamestop

import (
	"fmt"

	"backend.juicedbot.io/juiced.infrastructure/entities"
	"backend.juicedbot.io/juiced.infrastructure/enums"
	"backend.juicedbot.io/juiced.sitescripts/util"
)

func Login(account *entities.Account) error {
	client, err := util.CreateClient()
	if err != nil {
		return err
	}

	resp, _, err := util.MakeRequest(&util.Request{
		Client:     &client,
		Method:     "GET",
		URL:        BaseLoginEndpoint,
		RawHeaders: util.DefaultRawHeaders,
	})
	switch resp.StatusCode {
	case 200:
	default:
		if err != nil {
			return err
		} else {
			return fmt.Errorf(enums.UnknownError, resp.StatusCode)
		}
	}

	// csrf, err := u.FindInString(body, `name="csrf_token" value="`, `"`)
	// if err != nil {
	// 	return err
	// }

	return nil
}

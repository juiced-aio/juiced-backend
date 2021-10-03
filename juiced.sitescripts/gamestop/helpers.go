package gamestop

import (
	"errors"
	"time"

	"backend.juicedbot.io/juiced.infrastructure/captcha"
	"backend.juicedbot.io/juiced.infrastructure/entities"
	"backend.juicedbot.io/juiced.infrastructure/enums"
)

func (task *Task) RequestCaptchaToken(url, action string, score float32) (string, error) {
	proxy := entities.Proxy{}
	if task.BaseTask.Proxy != nil {
		proxy = *task.BaseTask.Proxy
	}
	token, err := captcha.RequestCaptchaToken(enums.ReCaptchaV2, enums.GameStop, url, action, score, proxy)
	if err != nil {
		return "", err
	}
	for token == nil {
		needToStop := task.BaseTask.CheckForStop()
		if needToStop {
			return "", nil
		}
		token = captcha.PollCaptchaTokens(enums.ReCaptchaV2, enums.GameStop, task.BaseTask.ProductInfo.ItemURL, proxy)
		time.Sleep(1 * time.Second / 10)
	}
	tokenInfo, ok := token.(entities.ReCaptchaToken)
	if !ok {
		return "", errors.New(enums.BadCaptchaTokenError)
	}

	return tokenInfo.Token, nil
}

package controller

import (
	"backend.juicedbot.io/juiced.api/responses"
	"backend.juicedbot.io/juiced.infrastructure/captcha"
	"backend.juicedbot.io/juiced.infrastructure/entities"
	"backend.juicedbot.io/juiced.infrastructure/staticstores"

	"github.com/gofiber/fiber/v2"
)

func GetSettings(c *fiber.Ctx) error {
	return c.Status(200).JSON(staticstores.GetSettings())
}

func UpdateSettings(c *fiber.Ctx) error {
	var newSettings entities.Settings

	if err := c.BodyParser(&newSettings); err != nil {
		return responses.ReturnErrorResponse(c, responses.UpdateSettingsParseErrorResponse, err)
	}

	currentSettings := staticstores.GetSettings()

	if newSettings.SuccessDiscordWebhook == "-1" {
		newSettings.SuccessDiscordWebhook = currentSettings.SuccessDiscordWebhook
	}
	if newSettings.FailureDiscordWebhook == "-1" {
		newSettings.FailureDiscordWebhook = currentSettings.FailureDiscordWebhook
	}
	if newSettings.TwoCaptchaAPIKey == "-1" {
		newSettings.TwoCaptchaAPIKey = currentSettings.TwoCaptchaAPIKey
	}
	if newSettings.AntiCaptchaAPIKey == "-1" {
		newSettings.AntiCaptchaAPIKey = currentSettings.AntiCaptchaAPIKey
	}
	if newSettings.CapMonsterAPIKey == "-1" {
		newSettings.CapMonsterAPIKey = currentSettings.CapMonsterAPIKey
	}
	aycdChanged := false
	if newSettings.AYCDAccessToken == "-1" {
		newSettings.AYCDAccessToken = currentSettings.AYCDAccessToken
	} else {
		aycdChanged = true
	}
	if newSettings.AYCDAPIKey == "-1" {
		newSettings.AYCDAPIKey = currentSettings.AYCDAPIKey
	} else {
		aycdChanged = true
	}
	if !newSettings.DarkModeUpdate {
		newSettings.DarkMode = currentSettings.DarkMode
	}
	if !newSettings.UseAnimationsUpdate {
		newSettings.UseAnimations = currentSettings.UseAnimations
	}

	err := staticstores.UpdateSettings(newSettings)
	if err != nil {
		return responses.ReturnErrorResponse(c, responses.UpdateSettingsUpdateErrorResponse, err)
	}

	if aycdChanged && newSettings.AYCDAccessToken != "" && newSettings.AYCDAPIKey != "" {
		err = captcha.ConnectToAycd(newSettings.AYCDAccessToken, newSettings.AYCDAPIKey)
		if err != nil {
			return responses.ReturnErrorResponse(c, responses.UpdateSettingsStartAYCDWarningResponse, err)
		}
	}

	return responses.ReturnSuccessResponse(c)
}

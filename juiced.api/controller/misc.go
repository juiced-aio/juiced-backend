package controller

import (
	"sync"
	"time"

	"backend.juicedbot.io/juiced.api/requests"
	"backend.juicedbot.io/juiced.api/responses"
	"backend.juicedbot.io/juiced.infrastructure/discord"
	rpc "backend.juicedbot.io/juiced.rpc"

	"github.com/gofiber/fiber/v2"
)

func SetVersion(c *fiber.Ctx) error {
	var input requests.SetVersionRequest

	if err := c.BodyParser(&input); err != nil {
		return responses.ReturnResponse(c, responses.SetVersionParseErrorResponse, err)
	}

	if input.Version == "" {
		return responses.ReturnResponse(c, responses.SetVersionEmptyInputErrorResponse, nil)
	}

	if err := rpc.SetActivity(input.Version); err != nil {
		return responses.ReturnResponse(c, responses.SetVersionStartRPCWarningResponse, nil)
	}

	return responses.ReturnSuccessResponse(c)
}

func TestDiscord(c *fiber.Ctx) error {
	var input requests.TestDiscordRequest

	if err := c.BodyParser(&input); err != nil {
		return responses.ReturnResponse(c, responses.TestDiscordParseErrorResponse, err)
	}

	if input.SuccessWebhook == "" && input.FailureWebhook == "" {
		return responses.ReturnResponse(c, responses.TestDiscordEmptyInputErrorResponse, nil)
	}

	embed := discord.DiscordEmbed{
		Footer: discord.DiscordFooter{
			Text:    "Juiced",
			IconURL: "https://media.discordapp.net/attachments/849430464036077598/855979506204278804/Icon_1.png?width=128&height=128",
		},
		Timestamp: time.Now(),
	}

	var successErr error
	var failureErr error
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		if input.SuccessWebhook != "" {
			embed.Title = "Success Webhook"
			embed.Color = 16742912
			successErr = discord.SendDiscordWebhook(input.SuccessWebhook, []discord.DiscordEmbed{embed})
		}
		wg.Done()
	}()
	go func() {
		if input.FailureWebhook != "" {
			embed.Title = "Failure Webhook"
			embed.Color = 14495044
			failureErr = discord.SendDiscordWebhook(input.FailureWebhook, []discord.DiscordEmbed{embed})
		}
		wg.Done()
	}()
	wg.Wait()

	if successErr != nil || failureErr != nil {
		if successErr != nil {
			return responses.ReturnResponse(c, responses.TestDiscordParseErrorResponse, successErr)
		}
		return responses.ReturnResponse(c, responses.TestDiscordParseErrorResponse, failureErr)
	}

	return c.SendStatus(200)
}

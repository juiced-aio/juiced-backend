package controller

import (
	"backend.juicedbot.io/juiced.api/requests"
	"backend.juicedbot.io/juiced.api/responses"
	"backend.juicedbot.io/juiced.infrastructure/entities"
	"backend.juicedbot.io/juiced.infrastructure/stores"
	"github.com/gofiber/fiber/v2"
)

func GetAccounts(c *fiber.Ctx) error {
	return c.Status(200).JSON(stores.GetAllAccounts())
}

func CreateAccount(c *fiber.Ctx) error {
	var account entities.Account

	if err := c.BodyParser(&account); err != nil {
		return responses.ReturnResponse(c, responses.CreateAccountParseErrorResponse, err)
	}

	accountPtr, err := stores.CreateAccount(account)
	if err != nil {
		return responses.ReturnResponse(c, responses.CreateAccountCreateErrorResponse, err)
	}

	return c.Status(200).JSON(*accountPtr)
}

func UpdateAccount(c *fiber.Ctx) error {
	accountID := c.Params("id")
	if accountID == "" {
		return responses.ReturnResponse(c, responses.UpdateAccountEmptyIDErrorResponse, nil)
	}

	var newAccount entities.Account

	if err := c.BodyParser(&newAccount); err != nil {
		return responses.ReturnResponse(c, responses.UpdateAccountParseErrorResponse, err)
	}

	_, err := stores.GetAccount(accountID)
	if err != nil {
		return responses.ReturnResponse(c, responses.UpdateAccountNotFoundErrorResponse, err)
	}

	accountPtr, err := stores.UpdateAccount(accountID, newAccount)
	if err != nil {
		return responses.ReturnResponse(c, responses.UpdateAccountUpdateErrorResponse, err)
	}

	return c.Status(200).JSON(*accountPtr)
}

func DeleteAccounts(c *fiber.Ctx) error {
	var input requests.AccountsRequest
	var err error

	if err = c.BodyParser(&input); err != nil {
		return responses.ReturnResponse(c, responses.DeleteAccountsParseErrorResponse, err)
	}

	if len(input.AccountIDs) == 0 {
		return responses.ReturnResponse(c, responses.DeleteAccountsEmptyInputErrorResponse, nil)
	}

	response := responses.AccountsSuccessResponse{}
	for _, accountID := range input.AccountIDs {
		_, err_ := stores.RemoveAccount(accountID)
		if err_ == nil {
			response.SuccessAccountIDs = append(response.SuccessAccountIDs, accountID)
		} else {
			if err == nil {
				err = err_
			}
			response.FailureAccountIDs = append(response.FailureAccountIDs, accountID)
		}
	}

	if len(response.SuccessAccountIDs) == 0 {
		return responses.ReturnResponse(c, responses.DeleteAccountsDeleteErrorResponse, err)
	}

	return c.Status(200).JSON(response)
}

func LogIntoAccounts(c *fiber.Ctx) error {
	var input requests.AccountsRequest
	var err error

	if err = c.BodyParser(&input); err != nil {
		return responses.ReturnResponse(c, responses.LogIntoAccountsParseErrorResponse, err)
	}

	if len(input.AccountIDs) == 0 {
		return responses.ReturnResponse(c, responses.LogIntoAccountsEmptyInputErrorResponse, nil)
	}

	response := responses.AccountsSuccessResponse{}
	for _, accountID := range input.AccountIDs {
		account, err_ := stores.GetAccount(accountID)
		if err_ == nil {
			err_ = stores.AccountLogin(account)
			if err_ == nil {
				response.SuccessAccountIDs = append(response.SuccessAccountIDs, accountID)
			} else {
				if err == nil {
					err = err_
				}
				response.FailureAccountIDs = append(response.FailureAccountIDs, accountID)
			}
		} else {
			if err == nil {
				err = err_
			}
			response.FailureAccountIDs = append(response.FailureAccountIDs, accountID)
		}
	}

	if len(response.SuccessAccountIDs) == 0 {
		return responses.ReturnResponse(c, responses.LogIntoAccountsLogInErrorResponse, err)
	}

	return c.Status(200).JSON(response)
}

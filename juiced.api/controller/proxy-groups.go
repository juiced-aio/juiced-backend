package controller

import (
	"backend.juicedbot.io/juiced.api/requests"
	"backend.juicedbot.io/juiced.api/responses"
	"backend.juicedbot.io/juiced.infrastructure/entities"
	"backend.juicedbot.io/juiced.infrastructure/stores"
	"github.com/gofiber/fiber/v2"
)

func GetProxyGroups(c *fiber.Ctx) error {
	return c.Status(200).JSON(stores.GetAllProxyGroups())
}

func CreateProxyGroup(c *fiber.Ctx) error {
	var proxyGroup entities.ProxyGroup

	if err := c.BodyParser(&proxyGroup); err != nil {
		return responses.ReturnResponse(c, responses.CreateProxyGroupParseErrorResponse, err)
	}

	proxyGroupPtr, err := stores.CreateProxyGroup(proxyGroup)
	if err != nil {
		return responses.ReturnResponse(c, responses.CreateProxyGroupCreateErrorResponse, err)
	}

	return c.Status(200).JSON(*proxyGroupPtr)
}

func UpdateProxyGroup(c *fiber.Ctx) error {
	proxyGroupID := c.Params("id")
	if proxyGroupID == "" {
		return responses.ReturnResponse(c, responses.UpdateProxyGroupEmptyIDErrorResponse, nil)
	}

	var newProxyGroup entities.ProxyGroup

	if err := c.BodyParser(&newProxyGroup); err != nil {
		return responses.ReturnResponse(c, responses.UpdateProxyGroupParseErrorResponse, err)
	}

	_, err := stores.GetProxyGroup(proxyGroupID)
	if err != nil {
		return responses.ReturnResponse(c, responses.UpdateProxyGroupNotFoundErrorResponse, err)
	}

	proxyGroupPtr, err := stores.UpdateProxyGroup(proxyGroupID, newProxyGroup)
	if err != nil {
		return responses.ReturnResponse(c, responses.UpdateProxyGroupUpdateErrorResponse, err)
	}

	return c.Status(200).JSON(*proxyGroupPtr)
}

func DeleteProxyGroups(c *fiber.Ctx) error {
	var input requests.ProxyGroupsRequest
	var err error

	if err = c.BodyParser(&input); err != nil {
		return responses.ReturnResponse(c, responses.DeleteProxyGroupsParseErrorResponse, err)
	}

	if len(input.ProxyGroupIDs) == 0 {
		return responses.ReturnResponse(c, responses.DeleteProxyGroupsEmptyInputErrorResponse, nil)
	}

	response := responses.ProxyGroupsSuccessResponse{}
	for _, proxyGroupID := range input.ProxyGroupIDs {
		_, err_ := stores.RemoveProxyGroup(proxyGroupID)
		if err_ == nil {
			response.SuccessProxyGroupIDs = append(response.SuccessProxyGroupIDs, proxyGroupID)
		} else {
			if err == nil {
				err = err_
			}
			response.FailureProxyGroupIDs = append(response.FailureProxyGroupIDs, proxyGroupID)
		}
	}

	if len(response.SuccessProxyGroupIDs) == 0 {
		return responses.ReturnResponse(c, responses.DeleteProxyGroupsDeleteErrorResponse, err)
	}

	return c.Status(200).JSON(response)
}

func CloneProxyGroups(c *fiber.Ctx) error {
	var input requests.ProxyGroupsRequest
	var err error

	if err = c.BodyParser(&input); err != nil {
		return responses.ReturnResponse(c, responses.DeleteProxyGroupsParseErrorResponse, err)
	}

	if len(input.ProxyGroupIDs) == 0 {
		return responses.ReturnResponse(c, responses.DeleteProxyGroupsEmptyInputErrorResponse, nil)
	}

	proxyGroups := []entities.ProxyGroup{}
	success := false
	for _, proxyGroupID := range input.ProxyGroupIDs {
		newProxyGroupPtr, err_ := stores.CloneProxyGroup(proxyGroupID)
		if err_ == nil {
			success = true
			proxyGroups = append(proxyGroups, *newProxyGroupPtr)
		} else {
			if err == nil {
				err = err_
			}
		}
	}

	if !success {
		return responses.ReturnResponse(c, responses.DeleteProxyGroupsDeleteErrorResponse, err)
	}

	return c.Status(200).JSON(proxyGroups)
}

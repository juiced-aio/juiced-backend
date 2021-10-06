package shopify

import (
	"strings"

	"backend.juicedbot.io/juiced.infrastructure/enums"
)

const SKU_FORMAT = "XXXXXXXXXXXXXX"

func ValidateMonitorInput(input string, monitorType enums.MonitorType, info map[string]interface{}) (MonitorInput, error) {
	shopifyMonitorInput := MonitorInput{}
	switch monitorType {
	case enums.SKUMonitor:
		if strings.Contains(input, "https") {
			return shopifyMonitorInput, &enums.InvalidSKUError{Retailer: enums.Shopify, Format: SKU_FORMAT}
		}
	case enums.URLMonitor:
		if !strings.Contains(input, "https") {
			return shopifyMonitorInput, &enums.InputIsNotURLError{Retailer: enums.Shopify}
		}

	default:
		return shopifyMonitorInput, &enums.UnsupportedMonitorTypeError{Retailer: enums.Shopify, MonitorType: monitorType}
	}

	shopifyRetailer, ok := info["shopifyRetailer"].(string)
	if !ok {
		return shopifyMonitorInput, &enums.InvalidInputTypeError{Field: "shopifyRetailer", ShouldBe: "string"}
	}
	shopifyMonitorInput.ShopifyRetailer = shopifyRetailer
	if shopifyRetailer == "" {
		return shopifyMonitorInput, &enums.EmptyInputFieldError{Field: "shopifyRetailer"}
	}

	siteURL, ok := info["siteURL"].(string)
	if !ok {
		return shopifyMonitorInput, &enums.InvalidInputTypeError{Field: "siteURL", ShouldBe: "string"}
	}
	shopifyMonitorInput.SiteURL = siteURL
	if shopifyRetailer == enums.GenericShopify {
		if siteURL == "" {
			return shopifyMonitorInput, &enums.EmptyInputFieldError{Field: "siteURL"}
		}
	} else {
		shopifyMonitorInput.SiteURL = enums.ShopifySiteURLs[shopifyRetailer]
		if shopifyMonitorInput.SiteURL == "" {
			return shopifyMonitorInput, &enums.InvalidRetailerError{Retailer: shopifyRetailer}
		}
	}

	sitePassword, ok := info["sitePassword"].(string)
	if !ok {
		return shopifyMonitorInput, &enums.InvalidInputTypeError{Field: "sitePassword", ShouldBe: "string"}
	}
	shopifyMonitorInput.SitePassword = sitePassword

	return shopifyMonitorInput, nil
}

func ValidateTaskInput(info map[string]interface{}) (TaskInput, error) {
	shopifyTaskInput := TaskInput{}

	shopifyRetailer, ok := info["shopifyRetailer"].(string)
	if !ok {
		return shopifyTaskInput, &enums.InvalidInputTypeError{Field: "shopifyRetailer", ShouldBe: "string"}
	}
	shopifyTaskInput.ShopifyRetailer = shopifyRetailer
	if shopifyRetailer == "" {
		return shopifyTaskInput, &enums.EmptyInputFieldError{Field: "shopifyRetailer"}
	}

	siteURL, ok := info["siteURL"].(string)
	if !ok {
		return shopifyTaskInput, &enums.InvalidInputTypeError{Field: "siteURL", ShouldBe: "string"}
	}
	shopifyTaskInput.SiteURL = siteURL
	if shopifyRetailer == enums.GenericShopify {
		if siteURL == "" {
			return shopifyTaskInput, &enums.EmptyInputFieldError{Field: "siteURL"}
		}
	} else {
		shopifyTaskInput.SiteURL = enums.ShopifySiteURLs[shopifyRetailer]
		if shopifyTaskInput.SiteURL == "" {
			return shopifyTaskInput, &enums.InvalidRetailerError{Retailer: shopifyRetailer}
		}
	}

	sitePassword, ok := info["sitePassword"].(string)
	if !ok {
		return shopifyTaskInput, &enums.InvalidInputTypeError{Field: "sitePassword", ShouldBe: "string"}
	}
	shopifyTaskInput.SitePassword = sitePassword

	couponCode, ok := info["couponCode"].(string)
	if !ok {
		return shopifyTaskInput, &enums.InvalidInputTypeError{Field: "couponCode", ShouldBe: "string"}
	}
	shopifyTaskInput.CouponCode = couponCode

	if info["taskType"] == enums.TaskTypeAccount {
		shopifyTaskInput.TaskType = enums.TaskTypeAccount
		if email, ok := info["email"].(string); !ok {
			return shopifyTaskInput, &enums.InvalidInputTypeError{Field: "email", ShouldBe: "string"}
		} else {
			if email == "" {
				return shopifyTaskInput, &enums.EmptyInputFieldError{Field: "email"}
			}
			shopifyTaskInput.Email = email
		}
		if password, ok := info["password"].(string); !ok {
			return shopifyTaskInput, &enums.InvalidInputTypeError{Field: "password", ShouldBe: "string"}
		} else {
			if password == "" {
				return shopifyTaskInput, &enums.EmptyInputFieldError{Field: "password"}
			}
			shopifyTaskInput.Password = password
		}
	} else {
		shopifyTaskInput.TaskType = enums.TaskTypeGuest
	}

	return shopifyTaskInput, nil
}

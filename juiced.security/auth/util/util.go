package util

import (
	"strings"
	"time"

	"backend.juicedbot.io/juiced.infrastructure/discord"
	"backend.juicedbot.io/juiced.infrastructure/entities"
	"backend.juicedbot.io/juiced.infrastructure/enums"
	"backend.juicedbot.io/juiced.infrastructure/staticstores"
)

func ProcessCheckout(task *entities.BaseTask, pci ProcessCheckoutInfo) {
	userInfo := staticstores.GetUserInfo()
	if !strings.Contains(task.Status, enums.OrderStatusFailed) {
		go DiscordWebhook(pci.Success, pci.Content, pci.Embeds, userInfo)
	}
	if pci.Success {
		go LogCheckout(pci.ProductInfo.ItemName, pci.ProductInfo.SKU, pci.Retailer, int(pci.ProductInfo.Price), pci.Quantity, userInfo)
		staticstores.CreateCheckout(entities.Checkout{
			ItemName:     pci.ProductInfo.ItemName,
			ImageURL:     pci.ProductInfo.ImageURL,
			SKU:          pci.ProductInfo.SKU,
			Price:        int(pci.ProductInfo.Price),
			Quantity:     pci.Quantity,
			Retailer:     task.Task.Retailer,
			ProfileName:  task.Profile.Name,
			MsToCheckout: pci.MsToCheckout,
			Time:         time.Now().Unix(),
		})
	}
	discord.QueueWebhook(pci.Success, pci.Content, pci.Embeds)
}

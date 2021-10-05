package shopify

import "strings"

func SplitCard(card string, cardType string) string {
	var cardSplit string
	switch cardType {
	case "AMEX":
		cardSplit = strings.Join([]string{card[:4], card[4:10], card[10:15]}, " ")
	case "Diners":
		cardSplit = strings.Join([]string{card[:4], card[4:10], card[10:14]}, " ")
	default:
		cardSplit = strings.Join([]string{card[:4], card[4:8], card[8:12], card[12:16]}, " ")
	}
	return cardSplit
}

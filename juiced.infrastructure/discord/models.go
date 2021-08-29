package discord

import (
	"time"
)

type WebhookInfo struct {
	Success bool
	Content string
	Embeds  []DiscordEmbed
}

type DiscordWebhook struct {
	Content interface{}    `json:"content"`
	Embeds  []DiscordEmbed `json:"embeds"`
}

type DiscordEmbed struct {
	Title     string           `json:"title"`
	Color     int              `json:"color"`
	Fields    []DiscordField   `json:"fields"`
	Footer    DiscordFooter    `json:"footer"`
	Timestamp time.Time        `json:"timestamp"`
	Thumbnail DiscordThumbnail `json:"thumbnail"`
}

type DiscordField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline,omitempty"`
}
type DiscordFooter struct {
	Text    string `json:"text"`
	IconURL string `json:"icon_url"`
}

type DiscordThumbnail struct {
	URL string `json:"url"`
}

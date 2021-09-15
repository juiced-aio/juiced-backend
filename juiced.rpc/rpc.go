package rpc

import (
	"time"

	"github.com/hugolgst/rich-go/client"
)

func EnableRPC() {
	client.Login("856936229223006248")
}

func SetActivity(version, channel string) error {
	start := time.Now()
	return client.SetActivity(client.Activity{
		Details:    version + channel,
		LargeImage: "main-juiced",
		LargeText:  "Juiced",
		SmallImage: "",
		SmallText:  "",
		Timestamps: &client.Timestamps{
			Start: &start,
		},
		Buttons: []*client.Button{
			{
				Label: "Dashboard",
				Url:   "https://dash.juicedbot.io/",
			},
		},
	})

}

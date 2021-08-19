package rpc

import (
	"time"

	"github.com/hugolgst/rich-go/client"
)

func EnableRPC() {
	client.Login("856936229223006248")
}

func SetActivity(ver string) {
	// No need to close the app if Discord RPC doesn't work. It's not a necessary feature.
	// If it breaks for everyone at once for some reason, don't want to entirely break the app without a hotfix.
	start := time.Now()
	client.SetActivity(client.Activity{
		Details:    "Beta - " + ver,
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

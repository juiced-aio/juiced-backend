package client

import (
	"fmt"
	"os"

	"backend.juicedbot.io/juiced.client/http"
)

var siteCerts = map[string][]string{
	"www.amazon.com":        {"f1mza6JyUxJihUMV4oh2e0dujS/lLV0Su9eZMjGBDBE=", "njN4rRG+22dNXAi+yb8e3UMypgzPUPHlv4+foULwl1g=", "i7WTqTvh0OioIruIfFR4kMPnBqrS2rdiVPl/s2uC/CY="},
	"www.bestbuy.com":       {"qhf2X0NH3RPgHLIGsxALo6Ut32jgKWOE48mRT6qMcqI=", "RRM1dGqnDFsCJXBTHky16vi1obOlCgFFn/yOhI/y+ho="},
	"www.gamestop.com":      {"wViJygUvgi0205z+xQfyvKNdRhNujpx+sHjt+9qqfdI=", "zUIraRNo+4JoAYA7ROeWjARtIoN4rIEbCpfCRQT6N6A="},
	"www.hottopic.com":      {"YL/kd6yshmbiq7TUP7TSfvPTSjEJY4IIMJitZ6G09kM=", "RRM1dGqnDFsCJXBTHky16vi1obOlCgFFn/yOhI/y+ho="},
	"www.target.com":        {"myQ7PtFcXzJLx6AD0hxWIYflpDcSTha3pXVEzQB+/Gs=", "S7kwF/US+qCLAH7QPb4nX6Ms8I/NUy0GV9/+wVRxQe4="},
	"www.walmart.com":       {"Q2B6B1QYMuF5EzLRVgxdis/jvrOYWINganul0jJ9BB8=", "hETpgVvaLC0bvcGG3t0cuqiHvr4XyP2MTwCiqhgRWwU=", "cGuxAXyFXFkWm61cF4HPWX8S0srS9j0aSqN0k4AP+4A="},
	"identity.juicedbot.io": {"0Ugw2FeRziz9vmBmylwjswrF8pQ8icmeqRweSfkkGAQ=", "n5dIU+KFaI00Y/prmvaZhqXOquF72TlPANCLxCA9HE8="},
}

func GetCerts(client http.Client) {
	os.Setenv("JUICED_MODE", "CERTS")
	for key := range siteCerts {
		_, err := client.Get("https://" + key)
		if err != nil {
			fmt.Println(err)
		}
	}
}

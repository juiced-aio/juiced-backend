package util

import (
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"backend.juicedbot.io/juiced.client/http"
	"backend.juicedbot.io/juiced.infrastructure/common"
	"backend.juicedbot.io/juiced.infrastructure/common/captcha"
	"backend.juicedbot.io/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/juiced.infrastructure/common/enums"
	"backend.juicedbot.io/juiced.infrastructure/common/events"
)

func TestMain(m *testing.M) {
	events.InitEventBus()
	eventBus := events.GetEventBus()
	common.InitDatabase()
	captcha.InitCaptchaStore(eventBus)
	m.Run()
}

func TestCreateClient(t *testing.T) {
	type args struct {
		proxy []entities.Proxy
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "Success W/O Proxy", wantErr: false},
		{name: "Success W Regular Proxy", args: args{proxy: []entities.Proxy{{Host: "localhost", Port: "3000"}}}, wantErr: false},
		{name: "Success W User-Pass Proxy", args: args{proxy: []entities.Proxy{{Host: "localhost", Port: "3000", Username: "admin", Password: "password"}}}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := CreateClient(tt.args.proxy...)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestMakeRequest(t *testing.T) {
	client, _ := CreateClient()
	type args struct {
		requestInfo *Request
	}

	tests := []struct {
		name     string
		args     args
		wantBody string
		wantErr  bool
	}{
		{name: "Bad Method", args: args{requestInfo: &Request{Client: client, Method: "NOT A REAL METHOD"}}, wantErr: true},
		{name: "Bad Request", args: args{requestInfo: &Request{Client: client, Method: "GET", URL: "BAD URL"}}, wantErr: true},
		{name: "Correct Body", args: args{requestInfo: &Request{Client: client, Method: "GET", URL: "https://jsonplaceholder.typicode.com/todos/1"}}, wantErr: false, wantBody: `{  "userId": 1,  "id": 1,  "title": "delectus aut autem",  "completed": false}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, got1, err := MakeRequest(tt.args.requestInfo)
			if (err != nil) != tt.wantErr {
				t.Errorf("MakeRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got1 != tt.wantBody {
				t.Errorf("MakeRequest() got1 = %v, want %v", got1, tt.wantBody)
			}
		})
	}
}

func TestSendDiscordWebhook(t *testing.T) {
	type args struct {
		discordWebhook string
		embeds         []Embed
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "Bad discordWebhook", args: args{discordWebhook: "hts://ptb.discord.com/api/webhooks/842204922404405288/DF10h94BxLRv-yYngD2O6JnPCveZuQ3eNKFI-_-SXJMkPRldBzdmKJizRY9aotzaGN5v"}, want: false},
		{name: "Fail", args: args{discordWebhook: "https://ptb.discord.com/api/webhooks/842204922404405288/DF10h94BxLRv-yYngD2O6JnPCveZuQ3eNKFI-_-SXJMkPRldBzdmKJizRY9aotzaGN5v", embeds: []Embed{{Title: "", Color: 0, Fields: []Field{}, Footer: Footer{}, Timestamp: time.Time{}, Thumbnail: Thumbnail{URL: ""}}}}, want: false},
		{name: "Success", args: args{discordWebhook: "https://ptb.discord.com/api/webhooks/842204922404405288/DF10h94BxLRv-yYngD2O6JnPCveZuQ3eNKFI-_-SXJMkPRldBzdmKJizRY9aotzaGN5v", embeds: []Embed{{Title: "Success", Color: 0, Fields: []Field{}, Footer: Footer{}, Timestamp: time.Time{}, Thumbnail: Thumbnail{URL: ""}}}}, want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SendDiscordWebhook(tt.args.discordWebhook, tt.args.embeds); got != tt.want {
				t.Errorf("SendDiscordWebhook() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCreateParams(t *testing.T) {
	type args struct {
		paramsLong map[string]string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{name: "One", args: args{paramsLong: map[string]string{"ONE": "TRUE"}}, want: "ONE=TRUE"},
		{name: "Two", args: args{paramsLong: map[string]string{"ONE": "TRUE", "TWO": "TRUE"}}, want: "ONE=TRUE&TWO=TRUE"},
		{name: "Three", args: args{paramsLong: map[string]string{"ONE": "TRUE", "TWO": "TRUE", "THREE": "TRUE"}}, want: "ONE=TRUE&TWO=TRUE&THREE=TRUE"},
		{name: "Three Bad", args: args{paramsLong: map[string]string{"ONE": "TRUE", "TWO": "TRUE", "THREE": "TRUE"}}, want: "ONE=TRUE&TWO=TRUE&THREE=WRONG", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			success := true
			params := make(map[string]string)
			badParams := []string{}
			got := CreateParams(tt.args.paramsLong)
			splitted1 := strings.Split(got, "&")
			for _, split1 := range splitted1 {
				splitted2 := strings.Split(split1, "=")
				params[splitted2[0]] = splitted2[1]
			}
			wantSplitted1 := strings.Split(tt.want, "&")
			for _, wantSplit1 := range wantSplitted1 {
				wantSplitted2 := strings.Split(wantSplit1, "=")
				param, ok := params[wantSplitted2[0]]
				if !ok {
					success = false
					badParams = append(badParams, wantSplitted2[0])
					break
				}
				if param != wantSplitted2[1] {
					success = false
					badParams = append(badParams, wantSplitted2[0])
					break
				}
			}
			if !success && !tt.wantErr {
				t.Errorf("CreateParams() returned wrong key(s) %v", badParams)
			}
		})
	}
}

func TestTernaryOperator(t *testing.T) {
	type args struct {
		condition    bool
		trueOutcome  interface{}
		falseOutcome interface{}
	}
	tests := []struct {
		name string
		args args
		want interface{}
	}{
		{name: "True", args: args{condition: "TEST" == "TEST", trueOutcome: "TRUE", falseOutcome: "FALSE"}, want: "TRUE"},
		{name: "False", args: args{condition: "TEST" == "TES", trueOutcome: "TRUE", falseOutcome: "FALSE"}, want: "FALSE"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := TernaryOperator(tt.args.condition, tt.args.trueOutcome, tt.args.falseOutcome); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TernaryOperator() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProxyCleaner(t *testing.T) {
	type args struct {
		proxyDirty entities.Proxy
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "Incorrect Proxy", args: args{proxyDirty: entities.Proxy{}}, want: ""},
		{name: "User-Pass Proxy", args: args{proxyDirty: entities.Proxy{Host: "randomprovider", Port: "3000", Username: "admin", Password: "password"}}, want: "http://admin:password@randomprovider:3000"},
		{name: "Regular Proxy", args: args{proxyDirty: entities.Proxy{Host: "randomprovider", Port: "3000"}}, want: "http://randomprovider:3000"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ProxyCleaner(tt.args.proxyDirty); got != tt.want {
				t.Errorf("ProxyCleaner() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFindInString(t *testing.T) {
	type args struct {
		str   string
		start string
		end   string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{name: "Success", args: args{str: "{TEST}", start: "{", end: "}"}, want: "TEST", wantErr: false},
		{name: "Cant Find", args: args{str: "{TEST", start: "{", end: "}"}, want: "", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FindInString(tt.args.str, tt.args.start, tt.args.end)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindInString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("FindInString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewAbck(t *testing.T) {
	client, _ := CreateClient()
	bestbuyURL := "https://www.bestbuy.com"
	akamaiURL := "https://www.bestbuy.com/Z43Qo-szvQDrezPFUWbI-oosQsM/9YOhShXz9OX1/D3ZjQkgC/EWdSfC5P/DlY"
	type args struct {
		abckClient     *http.Client
		location       string
		BaseEndpoint   string
		AkamaiEndpoint string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "Success", args: args{abckClient: &client, location: bestbuyURL, BaseEndpoint: bestbuyURL, AkamaiEndpoint: akamaiURL}, wantErr: false},
		{name: "Bad Url1", args: args{abckClient: &client, location: bestbuyURL, BaseEndpoint: bestbuyURL, AkamaiEndpoint: "BAD URL"}, wantErr: true},
		{name: "Bad Url2", args: args{abckClient: &client, location: "BAD URL", BaseEndpoint: bestbuyURL, AkamaiEndpoint: "BAD URL"}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := NewAbck(tt.args.abckClient, tt.args.location, tt.args.BaseEndpoint, tt.args.AkamaiEndpoint); (err != nil) != tt.wantErr {
				t.Errorf("NewAbck() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetPXCookie(t *testing.T) {
	walmartURL := "https://www.walmart.com"
	type args struct {
		site  string
		proxy entities.Proxy
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "Success", args: args{site: walmartURL}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := GetPXCookie(tt.args.site, tt.args.proxy)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetPXCookie() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

		})
	}
}

func TestGetPXCapCookie(t *testing.T) {
	walmartURL := "https://www.walmart.com"
	_, pxValues, err := GetPXCookie(walmartURL, entities.Proxy{})
	if err != nil {
		t.Fatal(err)
	}
	requestCaptchaTokenInfo := RequestCaptchaTokenInfo{}
	requestCaptchaTokenInfo.CaptchaType = enums.ReCaptchaV2
	requestCaptchaTokenInfo.Retailer = enums.Walmart
	requestCaptchaTokenInfo.Url = walmartURL + "/blocked"
	requestCaptchaTokenInfo.MinScore = 0

	token, err := captcha.RequestCaptchaToken(requestCaptchaTokenInfo)
	if err != nil {
		t.Fatal(err)
	}
	for token == nil {
		token = captcha.PollCaptchaTokens(enums.ReCaptchaV2, enums.Walmart, walmartURL+"/blocked", entities.Proxy{})
		time.Sleep(1 * time.Second / 10)
	}
	tokenInfo, ok := token.(entities.ReCaptchaToken)
	if !ok {
		err = errors.New("token could not be parsed")
		t.Fatal(err)
	}

	type args struct {
		site  string
		setID string
		vid   string
		uuid  string
		token string
		proxy entities.Proxy
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "Success", args: args{site: walmartURL, setID: pxValues.SetID, vid: pxValues.VID, uuid: pxValues.UUID, token: tokenInfo.Token, proxy: entities.Proxy{}}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := GetPXCapCookie(tt.args.site, tt.args.setID, tt.args.vid, tt.args.uuid, tt.args.token, tt.args.proxy)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetPXCapCookie() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

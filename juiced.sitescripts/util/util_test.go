package util

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"backend.juicedbot.io/juiced.client/http"
	"backend.juicedbot.io/juiced.infrastructure/captcha"
	"backend.juicedbot.io/juiced.infrastructure/database"
	"backend.juicedbot.io/juiced.infrastructure/entities"
	"backend.juicedbot.io/juiced.infrastructure/enums"
	"backend.juicedbot.io/juiced.infrastructure/events"
)

func TestMain(m *testing.M) {
	events.InitEventBus()
	eventBus := events.GetEventBus()
	database.InitDatabase()
	captcha.InitCaptchaStore(eventBus)
	m.Run()
}

func TestMakeRequest(t *testing.T) {
	client := *http.DefaultClient
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
		proxyDirty *entities.Proxy
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "Incorrect Proxy", args: args{proxyDirty: &entities.Proxy{}}, want: ""},
		{name: "User-Pass Proxy", args: args{proxyDirty: &entities.Proxy{Host: "randomprovider", Port: "3000", Username: "admin", Password: "password"}}, want: "http://admin:password@randomprovider:3000"},
		{name: "Regular Proxy", args: args{proxyDirty: &entities.Proxy{Host: "randomprovider", Port: "3000"}}, want: "http://randomprovider:3000"},
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
	//entities.Proxy{Host: "localhost", Port: "8888"}
	client := *http.DefaultClient
	bestbuyURL := "https://www.bestbuy.com/"
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
		proxy *entities.Proxy
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
			_, _, _, err := GetPXCookie(tt.args.site, tt.args.proxy, &CancellationToken{})
			if (err != nil) != tt.wantErr {
				t.Errorf("GetPXCookie() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

		})
	}
}

func TestGetPXCapCookie(t *testing.T) {
	walmartURL := "https://www.walmart.com"
	_, pxValues, _, err := GetPXCookie(walmartURL, &entities.Proxy{}, &CancellationToken{})
	if err != nil {
		t.Fatal(err)
	}

	token, err := captcha.RequestCaptchaToken(enums.ReCaptchaV2, enums.Walmart, walmartURL+"/blocked", "", 0, entities.Proxy{})
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
		proxy *entities.Proxy
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "Success", args: args{site: walmartURL, setID: pxValues.SetID, vid: pxValues.VID, uuid: pxValues.UUID, token: tokenInfo.Token, proxy: &entities.Proxy{}}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := GetPXCapCookie(tt.args.site, tt.args.setID, tt.args.vid, tt.args.uuid, tt.args.token, tt.args.proxy, &CancellationToken{})
			if (err != nil) != tt.wantErr {
				t.Errorf("GetPXCapCookie() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestRandomLeastUsedProxy(t *testing.T) {
	type args struct {
		proxies []*entities.Proxy
	}
	tests := []struct {
		args args
		want *entities.Proxy
	}{
		{args{}, &entities.Proxy{}},
		{args{[]*entities.Proxy{{Count: 1}}}, &entities.Proxy{Count: 1}},
		{args{[]*entities.Proxy{{Count: 1}, {Count: 2}, {Count: 3}}}, &entities.Proxy{Count: 1}},
		{args{[]*entities.Proxy{{Count: 1}, {Count: 2}, {Count: 2}, {Count: 2}, {Count: 3}, {Count: 3}, {Count: 3}, {Count: 3}}}, &entities.Proxy{Count: 1}},
	}
	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			if got := RandomLeastUsedProxy(tt.args.proxies); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RandomLeastUsedProxy() = %v, want %v", got, tt.want)
			}
		})
	}
}

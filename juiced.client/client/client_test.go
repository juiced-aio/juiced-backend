package client

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"testing"

	tls "github.com/Titanium-ctrl/utls"

	"backend.juicedbot.io/juiced.client/http"
)

type JA3Response struct {
	JA3Hash   string `json:"ja3_hash"`
	JA3       string `json:"ja3"`
	UserAgent string `json:"User-Agent"`
}

func readAndClose(r io.ReadCloser) ([]byte, error) {
	readBytes, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return readBytes, r.Close()
}

const Chrome83Hash = "b32309a26951912be7dba376398abc3b"

//, "http://localhost:8888"
var client, _ = NewClient(tls.HelloChrome_Auto, "http://209.127.191.180:9279") // cannot throw an error because there is no proxy

func TestCClient_JA3(t *testing.T) {
	resp, err := client.Get("https://ja3er.com/json")
	if err != nil {
		t.Fatal(err)
	}

	respBody, err := readAndClose(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	var ja3Response JA3Response
	if err := json.Unmarshal(respBody, &ja3Response); err != nil {
		t.Fatal(err)
	}

	if ja3Response.JA3Hash != Chrome83Hash {
		t.Error("unexpected JA3 hash; expected:", Chrome83Hash, "| got:", ja3Response.JA3Hash)
	}

}

func TestCClient_HTTP2(t *testing.T) {
	//https://http2.golang.org/serverpush
	req, _ := http.NewRequest("GET", "https://www.google.com", nil)

	req.RawHeader = [][2]string{
		//{"content-length", "5"},
		{"content-type", "application/json"},
		{"cookie", "fie"},
		{"Become", "awdwad"},
		{"Header", "Order"},
		{"accept-encoding", "gzip, deflate, br"},
		{"eee", "e IS ee"},
		{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.212 Safari/537.36"},
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	/* body, err := readAndClose(resp.Body)
	if err != nil {
		t.Fatal("err")
	} */

	resp.Body.Close()
	if resp.ProtoMajor != 2 || resp.ProtoMinor != 0 {
		t.Error("unexpected response proto; expected: HTTP/2.0 | got: ", resp.Proto)
	}

}

func TestProxy(t *testing.T) {
	//
	resp, err := client.Get("https://www.google.com")
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(resp)

	UpdateProxy(&client, "http://localhost:8888")
	resp, err = client.Get("https://www.google.com")
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(resp.StatusCode)

}

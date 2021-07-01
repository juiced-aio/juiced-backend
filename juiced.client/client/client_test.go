package client

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"testing"

	"backend.juicedbot.io/juiced.client/http"
	utls "backend.juicedbot.io/juiced.client/utls"
)

type FingerprintResponse struct {
	ID string `json:"id"`
}

func readAndClose(r io.ReadCloser) ([]byte, error) {
	readBytes, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return readBytes, r.Close()
}

const Chrome90ID = "8466c4390d4bc355"

//, "http://localhost:8888"
//, "http://209.127.191.180:9279"
var client, _ = NewClient(utls.HelloChrome_90, "http://localhost:8888") // cannot throw an error because there is no proxy

func TestClient_ID(t *testing.T) {
	os.Setenv("JUICED_MODE", "DEV")
	resp, err := client.Get("https://client.tlsfingerprint.io:8443/")
	if err != nil {
		t.Fatal(err)
	}

	respBody, err := readAndClose(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	var fingerprintResponse FingerprintResponse
	if err := json.Unmarshal(respBody, &fingerprintResponse); err != nil {
		t.Fatal(err)
	}

	if fingerprintResponse.ID != Chrome90ID {
		t.Error("unexpected ID; expected:", Chrome90ID, "| got:", fingerprintResponse.ID)
	}

}

func TestClient_HTTP2(t *testing.T) {
	os.Setenv("JUICED_MODE", "DEV")
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

	/* 	body, err := readAndClose(resp.Body)
	   	if err != nil {
	   		t.Fatal("err")
	   	}
	   	fmt.Println(string(body)) */
	fmt.Println(resp.Status)
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

func TestGetCerts(t *testing.T) {
	client, _ := NewClient(utls.HelloChrome_90)
	GetCerts(client)
}

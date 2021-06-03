package backend

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"
)

// TestStartTask starts a Test Task which can be modified in the json files
func TestStartTask(t *testing.T) {
	proxyUrl, _ := url.Parse("http://localhost:8888")
	client := http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyUrl),
		}}

	req, err := http.NewRequest("POST", "http://localhost:10000/api/testtask/5ad9a913478c26d220afb681/start", nil)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Origin", "http://localhost:3000")

	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(resp)

}

// TestStopTask stops the Test Task
func TestStopTask(t *testing.T) {
	proxyUrl, _ := url.Parse("http://localhost:8888")
	client := http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyUrl),
		}}

	req, err := http.NewRequest("POST", "http://localhost:10000/api/testtask/5ad9a913478c26d220afb681/stop", nil)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Origin", "http://localhost:3000")

	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(resp)

}

package utils

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
)

var (
	client      *http.Client
	proxyClient *http.Client
	once        sync.Once
)

func GetHttpClient() *http.Client {
	once.Do(func() {
		client = &http.Client{}
	})
	return client
}

func HttpGet(url string, header map[string]string) ([]byte, error) {
	client := GetHttpClient()
	req, _ := http.NewRequest("GET", url, nil)
	if header != nil {
		for k, v := range header {
			req.Header.Add(k, v)
		}
	}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	return respBody, nil
}

func HttpPost(url string, reqBody []byte, query map[string]string) ([]byte, error) {
	client := GetHttpClient()
	payload := bytes.NewReader(reqBody)
	req, _ := http.NewRequest("POST", url, payload)
	if query != nil {
		q := req.URL.Query()
		for k, v := range query {
			q.Add(k, v)
		}
		req.URL.RawQuery = q.Encode()
	}
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	return respBody, nil
}

func GetProxyClient() *http.Client {
	once.Do(func() {
		config, err := LoadConfig()
		if err != nil {
			fmt.Println(err)
		}
		if config.UseProxy {
			// Use proxy in China main land
			proxyURL, err := url.Parse(config.ProxyUrl)
			if err != nil {
				fmt.Println(err)
			}
			transport := &http.Transport{
				Proxy: http.ProxyURL(proxyURL),
			}
			proxyClient = &http.Client{
				Transport: transport,
			}
		} else {
			// Do not use proxy
			proxyClient = &http.Client{}
		}
	})
	return proxyClient
}

func HttpProxyGet(url string) ([]byte, error) {
	proxyClient := GetProxyClient()
	resp, err := proxyClient.Get(url)
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	return respBody, nil
}

package okex

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/net/proxy"
)

const (
	okex_url      = "https://www.okex.com"
	bb_get_ticker = "/api/v1/ticker.do"
)

type Spot_bb_get_ticker_request struct {
	symbol string
}
type Spot_bb_ticker struct {
	Buy  string `json:"buy"`
	High string `json:"high"`
	Last string `json:"last"`
	Low  string `json:"low"`
	Sell string `json:"sell"`
	Vol  string `json:"vol"`
}
type Spot_bb_get_ticker_response struct {
	Date   string         `json:"date"`
	Ticker Spot_bb_ticker `json:"ticker"`
}

func Ex_Spot_bb_get_ticker(symbol string) (string, error) {
	var request Spot_bb_get_ticker_request
	request.symbol = symbol
	resp, err := Spot_bb_get_ticker(request)
	if err != nil {
		return "", err
	}

	jsonBytes, err := json.Marshal(*resp)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), err
}

func Spot_bb_get_ticker(reqeust Spot_bb_get_ticker_request) (*Spot_bb_get_ticker_response, error) {
	url := okex_url + bb_get_ticker + "?symbol=" + reqeust.symbol
	resp, err := spot_bb_get_url(url)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	fmt.Println(string(resp))

	var tickerResp Spot_bb_get_ticker_response
	jsonErr := json.Unmarshal(resp, &tickerResp)
	if jsonErr != nil {
		fmt.Println(jsonErr.Error())
		return nil, jsonErr
	}

	return &tickerResp, jsonErr
}

func spot_bb_get_url(url string) ([]byte, error) {
	//resp, err := http.Get(url)
	timeout := time.Duration(5 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return content, err
}

func proxyHTTP(webUrl string) (*http.Response, error) {
	/*
		1. 代理请求
		2. 跳过https不安全验证
	*/
	// webUrl := "http://ip.gs/"
	var proxyUrl string
	proxyUrl = "http://127.0.0.1:1080"
	proxy, _ := url.Parse(proxyUrl)
	tr := &http.Transport{
		Proxy:           http.ProxyURL(proxy),
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Transport: tr,
		Timeout:   time.Second * 5, //超时时间
	}
	return client.Get(webUrl)

}

func proxySOCKS5(webUrl string) (*http.Response, error) {
	dialer, err := proxy.SOCKS5("tcp", "ubuntu.urwork.qbtrade.org:1080", nil, proxy.Direct)
	if err != nil {
		fmt.Println("cant connect to the proxy : " + err.Error())
		return nil, err
	}

	httpTransport := &http.Transport{}
	httpTransport.Dial = dialer.Dial
	httpClient := &http.Client{Transport: httpTransport}
	httpClient.Timeout = 5 * time.Second
	return httpClient.Get(webUrl)
}

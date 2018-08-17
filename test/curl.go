package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"golang.org/x/net/proxy"
)

func init1() {

	for _, arg := range os.Args {
		fmt.Println(arg)
	}

	if len(os.Args) != 2 {
		fmt.Println("params count must be 2!")
		os.Exit(-1)
	}
}

func main() {
	timeBegin := time.Now()
	r, err := http.Get(os.Args[1])
	if err != nil {
		fmt.Println("xx:" + err.Error())
		return
	}
	timeEnd := time.Now()
	diffTime := (timeEnd.UnixNano() - timeBegin.UnixNano()) / 1000
	fmt.Println(strconv.FormatInt(diffTime, 10))

	io.Copy(os.Stdout, r.Body)
	if err := r.Body.Close(); err != nil {
		fmt.Println(err)
	}
}
func proxyGet(webUrl string) (*http.Response, error) {
	/*
		1. 代理请求
		2. 跳过https不安全验证
	*/
	// webUrl := "http://ip.gs/"

	//url代理
	// var proxyUrl string
	// proxyUrl = "http://ubuntu.urwork.qbtrade.org:1080"
	// proxy, _ := url.Parse(proxyUrl)
	// tr := &http.Transport{
	// 	Proxy:           http.ProxyURL(proxy),
	// 	TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	// }
	// client := &http.Client{
	// 	Transport: tr,
	// 	Timeout:   time.Second * 5, //超时时间
	// }
	// return client.Get(webUrl)

	//socks5代理
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

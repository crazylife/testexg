package okef

import (
	"crypto/md5"
	"encoding/hex"
	"io/ioutil"
	"math"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/buger/jsonparser"
)

const (
	RestApiRoot        = "https://www.okex.com"
	OKEXRequestTimeout = 10 * time.Second

	WSRootAPI              = "wss://real.okex.com:10440/websocket/okexapi"
	UDSChannleBufferLength = 10
)

func Sign(params string, secretKey string) string {
	data := []byte(params + "&secret_key=" + secretKey)
	m := md5.New()
	m.Write(data)
	sign := strings.ToUpper(hex.EncodeToString(m.Sum(nil)))
	return sign
}

func IsEqual(v1 float64, v2 float64, abs_tol float64) bool {
	return math.Abs(v1-v2) < abs_tol
}

func exgRespHandle(response *http.Response, err error) ([]byte, int, string) {
	if err != nil {
		switch err := err.(type) {
		case net.Error:
			if err.Timeout() {
				return nil, -2, "okef timeout:" + err.Error()
			}
		}
		return nil, -3, err.Error()
	}

	if response.StatusCode == 403 {
		return nil, -5, "okef http 403:ip disable"
	}

	rspData, err := ioutil.ReadAll(response.Body)
	defer response.Body.Close()
	if err != nil {
		return nil, -1, err.Error()
	}

	bResult, err := jsonparser.GetBoolean(rspData, "result")
	if err != nil {
		return nil, -4, "okef result not found"
	}

	//exg返回error
	if !bResult {
		errorCode, err := jsonparser.GetInt(rspData, "error_code")
		if err != nil {
			return nil, -4, "okef error_code not found"
		}
		return nil, int(errorCode), ""
	}

	//exg正确返回
	return rspData, 0, ""
}

//yangq:设置timeout，判断timeout
//第二个参数返回值：errNumber 0：正确，-1：一般错误，-2timeout，-3：tcp被exg拒绝等？这种网络错误是报在这层吗？,-4:交易所返回数据格式错误，-5：用户请求过快，IP被屏蔽
func get(url string) ([]byte, int, string) {
	client := http.Client{
		Timeout: OKEXRequestTimeout,
	}
	response, err := client.Get(url)
	return exgRespHandle(response, err)
}

//第二个参数返回值：errNumber 0：正确，-1：一般错误，-2timeout，-3：tcp被exg拒绝等？这种网络错误是报在这层吗？,-4:交易所返回数据格式错误
func post(urlX, params string) ([]byte, int, string) {
	proxyUrl, err := url.Parse("http://127.0.0.1:1080")
	if err != nil {
		return nil, -1, "proxy error"
	}

	client := http.Client{
		Timeout:   OKEXRequestTimeout,
		Transport: &http.Transport{Proxy: http.ProxyURL(proxyUrl)},
	}
	response, err := client.Post(urlX, "application/x-www-form-urlencoded", strings.NewReader(params))
	return exgRespHandle(response, err)
}

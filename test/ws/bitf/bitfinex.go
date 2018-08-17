package bitfinex

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/buger/jsonparser"

	"github.com/gorilla/websocket"
)

const (
	WSRootAPI              = "wss://api.bitfinex.com/ws/"
	UDSChannleBufferLength = 10
)

type TradeInfo struct {
	Price     float64
	Amount    float64
	Type      string
	Timestamp int64
}
type WSConn struct {
	Ctx    context.Context
	Cancel context.CancelFunc
}
type IBifinexService interface {
	CreateWS(string, string) (chan *TradeInfo, chan string, int, string)
	Disconn()
}

func NewBifinexWSService() IBifinexService {
	ctx, cancelFun := context.WithCancel(context.Background())
	return &WSConn{Ctx: ctx, Cancel: cancelFun}
}

//contractType:this_week,next_week,quarter
// 0：正确，-2：timeout，-3：network error
func (ws *WSConn) CreateWS(xtc string, contractType string) (chan *TradeInfo, chan string, int, string) {
	wsurl := WSRootAPI
	goU, _ := url.Parse("http://127.0.0.1:1080")
	dialer := websocket.Dialer{
		Proxy: http.ProxyURL(goU),
	}
	c, _, err := dialer.Dial(wsurl, nil)
	if err != nil {
		switch err := err.(type) {
		case net.Error:
			if err.Timeout() {
				return nil, nil, -2, "timeout:" + err.Error()
			}
		}
		return nil, nil, -3, err.Error()
	}

	format := "{\"event\":\"subscribe\",\"channel\": \"trades\",\"pair\": \"BTCUSD\"}"
	format = strings.Replace(format, "X", xtc, -1)
	format = strings.Replace(format, "Y", contractType, -1)
	err = c.WriteMessage(websocket.TextMessage, []byte(format))
	if err != nil {
		return nil, nil, -1, "WriteMessage Error : " + err.Error()
	}

	aech := make(chan *TradeInfo, UDSChannleBufferLength)
	errCh := make(chan string)

	//发送ping心跳，维持连接
	//ping属于ws自己维持的，和ws接收信息是同步的，故这个模块自己维护
	wsConnCh := make(chan bool)
	go func() {
		ticker := time.NewTicker(60 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if err := c.WriteMessage(websocket.PingMessage, []byte("")); err != nil {
					//yangq:可以关掉连接吗？
					fmt.Println(err.Error())
				}
				break
			//ws关闭，结束心跳
			case <-wsConnCh:
				return
			}
		}
	}()

	var activeClose bool
	go func() {
		select {
		case <-ws.Ctx.Done():
			fmt.Println("conn ctx done!!")
			activeClose = true
			close(wsConnCh)
			c.Close()
		}
	}()

	go func() {
		defer close(errCh)
		count := 0
		for {
			//这边是阻塞的，不会触发上面的case
			_, message, err := c.ReadMessage()
			if err != nil {
				if !activeClose {
					errCh <- ("wsRead:" + err.Error())
				}
				return
			}

			count++
			_, err1 := jsonparser.GetString(message, "event")
			if err1 == nil {
				continue
			}
			if count == 3 {
				continue
			}
			// 一：{"event":"info","version":1.1,"serverId":"98545572-f934-4113-a234-8377e027da5c","platform":{"status":1}}
			// 二：{"event":"subscribed","channel":"trades","chanId":2,"pair":"BTCUSD"}
			// seq,id, timestamp,price,amount
			// [10,"tu","3884763-BTCUSD",280319354,1534251475,6106,0.041]
			// [10,"te","3884764-BTCUSD",1534251475,6105.98223992,-0.04130451]
			// [10,"tu","3884764-BTCUSD",280319368,1534251475,6105.98223992,-0.04130451]
			// [10,"hb"]

			data := make([]interface{}, 0)
			err = json.Unmarshal(message, &data)
			if err != nil {
				errCh <- "json.Unmarshal(message, data):" + err.Error()
				return
			}

			uda := TradeInfo{}
			if len(data) == 6 {
				var ok1, ok2, ok3 bool
				var fTime float64
				uda.Amount, ok1 = data[5].(float64)
				uda.Price, ok2 = data[4].(float64)
				fTime, ok3 = data[3].(float64)
				uda.Timestamp = int64(fTime)
				if !ok1 || !ok2 || !ok3 {
					errCh <- "5 !ok1|| !ok2||!ok3:" + err.Error()
					return
				}
				uda.Type = "buy"
				if uda.Amount < 0 {
					uda.Amount = 0 - uda.Amount
					uda.Type = "sell"
				}
				aech <- &uda
				continue
			}

			if len(data) == 7 {
				var ok1, ok2, ok3 bool
				var fTime float64
				uda.Amount, ok1 = data[6].(float64)
				uda.Price, ok2 = data[5].(float64)
				fTime, ok3 = data[4].(float64)
				uda.Timestamp = int64(fTime)
				uda.Type = "buy"
				if !ok1 || !ok2 || !ok3 {
					errCh <- "5 !ok1|| !ok2||!ok3:" + err.Error()
					return
				}
				if uda.Amount < 0 {
					uda.Amount = 0 - uda.Amount
					uda.Type = "sell"
				}
				aech <- &uda
				continue
			}
		}
	}()

	return aech, errCh, 0, ""

}

func (ws *WSConn) Disconn() {
	ws.Cancel()
}

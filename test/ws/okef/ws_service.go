package okef

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/buger/jsonparser"
	"github.com/gorilla/websocket"
)

type IOkefService interface {
	CreateWS(string, string, string, string) (chan string, int, string, *DataChans)
	Disconn()
}

func NewOkefWSService() IOkefService {
	ctx, cancelFun := context.WithCancel(context.Background())
	return &wsConn{Ctx: ctx, Cancel: cancelFun}
}

type wsConn struct {
	Ctx    context.Context
	Cancel context.CancelFunc
}

//contractType:this_week,next_week,quarter
// 0：正确，-2：timeout，-3：network error
func (ws *wsConn) CreateWS(xtc string, contractType string, apiKey string, secretKey string) (chan string, int, string, *DataChans) {
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
				return nil, -2, "timeout:" + err.Error(), nil
			}
		}
		return nil, -3, err.Error(), nil
	}

	sMsg, datasChan := CreateChannels(apiKey, secretKey)
	fmt.Println(sMsg)
	err = c.WriteMessage(websocket.TextMessage, []byte(sMsg))
	if err != nil {
		return nil, -1, "WriteMessage Error : " + err.Error(), nil
	}

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

		for {
			//这边是阻塞的，不会触发上面的case
			_, message, err := c.ReadMessage()
			if err != nil {
				if !activeClose {
					errCh <- ("wsRead:" + err.Error())
				}
				return
			}

			//write 返://[{"binary":0,"channel":"addChannel","data":{"result":true,"channel":"ok_sub_futureusd_btc_depth_this_week"}}]
			//{"event":"info","version":1.1,"serverId":"ba5f62eb-4703-44bd-8550-e47a55c1ab9e","platform":{"status":1}}---这是什么鬼

			// 	[
			//     {
			//         "data":
			//         "channel": "ok_sub_futureusd_btc_depth_this_week_20"
			//     }
			// ]

			isOK := true
			_, err = jsonparser.ArrayEach(message, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
				if !isOK {
					return
				}

				channle, err := jsonparser.GetString(value, "channel")
				if err != nil {
					isOK = false
					errCh <- "jsonparser.GetString(value, \"channel\")----" + string(value) + "----" + err.Error()
					return
				}

				data, _, _, err := jsonparser.Get(value, "data")
				if err != nil {
					isOK = false
					errCh <- "jsonparser.GetString(value, \"channel\")----" + string(value) + "----" + err.Error()
					return
				}

				err = subDataHandler(channle, data)
				if err != nil {
					isOK = false
					errCh <- "subDataHandler error : ----" + channle + "----" + string(data) + err.Error()
					return
				}
			})

			//结束循环！
			if !isOK {
				return
			}
		}
	}()
	return errCh, 0, "", &datasChan

}

func (ws *wsConn) Disconn() {
	ws.Cancel()
}

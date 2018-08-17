package main

import (
	"encoding/json"
	"fmt"
	"time"

	bitfinex "test.yang/test/ws/bitf"
	okef "test.yang/test/ws/okef"
)

const (
	apiKey    = "73c5fd7b-6008-4cef-ac3c-cce45de458b8"
	secretKey = "DF745ECE612D9B65E9B75E4C52BD7FFC"
)

func main() {

	//初始化数据库
	err := sqlCreate()
	if err != nil {
		panic(err.Error())
	}

	var cmd string
	for {
		fmt.Scanln(&cmd)
		if cmd == "exit" {
			break
		}
	}
	fmt.Println("3秒后退出程序")
	time.Sleep(3 * time.Second)
}

func beginMonitor() {
	//ok 行情数据
	okefIns := okef.NewOkefWSService()
	abErrCh, code, msg, chans := okefIns.CreateWS("btc", "quarter", apiKey, secretKey)
	if code != 0 {
		panic(msg)
	}
	//bifinex行情数据
	bif := bitfinex.NewBifinexWSService()
	bifCh, bifErrCh, code, msg := bif.CreateWS("btc", "this_week")
	if code != 0 {
		panic(msg)
	}

	go func() {
		for {
			select {
			case errMsg := <-abErrCh:
				okefIns.Disconn()
				bif.Disconn()
				fmt.Println(errMsg)
				//重启程序
				//go beginMonitor()
				return
			case md, ok := <-chans.MD:
				//yangq:需要return?
				if !ok {
					time.Sleep(1 * time.Second)
					break
				}
				go func() {
					//asks数据是按照价格由高到低排的，bids也是
					bs, _ := json.Marshal(md)
					insert(md.Timestamp, md.Asks[len(md.Asks)-1].Price, md.Bids[0].Price, string(bs))
				}()
				break
			case tradeInfo, ok := <-chans.Trade:
				if !ok {
					time.Sleep(1 * time.Second)
					break
				}
				go func() {
					for _, t := range *tradeInfo {
						insertOkefTrade(t.ID, t.Price, t.Amount, t.Time, t.Type)
					}
				}()
			case userTrade, ok := <-chans.UserTrade:
				if !ok {
					time.Sleep(1 * time.Second)
					break
				}
				fmt.Println(userTrade)
				break
			case userPos, ok := <-chans.UserPos:
				if !ok {
					time.Sleep(1 * time.Second)
					break
				}
				fmt.Println(userPos)
				break
			//////////bifinex
			case bifTradeInfo, ok := <-bifCh:
				if !ok {
					return
				}
				go func() {
					bitfInsert(bifTradeInfo.Price, bifTradeInfo.Amount, bifTradeInfo.Type, bifTradeInfo.Timestamp)
				}()
				break
			case errMsg := <-bifErrCh:
				bif.Disconn()
				okefIns.Disconn()
				fmt.Println(errMsg)
				return
			}
		}

	}()
}

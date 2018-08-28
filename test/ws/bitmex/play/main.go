package main

import (
	"fmt"
	"time"

	"test.yang/test/ws/bitmex"
)

func main() {
	//初始化数据库
	err := sqlCreate()
	if err != nil {
		panic(err.Error())
	}
	start := 500
	for {
		datas, err := bitmex.GetHistoryTradeData("2018-07-01 00:00", 500, "XBTUSD", start)
		if err != nil {
			fmt.Println(err.Error())
			break
		}
		for _, d := range *datas {
			err := insertBitmexTrade(d.Timestamp, d.Symbol, d.BS, float64(d.Size), float64(d.Price), d.TickDirection, d.TrdMatchID,
				float64(d.GrossValue), d.HomeNotional, float64(d.ForeignNotional))
			if err != nil {
				break
			}
		}
		start += 500
		time.Sleep(time.Second)
	}
	fmt.Println("GAME OVER!")
	var s string
	fmt.Sscanln(s)
	fmt.Println(s)
}

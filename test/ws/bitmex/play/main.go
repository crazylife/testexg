package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"test.yang/test/ws/bitmex"
	"test.yang/test/ws/bitmex/play/xx"
)

func main() {

	if len(os.Args) < 2 {
		fmt.Println("args length is smaller than 2")
		return
	}

	start, err := strconv.Atoi(os.Args[1])
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	//初始化数据库
	err = sqlCreate()
	if err != nil {
		panic(err.Error())
	}
	//start := 0
	for {
		datas, err := bitmex.GetHistoryTradeData("2018-07-16 09:43", 500, "XBTUSD", start)
		if err != nil {
			fmt.Println(err.Error())
			time.Sleep(10 * time.Second)
			continue
		}
		for _, d := range *datas {
			err := insertBitmexTrade(d.Timestamp, d.Symbol, d.BS, float64(d.Size), float64(d.Price), d.TickDirection, d.TrdMatchID,
				float64(d.GrossValue), d.HomeNotional, float64(d.ForeignNotional))
			if err != nil {
				fmt.Println(err.Error())
				break
			}
		}
		start += 500
		time.Sleep(2 * time.Second)
	}
	fmt.Println("GAME OVER!")
	var s string
	fmt.Sscanln(s)
	fmt.Println(s)
}

func test() {
	aa := xx.NewRequest()
	aa.B = "test"
	fmt.Println(aa.A + aa.B)
}

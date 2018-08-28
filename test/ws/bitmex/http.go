package bitmex

import (
	"encoding/json"
	"errors"
	"fmt"
	neturl "net/url"
	"strconv"
)

type TradeData struct {
	Timestamp       string  `json:"timestamp"`
	Symbol          string  `json:"symbol"`
	BS              string  `json:"side"`
	Size            int64   `json:"size"`
	Price           float64 `json:"price"`
	TickDirection   string  `json:"tickDirection"`
	TrdMatchID      string  `json:"trdMatchID"`
	GrossValue      int64   `json:"grossValue"`
	HomeNotional    float64 `json:"homeNotional"`
	ForeignNotional int64   `json:"foreignNotional"`
}

func GetHistoryTradeData(time string, count int, symbol string, start int) (*[]TradeData, error) {
	urlPath := "count=" + strconv.Itoa(count) + "&symbol=" + symbol + "&reverse=false&startTime=" + neturl.QueryEscape(time) + "&start=" + strconv.Itoa(start)
	//urlPath = neturl.QueryEscape(urlPath)
	url := "https://www.bitmex.com/api/v1/trade?" + urlPath

	b, errCode, errMsg := get(url)
	if errCode != 0 {
		fmt.Println(errMsg)
		return nil, errors.New(errMsg)
	}
	fmt.Println(string(b))

	data := make([]TradeData, 0)

	err := json.Unmarshal(b, &data)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	return &data, nil
}

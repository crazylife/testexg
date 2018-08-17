package okef

import (
	"strconv"

	"github.com/buger/jsonparser"
)

type MarketFutureTrade struct {
	Symbol       string `json "symbol"`
	ContractType string `json "contract_type"`
	Apikey       string `json "api_key"`
	Sign         string `json "sign"`
	Price        string `json "price"`
	Amount       string `json "amount"`
	Type         string `json "type"`
	//Match_price   string `json "match_price"`
	//Lever_rate    string `json "lever_rate"`
}

func (thisObject *MarketFutureTrade) SignParams(secretKey string) string {
	signStr := "amount=" + thisObject.Amount + "&api_key=" + thisObject.Apikey + "&contract_type=" + thisObject.ContractType +
		"&price=" + thisObject.Price + "&symbol=" + thisObject.Symbol + "&type=" + thisObject.Type
	return Sign(signStr, secretKey)
}

func (thisObject *MarketFutureTrade) GetOkefResponse() (string, int, string) {
	url := RestApiRoot + "/api/v1/future_trade.do"
	params := "amount=" + thisObject.Amount + "&api_key=" + thisObject.Apikey + "&contract_type=" + thisObject.ContractType +
		"&price=" + thisObject.Price + "&symbol=" + thisObject.Symbol + "&type=" + thisObject.Type + "&sign=" + thisObject.Sign

	rspData, errorCode, errorInfo := post(url, params)
	if errorCode != 0 {
		return "", errorCode, errorInfo
	}

	orderID, err := jsonparser.GetInt(rspData, "order_id")
	if err != nil {
		return "", -4, err.Error()
	}
	return strconv.FormatInt(orderID, 10), 0, ""
}

func PostOrder(bs int, price float64, amount float64, apiKey string, secretKey string) (string, int, string) {
	mft := MarketFutureTrade{}
	mft.Symbol = "btc_usd"
	mft.ContractType = "quarter"
	mft.Type = strconv.Itoa(bs)
	mft.Price = strconv.FormatFloat(price, 'f', -1, 64)
	mft.Amount = strconv.FormatFloat(amount, 'f', -1, 64)
	mft.Apikey = apiKey
	mft.Sign = mft.SignParams(secretKey)
	return mft.GetOkefResponse()
}

package okef

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/buger/jsonparser"
)

type handler func(val []byte) error

var (
	channels = map[string]handler{
		//市场深度
		"ok_sub_futureusd_btc_depth_quarter_20": publicSubFutureusdBTCDepthQuarter,
		//交易信息
		"ok_sub_futureusd_btc_trade_quarter": publicSubFutureusdBTCTradeQuarter,

		////个人ws数据流
		//交易信息
		"ok_sub_futureusd_trades": personSubFutureusdTrades,
		//合约账户信息
		//"ok_sub_futureusd_userinfo": "",
		//合约持仓信息
		"ok_sub_futureusd_positions": personSubFutureusdPositions,

		"addChannel": publicSubResponse,
		"login":      publicSubResponse,
	}

	//传送数据的chann
	datasChan = DataChans{}
)

// CreateChannels 返回请求链接的信息，并返回接收数据的通道
// 特意复制一个datachan出来，但是没有关闭接口！！！——让go自动回收，会有问题吗？
func CreateChannels(apiKey string, secretKey string) (string, DataChans) {
	type wsAddEvent struct {
		Event     string `json:"event"`
		Channel   string `json:"channel"`
		Paramters struct {
			APIKey string `json:"api_key"`
			Sign   string `json:"sign"`
		} `json:"paramters"`
	}

	event := wsAddEvent{Event: "login"}
	event.Paramters.APIKey = apiKey
	event.Paramters.Sign = Sign("api_key="+apiKey, secretKey)

	events := []wsAddEvent{
		{Event: "addChannel", Channel: "ok_sub_futureusd_btc_depth_quarter_20"},
		{Event: "addChannel", Channel: "ok_sub_futureusd_btc_trade_quarter"},
		event,
	}

	b, _ := json.Marshal(events)

	datasChan.MD = make(chan *MarketDepth, 100)
	datasChan.Trade = make(chan *[]TradeInfo, 100)
	datasChan.UserTrade = make(chan *UserTradeInfo, 100)
	datasChan.UserPos = make(chan *UserPosition, 100)

	return string(b), datasChan
}

func subDataHandler(channel string, data []byte) error {
	handler, ok := channels[channel]
	if !ok {
		fmt.Println("not support channel----" + channel + "----" + string(data))
		return nil
	}
	return handler(data)
}

func publicSubResponse(val []byte) error {
	fmt.Println(string(val))
	ok, err := jsonparser.GetBoolean(val, "result")
	if err != nil {
		return errors.New("jsonparser.GetBoolean(data, \"result\")----" + string(val) + "----" + err.Error())
	}
	if !ok {
		return errors.New("get boolean is false----" + string(val))
	}
	return nil
}

//统一返回的数据格式为：[{"channel":"channel","success":"","errorcode":"","data":{}}, {"channel":"channel","success":"","errorcode":1,"data":{}}]

func publicSubFutureusdBTCDepthQuarter(val []byte) error {
	//write 返://[{"binary":0,"channel":"addChannel","data":{"result":true,"channel":"ok_sub_futureusd_btc_depth_this_week"}}]

	// 	[
	//     {
	//         "data": {
	//             "timestamp": 1490337551299,
	//             "asks": [
	//                 [
	//                     "996.72",
	//                     "20.0",
	//                     "2.0065",
	//                     "85.654",
	//                     "852.0"
	//                 ]
	//             ],
	//             "bids": [
	//                 [
	//                     "991.67",
	//                     "6.0",
	//                     "0.605",
	//                     "0.605",
	//                     "6.0"
	//                 ]
	//         },
	//         "channel": "ok_sub_futureusd_btc_depth_this_week_20"
	//     }
	// ]

	data := &struct {
		Timestamp int64       `json:"timestamp"`
		Asks      [][]float64 `json:"asks"`
		Bids      [][]float64 `json:"bids"`
		Channel   string      `json:"channel"`
	}{}

	err := json.Unmarshal(val, data)
	if err != nil {
		return errors.New("json.Unmarshal(value, data):" + err.Error())
	}
	var md MarketDepth
	md.Timestamp = data.Timestamp
	md.Asks = make([]Ticker, 0, 20)
	for _, ask := range data.Asks {
		//数量为0的跳过
		if ask[1] < 1e-8 {
			continue
		}
		md.Asks = append(md.Asks, Ticker{Price: ask[0], Amount: ask[1], XtcPrice: ask[2], SumXtcPrice: ask[3], SumAmount: ask[4]})
	}
	for _, bid := range data.Bids {
		//数量为0的跳过
		if bid[1] < 1e-8 {
			continue
		}
		md.Bids = append(md.Bids, Ticker{Price: bid[0], Amount: bid[1], XtcPrice: bid[2], SumXtcPrice: bid[3], SumAmount: bid[4]})
	}

	if len(md.Asks) == 0 || len(md.Bids) == 0 {
		return nil
	}

	datasChan.MD <- &md
	return nil
}

func publicSubFutureusdBTCTradeQuarter(val []byte) error {
	// 			# Response
	// [
	//     {
	//         "data": [
	//             [
	//                 "732916869",
	//                 "999.49",
	//                 "12.0",
	//                 "15:25:03",
	//                 "ask",
	//                 "1.2006"
	//             ],
	//             [
	//                 "732916871",
	//                 "999.49",
	//                 "2.0",
	//                 "15:25:03",
	//                 "ask",
	//                 "0.2001"
	//             ],
	//             [
	//                 "732916899",
	//                 "999.49",
	//                 "2.0",
	//                 "15:25:04",
	//                 "ask",
	//                 "0.2001"
	//             ]
	//         ]
	//         "channel": "ok_sub_futureusd_btc_trade_this_week"
	//     }
	// ]

	data := make([][]interface{}, 0)

	//yangq:err := json.Unmarshal(val, data)----少了个&就报错。。。。
	err := json.Unmarshal(val, &data)
	if err != nil {
		return errors.New("json.Unmarshal(value, data):" + err.Error())

	}

	tradeInfos := make([]TradeInfo, 0)
	for _, t := range data {
		var tradeInfo TradeInfo

		var err1, err2 error
		tradeInfo.ID = t[0].(string)
		tradeInfo.Price, err1 = strconv.ParseFloat(t[1].(string), 64)
		tradeInfo.Amount, err2 = strconv.ParseFloat(t[2].(string), 64)
		if err1 != nil || err2 != nil {
			return errors.New("strconv.ParseFloat:" + t[1].(string) + "!!!" + t[2].(string))
		}
		tradeInfo.Time = t[3].(string)
		tradeInfo.Type = t[4].(string)
		//tradeInfo.XtcAmount = t[4].(float64)

		tradeInfos = append(tradeInfos, tradeInfo)
	}

	datasChan.Trade <- &tradeInfos
	return nil

}

func personSubFutureusdTrades(val []byte) error {
	// 	[{
	//         "data":{
	//             amount:1
	//             contract_id:20170331115,
	//             contract_name:"LTC0331",
	//             contract_type:"this_week",
	//             create_date:1490583736324,
	//             create_date_str:"2017-03-27 11:02:16",
	//             deal_amount:0,
	//             fee:0,
	//             lever_rate:20,
	//             orderid:5058491146,
	//             price:0.145,
	//             price_avg:0,
	//             status:0,
	//             system_type:0,
	//             type:1,
	//             unit_amount:10,
	//             user_id:101
	//         },
	//         "channel":"ok_sub_futureusd_trades"
	//     }
	// ]

	// data := struct {
	// 	Amount        float64 `json:"amount`
	// 	ContractID    int64   `json:"contract_id`
	// 	ContractName  string  `json:"contract_name`
	// 	ContractType  string  `json:"contract_type`
	// 	CreateDate    int64   `json:"create_date`
	// 	CreateDateStr string  `json:"create_date_str`
	// 	DealAmount    float64 `json:"deal_amount`
	// 	Fee           float64 `json:"fee`
	// 	LeverRate     float64 `json:"lever_rate` //杠杆倍数  value:10/20  默认10
	// 	OrderID       int64   `json:"orderid`
	// 	Price         float64 `json:"price`
	// 	PriceAvg      float64 `json:"price_avg`
	// 	Status        int64   `json:"status`      //订单状态(0等待成交 1部分成交 2全部成交 -1撤单 4撤单处理中)
	// 	SystemType    int64   `json:"system_type` //订单类型 0:普通 1:交割 2:强平 4:全平 5:系统反单
	// 	Type          int64   `json:"type`        //订单类型 1：开多 2：开空 3：平多 4：平空
	// 	UnitAmount    float64 `json:"unit_amount`
	// 	UserID        int64   `json:"user_id`
	// }{}

	data := UserTradeInfo{}
	err := json.Unmarshal(val, data)
	if err != nil {
		return errors.New("json.Unmarshal(value, data):" + err.Error())
	}

	datasChan.UserTrade <- &data
	return nil
}

func personSubFutureusdUserinfo(val []byte) {

}

func personSubFutureusdPositions(val []byte) error {
	//逐仓返回
	// [{
	// 	"data":{
	// 		"positions":[
	// 			{
	// 				"position":"1",
	// 				"profitreal":"0.0",
	// 				"contract_name":"LTC0407",
	// 				"costprice":"0.0",
	// 				"bondfreez":"1.64942529",
	// 				"forcedprice":"0.0",
	// 				"avgprice":"0.0",
	// 				"lever_rate":10,
	// 				"fixmargin":0,
	// 				"contract_id":20170407135,
	// 				"balance":"0.0",
	// 				"position_id":27864057,
	// 				"eveningup":"0.0",
	// 				"hold_amount":"0.0"
	// 			},
	// 			{
	// 				"position":"2",
	// 				"profitreal":"0.0",
	// 				"contract_name":"LTC0407",
	// 				"costprice":"0.0",
	// 				"bondfreez":"1.64942529",
	// 				"forcedprice":"0.0",
	// 				"avgprice":"0.0",
	// 				"lever_rate":10,
	// 				"fixmargin":0,
	// 				"contract_id":20170407135,
	// 				"balance":"0.0",
	// 				"position_id":27864057,
	// 				"eveningup":"0.0",
	// 				"hold_amount":"0.0"
	// 			}
	// 		"symbol":"ltc_usd",
	// 		"user_id":101
	// 	}]

	// data := struct {
	// 	Symbol    string `json:"symbol"`
	// 	UserID    int64  `json:"user_id"`
	// 	Positions []struct {
	// 		Position     string  `json:"position"`
	// 		Profitreal   string  `json:"profitreal"` //收益
	// 		ContractName string  `json:"contract_name"`
	// 		ContractID   int64   `json:"contract_id"`
	// 		CostPrice    string  `json:"costprice"`   //开仓价格
	// 		BondFreez    string  `json:"bondfreez"`   //当前合约冻结保证金
	// 		ForcedPrice  string  `json:"forcedprice"` //强平价格
	// 		AvgPrice     string  `json:"avgprice"`    //开仓均价
	// 		LeverRate    float64 `json:"lever_rate"`
	// 		FixMargin    float64 `json:"fixmargin"` //固定保证金
	// 		Balance      string  `json:"balance"`   //合约账余额
	// 		PositionID   int64   `json:"position_id"`
	// 		EveningUp    string  `json:"eveningup"`   //可平仓量
	// 		HoldAmount   string  `json:"hold_amount"` //持仓量
	// 	} `json:"positions"`
	// }{}

	data := UserPosition{}
	err := json.Unmarshal(val, data)
	if err != nil {
		return errors.New("json.Unmarshal(value, data):" + err.Error())
	}
	datasChan.UserPos <- &data
	return nil
}

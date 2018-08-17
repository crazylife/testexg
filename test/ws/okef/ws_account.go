package okef

// DataChans 对外提供的数据
type DataChans struct {
	MD        chan *MarketDepth
	Trade     chan *[]TradeInfo
	UserTrade chan *UserTradeInfo
	UserPos   chan *UserPosition
}

type Ticker struct {
	Price       float64
	Amount      float64
	XtcPrice    float64
	SumXtcPrice float64
	SumAmount   float64
}
type MarketDepth struct {
	Timestamp int64
	Asks      []Ticker
	Bids      []Ticker
}

//交易数据
type TradeInfo struct {
	ID     string
	Price  float64
	Amount float64 //成交量：张
	Time   string
	Type   string
	//XtcAmount float64 //成交量：币
}

type UserTradeInfo struct {
	Amount        float64 `json:"amount`
	ContractID    int64   `json:"contract_id`
	ContractName  string  `json:"contract_name`
	ContractType  string  `json:"contract_type`
	CreateDate    int64   `json:"create_date`
	CreateDateStr string  `json:"create_date_str`
	DealAmount    float64 `json:"deal_amount`
	Fee           float64 `json:"fee`
	LeverRate     float64 `json:"lever_rate` //杠杆倍数  value:10/20  默认10
	OrderID       int64   `json:"orderid`
	Price         float64 `json:"price`
	PriceAvg      float64 `json:"price_avg`
	Status        int64   `json:"status`      //订单状态(0等待成交 1部分成交 2全部成交 -1撤单 4撤单处理中)
	SystemType    int64   `json:"system_type` //订单类型 0:普通 1:交割 2:强平 4:全平 5:系统反单
	Type          int64   `json:"type`        //订单类型 1：开多 2：开空 3：平多 4：平空
	UnitAmount    float64 `json:"unit_amount`
	UserID        int64   `json:"user_id`
}

type UserPosition struct {
	Symbol    string `json:"symbol"`
	UserID    int64  `json:"user_id"`
	Positions []struct {
		Position     string  `json:"position"`
		Profitreal   string  `json:"profitreal"` //收益
		ContractName string  `json:"contract_name"`
		ContractID   int64   `json:"contract_id"`
		CostPrice    string  `json:"costprice"`   //开仓价格
		BondFreez    string  `json:"bondfreez"`   //当前合约冻结保证金
		ForcedPrice  string  `json:"forcedprice"` //强平价格
		AvgPrice     string  `json:"avgprice"`    //开仓均价
		LeverRate    float64 `json:"lever_rate"`
		FixMargin    float64 `json:"fixmargin"` //固定保证金
		Balance      string  `json:"balance"`   //合约账余额
		PositionID   int64   `json:"position_id"`
		EveningUp    string  `json:"eveningup"`   //可平仓量
		HoldAmount   string  `json:"hold_amount"` //持仓量
	} `json:"positions"`
}

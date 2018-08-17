package qb

import (
	"backend-go/config"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"time"

	"github.com/DataDog/datadog-go/statsd"
)

type Pankou struct {
	Price  float64 `json:"price"`
	Volume float64 `json:"volume"`
}
type Tick struct {
	Bids     []Pankou
	Asks     []Pankou
	Time     time.Time
	Contract string
	Last     float64
	Volume   float64
	Amount   float64
	Source   string
}

func NewTickV2(str string) *Tick {
	var data Tick
	json.Unmarshal([]byte(str), &data)
	if data.Bids == nil {
		data.Bids = make([]Pankou, 0)
	}
	if data.Asks == nil {
		data.Asks = make([]Pankou, 0)
	}
	return &data
}

type ResponseInfluxdb struct {
	Database    string            `json:"database"`
	Measurement string            `json:"measurement"`
	Tags        map[string]string `json:"tags"`
	Fields      map[string]Number `json:"fields"`
	Time        time.Time         `json:"time"`
}

type ResponseRedisTime struct {
	time.Time
}

func (n ResponseRedisTime) MarshalJSON() ([]byte, error) {
	// 强制只能到 10-6 不能输出10-9 不然python的arrow库会解析错误

	return []byte(`"` + n.Format("2006-01-02T15:04:05.999999Z07:00") + `"`), nil
}

type ResponseRedis struct {
	Bids     []Pankou          `json:"bids"`
	Asks     []Pankou          `json:"asks"`
	Time     ResponseRedisTime `json:"time"`
	Contract string            `json:"contract"`
	Last     float64           `json:"last"`
	Volume   float64           `json:"volume"`
	Source   string            `json:"source"`
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

type Number float64

func (n Number) MarshalJSON() ([]byte, error) {
	if 1e-10 < math.Abs(float64(n)) && math.Abs(float64(n)) < 0.001 {
		return []byte(fmt.Sprintf("%.10f", n)), nil
	}
	return []byte(fmt.Sprintf("%.8f", n)), nil
}

func (t *Tick) RedisString() string {
	r := ResponseRedis{
		Contract: t.Contract,
		Bids:     t.Bids,
		Asks:     t.Asks,
		Volume:   t.Volume,
		Time:     ResponseRedisTime{t.Time},
		Last:     t.Last,
		Source:   t.Source,
	}
	val, err := json.Marshal(r)
	if err != nil {
		panic(err)
	}
	return string(val)
}

func (t *Tick) InfluxdbString() string {
	tags := make(map[string]string)
	tags["contract"] = t.Contract
	fields := make(map[string]Number)
	for i := 0; i < min(2, len(t.Asks)); i++ {
		y := fmt.Sprintf("ask%d", i+1)
		fields[y] = Number(t.Asks[i].Price)
		fields[y+"v"] = Number(t.Asks[i].Volume)
	}
	for i := 0; i < min(2, len(t.Bids)); i++ {
		y := fmt.Sprintf("bid%d", i+1)
		fields[y] = Number(t.Bids[i].Price)
		fields[y+"v"] = Number(t.Bids[i].Volume)
	}
	fields["last"] = Number(t.Last)
	fields["volume"] = Number(t.Volume)
	r := ResponseInfluxdb{
		Database:    "quote",
		Measurement: "tick",
		Tags:        tags,
		Fields:      fields,
		Time:        t.Time,
	}
	val, _ := json.Marshal(r)
	return string(val)
}
func Scheme() string {
	if IsUrwork() {
		return "https"
	}
	return "http"
}

func PanicIf(err error) {
	if err != nil {
		panic(err)
	}
}

func IsProd() bool {
	prod := os.Getenv("QB_ENV")
	if prod == "prod" {
		return true
	}
	return false
}

func IsUrwork() bool {
	region := os.Getenv("QB_REGION")
	if region == "urwork" {
		return true
	}
	return false
}

var Statsd, _ = statsd.New(config.Statsd)

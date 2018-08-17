package main

import (
	"time"

	"TEST.YANG/test/okex"
	qb "TEST.YANG/test/util"
)

func main() {

	okex.Ex_Spot_bb_get_ticker("ltc_btc")
	for i := 0; i < 10; i++ {
		time.Sleep(time.Duration(1) * time.Second)
		qb.Statsd.Gauge("test-qy", float64(i), nil, 1)
	}

	// var c bytes.Buffer
	// c.Write([]byte("hello"))
	// fmt.Fprintf(&c, "@world@")
	// io.Copy(os.Stdout, &c)

	// array := [...]*int{0: new(int), 1: new(int)}
	// *array[0] = 10
	// *array[1] = 20

	// slice := []string{"0", "1", "2", "3", "4"}

	// fmt.Printf("sizeof:%d,int sizeof:%d", unsafe.Sizeof(slice))

	// for index, value := range slice {
	// 	fmt.Printf("Index: %d  Value: %s\n", index, value)
	// 	fmt.Printf("value: %d;value addr: %X;array value addr: %X;", value, &value, &slice[index])
	// }

	// tt.Ar1()

}

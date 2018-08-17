package array2

import (
	"fmt"
)

//slice是不是深copy
func Ar1() {

	slice := [][]int{{1}, {2, 3}}

	b := &slice
	fmt.Printf("before append----slice[0] addr:%p,&slice[0] addr:%p,&slice[0][0] addr:%X,slice[0] length:%d", slice[0], &slice[0], &slice[0][0], b)

	intSlice := []int{5}
	slice = append(slice, intSlice)

	fmt.Printf("after append----slice[0] addr:%p,&slice[0] addr:%p,slice[0][0] addr:%X", slice[0], &slice[0], &slice[0][0])

}

func ar2() {
	fmt.Printf("yituoshit %d", 100)
}

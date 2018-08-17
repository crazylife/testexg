package main

import (
	"fmt"
)

type user struct {
	name string
	addr string
}

type notifier interface {
	notify() int
}

func (u *user) notify() int {
	fmt.Println(u.name)
	return 0
}

func main2() {
	u := user{"123", "http"}
	fmt.Println("return vlaue : %d\n", sendNotification(&u))
}

func sendNotification(n notifier) int {
	return n.notify()
}

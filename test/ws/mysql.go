package main

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var (
	db *sql.DB
)

func sqlCreate() error {
	dbsql, err := sql.Open("mysql", "root:root@tcp(127.0.0.1:3306)/test?parseTime=true")
	if err != nil {
		return err
	}

	err = dbsql.Ping()
	if err != nil {
		return err
	}

	db = dbsql
	return nil
}

func insert(timestamp int64, ask1 float64, bid1 float64, content string) {
	sql := "insert into test.okex(content,ask1,bid1,timestamp,time) values(?,?,?,?,?)"
	_, err := db.Exec(sql, content, ask1, bid1, timestamp, time.Now().UnixNano()/1000000)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
}

func insertOkefTrade(tid string, price float64, amount float64, timestamp string, bs string) {
	sql := "insert into test.tradeinfo(tid,price,amount,timestamp,type,time) values(?,?,?,?,?,?)"
	_, err := db.Exec(sql, tid, price, amount, timestamp, bs, time.Now().UnixNano()/1000000)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
}

func bitfInsert(price float64, amount float64, bs string, timestamp int64) {
	sql := "insert into test.bitfinex(price,amount,type,timestamp,time) values(?,?,?,?,?)"
	_, err := db.Exec(sql, price, amount, bs, timestamp, time.Now().UnixNano()/1000000)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
}

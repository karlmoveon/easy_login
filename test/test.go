package main

import (
	"crypto/md5"
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"io"
	"strconv"
	"time"
)

var pictureUrl = "http://pic28.photophoto.cn/20130818/0020033143720852_b.jpg"
var localDsn = "root:@(127.0.0.1:3306)/go_task?charset=utf8"
var remoteDsn = "root:mysql.3306@(127.0.0.1:3306)/go_task?charset=utf8"

func insertTestData(start, end int) {

	fmt.Println("-- test insertTestData in")
	db, err := sql.Open("mysql", localDsn)
	if err != nil {
		fmt.Println("db open failed")
		fmt.Printf("db open failed: %s", err)
		return
	}
	defer db.Close()

	/* prepare参数绑定 */
	stmt, err := db.Prepare("insert into user (username, password, nickname, pictureurl) values (?, ?, ?, ?)")
	if err != nil {
		fmt.Println("prepare failed:%v", err)
		return
	}
	defer stmt.Close()

	begin := time.Now()
	for i := start; i < end; i++ {
		username := strconv.Itoa(i)
		password := strconv.Itoa(i)
		w := md5.New()
		io.WriteString(w, password+"karl")
		md5pwd := fmt.Sprintf("%x", w.Sum(nil))
		fmt.Println("origin pwd " + password + "md5 pwd " + md5pwd)
		nickname := strconv.Itoa(i)

		_, err := stmt.Exec(username, md5pwd, "nick_"+nickname, pictureUrl)
		if err != nil {
			fmt.Println("exec insert data error:%v", err)
		}

		if (i-start)%1000000 == 0 {
			elapsed := time.Since(begin)
			fmt.Println("already insert", i-start, "users")
			fmt.Println("inserted", i-start, "data into mysql cost", elapsed)
		}
	}
	cost := time.Since(begin)
	fmt.Println("ALL done. Insert", end-start, "data into mysql cost", cost)
}

func main() {
	fmt.Println("-- start insertTestData")
	insertTestData(1, 10000)
	fmt.Println("-- end insertTestData")
}

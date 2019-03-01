package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"
)

var address = "localhost:9090"
var nick = "testnick"
var pictureUrl = "http://wx4.sinaimg.cn/bmiddle/006oOWahly1fmn7ygvn5tj30b40b4q2x.jpg"

/* 计算QPS : 200个goroutine, chan */

func getTokenFromResp(resp string) string {

	cookie := strings.Split(resp, "usertoken=")[0]
	usertoken := strings.Split(cookie, ";")[0]
	return usertoken
}

func testLogin(conn net.Conn, username, password string) (string, int) {
	fmt.Println("test login in")
	var post string
	content := "username=" + username + "&password=" + password

	post += "POST /login HTTP/1.1\r\n"
	post += "Content-Type: application/x-www-form-urlencoded\r\n"
	post += "Content-Length: " + strconv.Itoa(len(content)) + "\r\n"
	post += "Connection: keep-alive\r\n"
	post += "Host: " + address + "\r\n"
	post += "Accept-Language: zh-CN,zh;q=0.8,en;q=0.6\r\n"
	post += "\r\n"
	post += content

	n, err := conn.Write([]byte(post))
	if err != nil {
		fmt.Println("write", n, "data failed")
	}
	defer conn.Close()

	respBuf := make([]byte, 1024)
	n, err = conn.Read(respBuf)

	resp := string(respBuf)
	/* resp为http resp，需要从中解出cookie，在后面使用 */
	usertoken := getTokenFromResp(resp)
	if len(usertoken) == 0 {
		fmt.Println("debug", username, " : login failed for resp", resp)
		return usertoken, -1
	}

	return usertoken, 0
}

func testNickname(conn net.Conn, usertoken, username, nickname string) int {
	fmt.Println("test nick in")
	var post string
	content := "username=" + username + "&nickname_submit=" + nickname

	post += "POST /editnickname HTTP/1.1\r\n"
	post += "Content-Type: application/x-www-form-urlencoded\r\n"
	post += "Content-Length: " + strconv.Itoa(len(content)) + "\r\n"
	post += "Connection: keep-alive\r\n"
	post += "Host: " + address + "\r\n"
	post += "Accept-Language: zh-CN,zh;q=0.8,en;q=0.6\r\n"
	post += "Cookie: usertoken=" + usertoken + "\r\n"
	post += "\r\n"
	post += content

	n, err := conn.Write([]byte(post))
	if err != nil {
		fmt.Println("write", n, "data failed")
	}

	respBuf := make([]byte, 1024)
	n, err = conn.Read(respBuf)

	resp := string(respBuf)
	fmt.Println("change nick resp:", resp)
	if !strings.Contains(resp, "302") {
		fmt.Println("debug :", username, " nick failed for resp", resp)
		return -1
	}

	return 0
}

func testPic(conn net.Conn, usertoken, username, pictureurl string) int {
	fmt.Println("test nick in")
	var post string
	content := "username=" + username + "&picture_submit=" + pictureurl

	post += "POST /uploadpic HTTP/1.1\r\n"
	post += "Content-Type: application/x-www-form-urlencoded\r\n"
	post += "Content-Length: " + strconv.Itoa(len(content)) + "\r\n"
	post += "Connection: keep-alive\r\n"
	post += "Host: " + address + "\r\n"
	post += "Accept-Language: zh-CN,zh;q=0.8,en;q=0.6\r\n"
	post += "Cookie: usertoken=" + usertoken + "\r\n"
	post += "\r\n"
	post += content

	n, err := conn.Write([]byte(post))
	if err != nil {
		fmt.Println("write", n, "data failed")
	}

	respBuf := make([]byte, 1024)
	n, err = conn.Read(respBuf)

	resp := string(respBuf)
	if !strings.Contains(resp, "302") {
		fmt.Println("debug :", username, " pic failed for resp", resp)
		return -1
	}

	return 0
}

func testServer(counter chan int, username, password string) {

	fmt.Println("test server in")
	conn, err := net.Dial("tcp", address)
	if err != nil {
		fmt.Println("testServer : Error dialing", err.Error())
	}

	_, res := testLogin(conn, username, password)
	fmt.Println(username, "testLogin return", res)

	//ret := testNickname(conn, usertoken, username, nick)
	//fmt.Println(username, "testNickname return", ret)
	//res |= ret
	//
	//ret = testPic(conn, usertoken, username, pictureUrl)
	//fmt.Println(username, "testPic return", ret)
	//res |= ret

	counter <- res
}

func main() {
	/* 模拟200个客户端 */
	var clients int
	flag.IntVar(&clients, "c", 20, "clients")
	flag.Parse()

	counter := make(chan int, clients)

	count, res, failed := 0, 0, 0

	begin := time.Now()
	var cost time.Duration
	go func() {
		for {
			select {
			case res = <-counter:
				count++
				if res != 0 {
					failed++
				}
				fmt.Println("res", res, "count", count, "failed", failed)

				if count == clients {
					cost = time.Since(begin)
					fmt.Println("all cost:", cost)
					outputResult(cost, clients, failed)
				}
			}
		}
	}()

	for i := 1; i <= clients; i++ {
		username := strconv.Itoa(i)
		password := strconv.Itoa(i)
		go testServer(counter, username, password)
	}

	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit
}

func outputResult(cost time.Duration, clients, failed int) {
	fmt.Println(clients, "clients cost time :", cost, "failed times:", failed)
}

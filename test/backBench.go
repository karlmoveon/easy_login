package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

func benchmarkBasicN(serverAddr string, n, c int32, isRan bool) (elapsed time.Duration) {
	readyGo := make(chan bool)
	var wg sync.WaitGroup
	wg.Add(int(c))

	remaining := n

	var transport http.RoundTripper = &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          int(c),
		MaxIdleConnsPerHost:   int(c),
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	cookieJar, err := cookiejar.New(nil)
	if err != nil {
		fmt.Println(err)
	}

	client := &http.Client{
		Transport: transport,
		Jar:       cookieJar,
	}

	cliRoutine := func(no int32) {
		for atomic.AddInt32(&remaining, -1) >= 0 {
			// continue
			data := url.Values{}

			var buffer bytes.Buffer
			// rand
			if isRan {
				buffer.WriteString(strconv.Itoa(rand.Intn(10000000)))
			} else {
				buffer.WriteString("1")
			}

			userNo := buffer.String()

			data.Set("username", userNo)
			data.Set("password", userNo)
			// req, err := http.NewRequest("GET", serverAddr, bytes.NewBufferString(data.Encode()))
			req, err := http.NewRequest("POST", serverAddr, bytes.NewBufferString(data.Encode()))
			//req.AddCookie(&http.Cookie{Name: "authToken", Value: username, Expires: time.Now().Add(120 * time.Second)})
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded") // This makes it work
			if err != nil {
				log.Println(err)
			}
			<-readyGo
			resp, err := client.Do(req)
			if err != nil {
				log.Println(err)
			}
			_, err1 := ioutil.ReadAll(resp.Body)
			if err1 != nil {
				log.Println(err1)
			}
			defer resp.Body.Close()
		}

		wg.Done()
	}

	for i := int32(0); i < c; i++ {
		go cliRoutine(i)
	}

	close(readyGo)
	start := time.Now()

	wg.Wait()

	return time.Since(start)
}

var num int64
var concurrency int64
var isRandom bool

func init() {
	flag.Int64Var(&num, "n", 2000, "num")
	flag.Int64Var(&concurrency, "c", 2000, "concurrency")
	flag.BoolVar(&isRandom, "r", false, "isRandom")
}

func main() {
	flag.Parse()
	elapsed := benchmarkBasicN("http://127.0.0.1:9090/login/", int32(num), int32(concurrency), isRandom)
	fmt.Println("HTTP server benchmark done:")
	fmt.Printf("\tTotal Requests(%v) - Concurrency(%v) - Cost(%s) - QPS(%v/sec)\n",
		num, concurrency, elapsed, math.Ceil(float64(num)/(float64(elapsed)/1000000000)))
}


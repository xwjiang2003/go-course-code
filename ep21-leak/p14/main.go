// 三个案例合一：跑完直接抓 pprof 分组表
//
//	go run ./p14
package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	_ "net/http/pprof"
	"runtime"
	"strings"
	"sync"
	"time"
)

// 案例一：无缓冲 channel，发送方没人接收
func leakSend(id int) {
	ch := make(chan int)
	go func() { ch <- id }()
}

// 案例二：ctx 没超时、没 cancel，Client 也没 Timeout
func leakCall(url string) {
	ctx, cancel := context.WithCancel(context.Background())
	_ = cancel
	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	io.Copy(io.Discard, resp.Body)
}

// 案例三：worker range 一个永不 close 的 channel
func leakWorkers(n int) {
	ch := make(chan int)
	for i := 0; i < n; i++ {
		go func() {
			for v := range ch {
				_ = v
			}
		}()
	}
	for i := 0; i < 5; i++ {
		ch <- i
	}
}

func main() {
	go http.ListenAndServe("127.0.0.1:16060", nil)
	fmt.Printf("T0 起动            = %d\n", runtime.NumGoroutine())

	hang := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))

	for i := 1; i <= 300; i++ {
		leakSend(i)
	}
	time.Sleep(300 * time.Millisecond)
	fmt.Printf("T1 案例一 x300     = %d\n", runtime.NumGoroutine())

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); leakCall(hang.URL) }()
	}
	time.Sleep(1500 * time.Millisecond)
	fmt.Printf("T2 案例二 x50      = %d\n", runtime.NumGoroutine())

	for r := 1; r <= 20; r++ {
		leakWorkers(10)
	}
	time.Sleep(500 * time.Millisecond)
	fmt.Printf("T3 案例三 x20x10   = %d\n", runtime.NumGoroutine())

	time.Sleep(2 * time.Second)
	fmt.Printf("T4 再等 2 秒        = %d  <- 不回落，这就是泄漏\n", runtime.NumGoroutine())

	fmt.Println()
	dump()
}

// dump 抓 /debug/pprof/goroutine?debug=1，只留"组头 + 最内层一帧"
func dump() {
	resp, err := http.Get("http://127.0.0.1:16060/debug/pprof/goroutine?debug=1")
	if err != nil {
		fmt.Println("pprof:", err)
		return
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	lines := strings.Split(string(body), "\n")

	for i := 0; i < len(lines); i++ {
		ln := lines[i]
		if strings.HasPrefix(ln, "goroutine profile:") {
			fmt.Println(ln)
			continue
		}
		if len(ln) > 0 && ln[0] >= '0' && ln[0] <= '9' && strings.Contains(ln, " @ ") {
			fmt.Println()
			fmt.Println(ln)
			if i+1 < len(lines) && strings.HasPrefix(lines[i+1], "#") {
				fmt.Println("   ", shortFrame(lines[i+1]))
			}
		}
	}
}

func shortFrame(s string) string {
	f := strings.Split(s, "\t")
	if len(f) >= 4 {
		return f[2] + "  " + f[3]
	}
	return s
}

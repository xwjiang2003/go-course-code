// 案例二 · 泄漏版：请求带了 ctx，但没有超时、没有 defer cancel，Client 也没设 Timeout
//
//	go run ./p09/leaky
//	go run ./p09/leaky -wait   # 挂住，方便抓 pprof
package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	_ "net/http/pprof"
	"os"
	"runtime"
	"sync"
	"time"
)

// call 看起来"用了 context"，其实三条保命线全漏了：
//  1. 没有 WithTimeout —— 这个 ctx 永远不会到点
//  2. 没有 defer cancel() —— 函数返回了，ctx 也不释放
//  3. Client 没有 Timeout —— 底层连接没有兜底
//
// 服务端不回，这个请求就挂到 TCP 超时才放手。
func call(url string) {
	ctx, cancel := context.WithCancel(context.Background())
	_ = cancel

	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	io.Copy(io.Discard, resp.Body)
	// 还忘了 resp.Body.Close()：连接不会被放回池子
}

func main() {
	go http.ListenAndServe("127.0.0.1:16063", nil)
	fmt.Printf("start      goroutines = %d\n", runtime.NumGoroutine())

	hang := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done() // 服务端故意永不返回
	}))

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); call(hang.URL) }()
	}

	time.Sleep(1500 * time.Millisecond)
	fmt.Printf("leaky x50  goroutines = %d\n", runtime.NumGoroutine())
	time.Sleep(2 * time.Second)
	fmt.Printf("wait 2s    goroutines = %d\n", runtime.NumGoroutine())

	if len(os.Args) > 1 && os.Args[1] == "-wait" {
		fmt.Println("pprof: http://127.0.0.1:16063/debug/pprof/goroutine?debug=1")
		select {}
	}
}

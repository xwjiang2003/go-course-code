// 案例一 · 泄漏版：发号器把号推给一个没人接收的无缓冲 channel
//
//	go run ./p06/leaky
//	go run ./p06/leaky -wait   # 挂住，方便抓 pprof
package main

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime"
	"time"
)

// dispatch 把订单号推给下游。ch 是无缓冲的：
// 发送必须等到有人接收，没人接收就永久阻塞在 ch <- orderID 这一行。
func dispatch(orderID int) {
	ch := make(chan int)
	go func() {
		ch <- orderID // ← goroutine 停在这一点，永不返回
	}()
}

func main() {
	go http.ListenAndServe("127.0.0.1:16061", nil)
	fmt.Printf("start     goroutines = %d\n", runtime.NumGoroutine())

	for i := 1; i <= 300; i++ {
		dispatch(i)
	}
	time.Sleep(300 * time.Millisecond)

	fmt.Printf("leaky x300 goroutines = %d\n", runtime.NumGoroutine())

	if len(os.Args) > 1 && os.Args[1] == "-wait" {
		fmt.Println("pprof: http://127.0.0.1:16061/debug/pprof/goroutine?debug=1")
		select {}
	}
}

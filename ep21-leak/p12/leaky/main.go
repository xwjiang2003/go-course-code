// 案例三 · 泄漏版：worker range 一个永远不 close 的 channel
//
//	go run ./p12/leaky
//	go run ./p12/leaky -wait
package main

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime"
	"time"
)

// startWorkers 起一批 worker 消费 ch。
// range 的结束条件是"channel 被 close"，而这里 ch 是每一轮新建的，
// 谁也没 close —— 所以这批 worker 永远不会退出。
//
// 真实场景：配置热更新，起了新一代 worker，旧的还在原来的 channel 上 range。
func startWorkers(n int) {
	ch := make(chan int)
	for i := 0; i < n; i++ {
		go func() {
			for v := range ch { // ← 卡在这里，永远等下一个值
				_ = v
			}
		}()
	}
	for i := 0; i < 5; i++ {
		ch <- i
	}
}

func main() {
	go http.ListenAndServe("127.0.0.1:16064", nil)
	fmt.Printf("start       goroutines = %d\n", runtime.NumGoroutine())

	for round := 1; round <= 20; round++ {
		startWorkers(10) // 每轮 10 个 worker，20 轮热更新
	}
	time.Sleep(500 * time.Millisecond)

	fmt.Printf("reload x20  goroutines = %d\n", runtime.NumGoroutine())

	if len(os.Args) > 1 && os.Args[1] == "-wait" {
		fmt.Println("pprof: http://127.0.0.1:16064/debug/pprof/goroutine?debug=1")
		select {}
	}
}

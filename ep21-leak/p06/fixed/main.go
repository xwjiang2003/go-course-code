// 案例一 · 修复版：缓冲兜住 + select 听取消，发送方不再无限等
//
//	go run ./p06/fixed
package main

import (
	"context"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"runtime"
	"time"
)

// dispatch 只做两件事：
//  1. 给 channel 一个缓冲，发送不必等接收方（没人收也能立刻返回）
//  2. select 同时听 ctx.Done()，上游撤销时立刻退出，不留在原地
func dispatch(ctx context.Context, orderID int) {
	ch := make(chan int, 1)
	go func() {
		select {
		case ch <- orderID:
		case <-ctx.Done():
		}
	}()
}

func main() {
	go http.ListenAndServe("127.0.0.1:16062", nil)
	fmt.Printf("start     goroutines = %d\n", runtime.NumGoroutine())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for i := 1; i <= 300; i++ {
		dispatch(ctx, i)
	}
	time.Sleep(300 * time.Millisecond)

	fmt.Printf("fixed x300 goroutines = %d\n", runtime.NumGoroutine())
	fmt.Println("↑ 回到基线：每个 goroutine 都自己退出了")
}

// 案例三 · 修复版：range 不是退出条件，close 和 ctx 才是
//
//	go run ./p12/fixed
package main

import (
	"context"
	"fmt"
	"runtime"
	"time"
)

// startWorkers 返回一个发送入口和一个 stop。
// 退出有两条路：channel 被 close（ok == false），或上游 ctx 被取消。
// 少了任何一条，worker 都可能留在原地。
func startWorkers(ctx context.Context, n int) (chan<- int, func()) {
	ch := make(chan int)
	for i := 0; i < n; i++ {
		go func() {
			for {
				select {
				case v, ok := <-ch:
					if !ok {
						return // 通道关了，收工
					}
					_ = v
				case <-ctx.Done():
					return // 上游撤了，收工
				}
			}
		}()
	}
	return ch, func() { close(ch) }
}

func main() {
	fmt.Printf("start       goroutines = %d\n", runtime.NumGoroutine())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for round := 1; round <= 20; round++ {
		in, stop := startWorkers(ctx, 10)
		for i := 0; i < 5; i++ {
			in <- i
		}
		stop() // 这一轮结束，明确告诉 worker：没有下一个值了
	}
	time.Sleep(500 * time.Millisecond)

	fmt.Printf("reload x20  goroutines = %d\n", runtime.NumGoroutine())
	fmt.Println("↑ 200 个 worker 全部退出，没有留一个")
}

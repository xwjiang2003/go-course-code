// 案例二 · 修复版：四条保命线，缺一条都白搭
//
//	go run ./p09/fixed
package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"runtime"
	"sync"
	"time"
)

// call 的四条线：
//  1. WithTimeout 给这次调用一个上限
//  2. defer cancel() 紧跟其后 —— 函数一返回就释放 ctx
//  3. Client 自己带 Timeout —— 就算外面塞了个没有 deadline 的 ctx 兜得住
//  4. defer resp.Body.Close() —— 连接才会被放回池子
func call(ctx context.Context, url string) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = io.Copy(io.Discard, resp.Body)
	return err
}

func main() {
	fmt.Printf("start      goroutines = %d\n", runtime.NumGoroutine())

	hang := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
		case <-time.After(10 * time.Second):
		}
	}))

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = call(context.Background(), hang.URL) // 2 秒后统一超时返回
		}()
	}
	wg.Wait()

	time.Sleep(300 * time.Millisecond)
	fmt.Printf("fixed x50  goroutines = %d\n", runtime.NumGoroutine())
	fmt.Println("↑ 50 个请求都超时了，但没有一个留在内存里")
	hang.Close()
}

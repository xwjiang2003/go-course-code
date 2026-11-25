// 第三道防线：goleak 单测兜底
//
//	go test ./p16 -run TestFixedDispatch -v    # 通过
//	go test ./p16 -run TestLeakyDispatch -v    # 失败，goleak 报出残留栈
package p16

import (
	"context"
	"testing"
	"time"

	"go.uber.org/goleak"
)

// VerifyTestMain 必须放在 TestMain 里才生效，写在普通测试函数里没用。
// 它在所有测试跑完后检查：还有没有本测试留下的 goroutine。
func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m,
		// 忽略测试框架自己的常驻 goroutine（不同 Go 版本名字会变）
		goleak.IgnoreTopFunction("internal/poll.runtime_pollWait"),
	)
}

// 泄漏版：测试函数返回了，goroutine 还卡在发送上
func TestLeakyDispatch(t *testing.T) {
	ch := make(chan int)
	go func() { ch <- 1 }()
	time.Sleep(100 * time.Millisecond)
}

// 修复版：缓冲兜住 + select 听取消
func TestFixedDispatch(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch := make(chan int, 1)
	go func() {
		select {
		case ch <- 1:
		case <-ctx.Done():
		}
	}()
	time.Sleep(100 * time.Millisecond)
}

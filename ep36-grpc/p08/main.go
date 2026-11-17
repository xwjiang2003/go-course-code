package main

// P08：ctx.Done() 该写在哪。
// 客户端 100ms 后主动取消，服务端必须立刻停下并回滚，而不是算完再发现没人要。

import (
	"context"
	"fmt"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	seckillpb "seckill/gen"
)

type inventory struct {
	seckillpb.UnimplementedInventoryServer
}

func (s *inventory) Deduct(ctx context.Context, req *seckillpb.DeductReq) (*seckillpb.DeductResp, error) {
	// 第一步：读库存
	remain, err := s.loadStock(ctx, req.SkuId)
	if err != nil {
		return nil, status.Error(codes.Internal, "load stock failed")
	}

	// 第二步：扣减前，先听取消
	select {
	case <-ctx.Done():
		// 用户已经走了，扣减不能继续
		fmt.Println("  server: ctx done -> rollback, 不扣减")
		return nil, status.FromContextError(ctx.Err()).Err()
	case <-time.After(600 * time.Millisecond):
		// 模拟一段耗时操作
	}

	return &seckillpb.DeductResp{Ok: true, Remain: remain - req.Count}, nil
}

func (s *inventory) loadStock(ctx context.Context, sku string) (int32, error) {
	if _, ok := ctx.Deadline(); !ok {
		// 没有 deadline 的请求，服务端自己兜底，避免永久占用
		fmt.Println("  server: 收到没有 deadline 的调用，使用本地兜底超时")
	}
	select {
	case <-ctx.Done():
		return 0, status.FromContextError(ctx.Err()).Err()
	case <-time.After(50 * time.Millisecond):
		return 100, nil
	}
}

func main() {
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}
	srv := grpc.NewServer()
	seckillpb.RegisterInventoryServer(srv, &inventory{})
	go srv.Serve(lis)
	defer srv.Stop()

	conn, err := grpc.NewClient(lis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	cli := seckillpb.NewInventoryClient(conn)

	// 场景一：客户端 100ms 后取消
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(100 * time.Millisecond)
		fmt.Println("  client: 用户关掉页面，cancel()")
		cancel()
	}()
	_, err = cli.Deduct(ctx, &seckillpb.DeductReq{SkuId: "SKU-1001", OrderId: "ORD-002", Count: 1})
	fmt.Printf("  client: code=%v msg=%q\n\n", status.Code(err), status.Convert(err).Message())

	// 场景二：给 2 秒预算，正常扣减
	ctx2, cancel2 := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel2()
	resp, err := cli.Deduct(ctx2, &seckillpb.DeductReq{SkuId: "SKU-1001", OrderId: "ORD-003", Count: 1})
	if err != nil {
		fmt.Println("  client: 正常路径也失败了:", err)
		return
	}
	fmt.Printf("  client: ok=%v remain=%d code=%v\n", resp.GetOk(), resp.GetRemain(), status.Code(err))
}

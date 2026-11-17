package main

// P05：上游没有时间预算，下游会把整条链拖住。
// 同一个库存接口，慢 800ms。两种调用方式对比：Background 与 400ms 预算。

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

type slowInventory struct {
	seckillpb.UnimplementedInventoryServer
	delay time.Duration
}

func (s *slowInventory) Deduct(ctx context.Context, req *seckillpb.DeductReq) (*seckillpb.DeductResp, error) {
	select {
	case <-time.After(s.delay):
		return &seckillpb.DeductResp{Ok: true, Remain: 12}, nil
	case <-ctx.Done():
		// 收到了上游的取消或超时，立刻停手，不做无用的扣减
		return nil, status.Error(codes.Canceled, "deduct aborted")
	}
}

func main() {
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}
	srv := grpc.NewServer()
	seckillpb.RegisterInventoryServer(srv, &slowInventory{delay: 800 * time.Millisecond})
	go srv.Serve(lis)
	defer srv.Stop()

	conn, err := grpc.NewClient(lis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	cli := seckillpb.NewInventoryClient(conn)
	req := &seckillpb.DeductReq{SkuId: "SKU-1001", OrderId: "ORD-20261116-001", Count: 1}

	// 写法一：没有预算，下游多久，上游就等多久
	t0 := time.Now()
	_, err = cli.Deduct(context.Background(), req)
	fmt.Printf("[Background] cost=%-6v code=%v\n", time.Since(t0).Round(time.Millisecond), status.Code(err))

	// 写法二：给库存 400ms 预算，到点直接放弃
	ctx, cancel := context.WithTimeout(context.Background(), 400*time.Millisecond)
	defer cancel()
	t1 := time.Now()
	_, err = cli.Deduct(ctx, req)
	fmt.Printf("[Timeout400] cost=%-6v code=%v\n", time.Since(t1).Round(time.Millisecond), status.Code(err))
}

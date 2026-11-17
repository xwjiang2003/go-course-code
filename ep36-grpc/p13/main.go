package main

// P13：连接复用与 round_robin。
// 每个请求 Dial 一次 = 每次都握手；NewClient 一次、全进程共用，才是常态。

import (
	"context"
	"fmt"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	seckillpb "seckill/gen"
)

type inventory struct {
	seckillpb.UnimplementedInventoryServer
}

func (s *inventory) Deduct(ctx context.Context, req *seckillpb.DeductReq) (*seckillpb.DeductResp, error) {
	return &seckillpb.DeductResp{Ok: true, Remain: 99}, nil
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

	// 只建一次连接：ServiceConfig 指定负载均衡策略
	conn, err := grpc.NewClient(
		lis.Addr().String(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(`{"loadBalancingConfig":[{"round_robin":{}}]}`),
	)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	cli := seckillpb.NewInventoryClient(conn)
	req := &seckillpb.DeductReq{SkuId: "SKU-1001", OrderId: "ORD-9001", Count: 1}

	t0 := time.Now()
	for i := 0; i < 200; i++ {
		if _, err := cli.Deduct(context.Background(), req); err != nil {
			fmt.Println("call failed:", err)
			return
		}
	}
	cost := time.Since(t0)
	fmt.Printf("200 次调用共用 1 条 ClientConn，总耗时 %v，平均 %v/次\n",
		cost.Round(time.Millisecond), (cost / 200).Round(time.Microsecond))
	fmt.Println("提示：多副本时地址要交给 resolver（dns:/// 或 K8s headless），round_robin 才会真的轮转")
}

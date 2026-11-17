package main

// P11：服务端流推送排队进度。
// 重点：io.EOF 是正常结束；客户端断开后 Send 会报错，服务端要及时退出。

import (
	"context"
	"fmt"
	"io"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	seckillpb "seckill/gen"
)

type orderServer struct {
	seckillpb.UnimplementedOrderServer
}

func (s *orderServer) QueueProgress(req *seckillpb.QueueReq, stream seckillpb.Order_QueueProgressServer) error {
	positions := []int32{3812, 2107, 806, 1}
	for i, pos := range positions {
		ev := &seckillpb.QueueEvent{
			OrderId:  req.GetOrderId(),
			Position: pos,
			EtaMs:    int32((4 - i) * 500),
			Status:   "waiting",
		}
		if i == len(positions)-1 {
			ev.Status = "created"
		}
		if err := stream.Send(ev); err != nil {
			// 客户端已经断开，继续循环只会白算
			fmt.Printf("  server: send failed (%v), stop pushing\n", err)
			return err
		}
		select {
		case <-stream.Context().Done():
			fmt.Println("  server: client gone, stop pushing")
			return stream.Context().Err()
		case <-time.After(300 * time.Millisecond):
		}
	}
	return nil
}

func (s *orderServer) Create(ctx context.Context, req *seckillpb.CreateReq) (*seckillpb.CreateResp, error) {
	return &seckillpb.CreateResp{OrderId: req.GetOrderId(), Status: "created"}, nil
}

func main() {
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}
	srv := grpc.NewServer()
	seckillpb.RegisterOrderServer(srv, &orderServer{})
	go srv.Serve(lis)
	defer srv.Stop()

	conn, err := grpc.NewClient(lis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	cli := seckillpb.NewOrderClient(conn)

	// 场景一：完整看完一段推送
	fmt.Println("--- 场景一：正常推送 ---")
	stream, err := cli.QueueProgress(context.Background(), &seckillpb.QueueReq{OrderId: "ORD-1001"})
	if err != nil {
		panic(err)
	}
	for {
		ev, err := stream.Recv()
		if err == io.EOF {
			fmt.Println("  client: io.EOF，服务端正常结束")
			break
		}
		if err != nil {
			fmt.Println("  client: recv error:", err)
			break
		}
		fmt.Printf("  client: %s 第 %d 位，预计 %dms\n", ev.GetOrderId(), ev.GetPosition(), ev.GetEtaMs())
	}

	// 场景二：只看一条就断开
	fmt.Println("--- 场景二：客户端提前断开 ---")
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	stream2, err := cli.QueueProgress(ctx, &seckillpb.QueueReq{OrderId: "ORD-1002"})
	if err != nil {
		panic(err)
	}
	ev, err := stream2.Recv()
	if err == nil {
		fmt.Printf("  client: 只收到第一条 %d，然后 disconnect\n", ev.GetPosition())
	}
	time.Sleep(900 * time.Millisecond) // 等服务端感知断开
	fmt.Println("  done")
}

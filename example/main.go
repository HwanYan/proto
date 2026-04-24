package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/HwanYan/proto/helloworld"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// GreeterServer 实现 gRPC 服务端
type GreeterServer struct {
	helloworld.UnimplementedGreeterServer
}

// SayHello 实现 gRPC 方法
func (s *GreeterServer) SayHello(ctx context.Context, req *helloworld.HelloRequest) (*helloworld.HelloResponse, error) {
	return &helloworld.HelloResponse{
		Msg: fmt.Sprintf("Hello, %s!", req.Msg),
	}, nil
}

// startGRPCServer 启动 gRPC 服务端
func startGRPCServer() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	helloworld.RegisterGreeterServer(s, &GreeterServer{})

	log.Printf("gRPC server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

// startHTTPServer 启动 HTTP 网关服务器
func startHTTPServer() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// 创建 gRPC 网关 mux
	mux := runtime.NewServeMux()

	// 注册 gRPC 网关处理器
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	err := helloworld.RegisterGreeterHandlerFromEndpoint(ctx, mux, "localhost:50051", opts)
	if err != nil {
		log.Fatalf("failed to register gateway: %v", err)
	}

	log.Printf("HTTP gateway server listening at :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("failed to serve HTTP: %v", err)
	}
}

// gRPCClientExample gRPC 客户端示例
func gRPCClientExample() {
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	client := helloworld.NewGreeterClient(conn)

	resp, err := client.SayHello(context.Background(), &helloworld.HelloRequest{Msg: "World"})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}

	fmt.Printf("gRPC Response: %s\n", resp.Msg)
}

// HTTPClientExample HTTP 客户端示例
func HTTPClientExample() {
	resp, err := http.Post("http://localhost:8080/v1/hello", "application/json",
		strings.NewReader(`{"msg": "World"}`))
	if err != nil {
		log.Fatalf("HTTP request failed: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("failed to read response: %v", err)
	}

	fmt.Printf("HTTP Response: %s\n", string(body))
}

func main() {
	// 在单独的 goroutine 中启动 gRPC 服务器
	go startGRPCServer()

	// 在单独的 goroutine 中启动 HTTP 网关
	go startHTTPServer()

	// 等待服务器启动
	time.Sleep(2 * time.Second)

	// 测试 gRPC 客户端
	fmt.Println("=== gRPC Client Test ===")
	gRPCClientExample()

	// 测试 HTTP 客户端
	fmt.Println("=== HTTP Client Test ===")
	HTTPClientExample()

	// 保持程序运行
	select {}
}

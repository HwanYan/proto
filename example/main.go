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
	// 创建一个 TCP 监听器，在本地所有网络接口的 50051 端口上监听传入的连接请求空 IP 即 0.0.0.0，接受来自任何 IP 的连接）
	// 返回的 lis（net.Listener）用于接受连接，后续 s.Serve(lis) 会循环接受连接并交给 gRPC 服务器处理
	// 如果端口已被占用或没有权限，err 会非 nil，代码会 log.Fatalf 退出
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// 创建一个新的 gRPC 服务器实例。该实例用于管理连接、处理请求、调用对应的服务方法
	s := grpc.NewServer()
	// 注册实现了 protobuf 定义的服务接口的结构体（如 helloworld.RegisterGreeterServer(s, &GreeterServer{})）
	helloworld.RegisterGreeterServer(s, &GreeterServer{})

	log.Printf("gRPC server listening at %v", lis.Addr())
	// 1. 启动接受循环
	// Serve 方法内部会执行一个无限循环，不断调用 lis.Accept() 来接受传入的 TCP 连接。每当有新连接建立，就会生成一个新的 goroutine 来处理该连接上的所有 RPC 请求
	// 2. 多路复用与流控
	// 对于每个连接，gRPC 基于 HTTP/2 协议复用多个请求/响应流。Serve 负责初始化解码器，处理 HTTP/2 的帧（Frame），并将每个收到的 RPC 请求路由到预先注册的服务方法。
	// 3. 请求处理流程
	// 	接收到一个完整的 RPC 请求后，gRPC 服务器根据方法名（如 /helloworld.Greeter/SayHello）找到对应的服务实现。
	// 	反序列化请求参数（protobuf 解码）。
	// 	调用用户注册的处理函数（如 GreeterServer 中的 SayHello）。
	// 	将响应序列化后发送回客户端。
	// 4. 优雅退出与生命周期
	// Serve 会一直运行直到发生错误（比如监听器关闭）或由外部主动调用 s.Stop()/s.GracefulStop()。当 Stop 被调用时，Serve 会返回 ErrServerStopped 错误，从而退出循环。
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

// startHTTPServer 启动 HTTP 网关服务器
func startHTTPServer() {
	// 创建一个可取消的 context，用于控制网关的生命周期
	// 当 cancel() 被调用时，网关会停止向后端 gRPC 服务发起新的请求
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// 创建 grpc-gateway 的 ServeMux（HTTP 路由复用器）
	// 它负责将 HTTP 请求（如 POST /v1/hello）映射到对应的 gRPC 方法（如 Greeter.SayHello）
	// 并完成 JSON <-> protobuf 的自动序列化/反序列化转换
	mux := runtime.NewServeMux()

	// 配置连接后端 gRPC 服务时使用的选项
	// insecure.NewCredentials() 表示不使用 TLS，以明文方式连接（仅适合本地开发）
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	// 将 Greeter 服务的 HTTP 路由注册到 mux 中
	// RegisterGreeterHandlerFromEndpoint 会在内部自动建立到 localhost:50051 的 gRPC 连接
	// 每当 HTTP 请求到来时，它会将请求转发给该 gRPC 端点并将响应回传给 HTTP 客户端
	err := helloworld.RegisterGreeterHandlerFromEndpoint(ctx, mux, "localhost:50051", opts)
	if err != nil {
		log.Fatalf("failed to register gateway: %v", err)
	}

	log.Printf("HTTP gateway server listening at :8080")
	// 启动标准 HTTP 服务器，监听 8080 端口，将所有请求交由 mux 处理
	// ListenAndServe 会阻塞当前 goroutine，直到发生错误（如端口被占用）才返回
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("failed to serve HTTP: %v", err)
	}
}

// gRPCClientExample gRPC 客户端示例
func gRPCClientExample() {
	// 建立到 gRPC 服务端的连接（ClientConn）
	// grpc.Dial 本身是非阻塞的，连接会在后台异步建立，首次 RPC 调用时才真正完成握手
	// insecure.NewCredentials() 表示不使用 TLS，以明文传输（仅适合本地开发）
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	// 函数返回时关闭连接，释放底层 TCP 资源
	defer conn.Close()

	// 基于已建立的连接创建 Greeter 服务的客户端 stub
	// stub 封装了所有 RPC 方法，调用 stub 上的方法等同于调用远端服务器上的对应方法
	client := helloworld.NewGreeterClient(conn)

	// 发起 SayHello RPC 调用
	// 1. stub 将 HelloRequest 序列化为 protobuf 二进制格式
	// 2. 通过 HTTP/2 将请求帧发送到服务端
	// 3. 服务端执行 SayHello 业务逻辑后，将 HelloResponse 序列化并回传
	// 4. stub 将响应反序列化为 HelloResponse 结构体返回给调用方
	resp, err := client.SayHello(context.Background(), &helloworld.HelloRequest{Msg: "World"})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}

	fmt.Printf("gRPC Response: %s\n", resp.Msg)
}

// HTTPClientExample HTTP 客户端示例
func HTTPClientExample() {
	// 向 HTTP 网关发送 POST 请求
	// 请求路径 /v1/hello 由 proto 文件中的 google.api.http 注解定义
	// 请求体为 JSON 格式，grpc-gateway 会自动将其反序列化为 HelloRequest protobuf 消息
	// 整个调用链路：HTTP Client -> grpc-gateway(:8080) -> gRPC Server(:50051)
	resp, err := http.Post("http://localhost:8080/v1/hello", "application/json",
		strings.NewReader(`{"msg": "World"}`))
	if err != nil {
		log.Fatalf("HTTP request failed: %v", err)
	}
	// 确保响应体被关闭，防止连接泄漏（HTTP 连接池复用依赖于此）
	defer resp.Body.Close()

	// 读取完整的响应体
	// grpc-gateway 已将 HelloResponse protobuf 消息序列化为 JSON 格式返回
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
	// select {}
}

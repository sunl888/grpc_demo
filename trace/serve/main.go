package main

import (
	"golang.org/x/net/context"
	"golang.org/x/net/trace" // 引入trace包
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
	pb "grpc_demo/proto/hello" // 引入编译生成的包
	"log"
	"net"
	"net/http"
)

const (
	// Address gRPC服务地址
	Address = "127.0.0.1:50052"
)

// 定义helloService并实现约定的接口
type helloService struct{}

// HelloService ...
var HelloService = helloService{}

func (h helloService) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	resp := new(pb.HelloReply)
	resp.Message = "Hello " + in.Name + "."

	return resp, nil
}

func main() {
	listen, err := net.Listen("tcp", Address)
	if err != nil {
		grpclog.Fatalf("failed to listen: %v", err)
	}

	// 实例化grpc Server
	s := grpc.NewServer()

	// 注册HelloService
	pb.RegisterHelloServer(s, HelloService)

	// 开启trace
	go startTrace()

	log.Println("Listen on " + Address)
	s.Serve(listen)
}

func startTrace() {
	trace.AuthRequest = func(req *http.Request) (any, sensitive bool) {
		return true, true
	}
	go http.ListenAndServe(":50051", nil)
	log.Println("Trace listen on 50051")
}

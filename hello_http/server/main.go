package main

import (
	"crypto/tls"
	"fmt"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	pb "grpc_demo/proto/hello_http"
	"log"
	"net"
	"net/http"
	"strings"
)

// 定义helloHttpService并实现约定的接口
type helloHttpService struct{}

// HelloHttpService ...
var HelloHttpService = helloHttpService{}

func (h helloHttpService) SayHello(ctx context.Context, in *pb.HelloHttpRequest) (*pb.HelloHttpReply, error) {
	resp := new(pb.HelloHttpReply)
	resp.Message = "Hello " + in.Name + "."

	return resp, nil
}

// grpcHandlerFunc 检查请求协议并返回http handler
func grpcHandlerFunc(grpcServer *grpc.Server, otherHandler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO(tamird): point to merged gRPC code rather than a PR.
		// This is a partial recreation of gRPC's internal checks https://github.com/grpc/grpc-go/pull/514/files#diff-95e9a25b738459a2d3030e1e6fa2a718R61
		if r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"), "application/grpc") {
			grpcServer.ServeHTTP(w, r)
		} else {
			otherHandler.ServeHTTP(w, r)
		}
	})
}

func main() {
	endpoint := "127.0.0.1:50052"

	conn, _ := net.Listen("tcp", endpoint)
	grpcServer := grpc.NewServer()
	pb.RegisterHelloHttpServer(grpcServer, HelloHttpService)

	// http-grpc gateway
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	dopts := []grpc.DialOption{grpc.WithInsecure()}
	gwmux := runtime.NewServeMux()
	err := pb.RegisterHelloHttpHandlerFromEndpoint(ctx, gwmux, endpoint, dopts)
	if err != nil {
		fmt.Printf("serve: %v\n", err)
		return
	}

	mux := http.NewServeMux()
	mux.Handle("/", gwmux)

	if err != nil {
		panic(err)
	}

	// 开启HTTP服务
	srv := &http.Server{
		Addr:    endpoint,
		Handler: grpcHandlerFunc(grpcServer, mux),
	}

	fmt.Printf("grpc and https on port: %d\n", 50052)

	err = srv.Serve(tls.NewListener(conn, srv.TLSConfig))

	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
	return
}
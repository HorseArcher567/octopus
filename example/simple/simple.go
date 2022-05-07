package main

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/k8s-practice/octopus"
	greeter "github.com/k8s-practice/octopus/example/simple/proto"
	"github.com/k8s-practice/octopus/pkg/log"
	"github.com/k8s-practice/octopus/pkg/service/ginsvc"
	"github.com/k8s-practice/octopus/pkg/service/grpcsvc"
	"github.com/k8s-practice/octopus/pkg/service/httpsvc"
	"google.golang.org/grpc"
	"net/http"
)

func main() {
	octopus.Init(octopus.WithConfigPath("./config/application.yaml"))

	grpcsvc.RegisterServer(func(server *grpc.Server) {
		greeter.RegisterGreeterServer(server, &GreeterServer{})
	})
	httpsvc.ServeMux().HandleFunc("/api/v1/greeter", func(w http.ResponseWriter, r *http.Request) {
		log.Info(w.Write([]byte("hello")))
	})
	ginsvc.Router().GET("/api/v1/greeter", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "hello")
	})

	octopus.Run()
}

type GreeterServer struct {
	greeter.UnimplementedGreeterServer
}

func (server *GreeterServer) Hello(context.Context, *greeter.HelloRequest) (*greeter.HelloResponse, error) {
	return &greeter.HelloResponse{
		Message: "hello",
	}, nil
}

func (server *GreeterServer) Bibi(context.Context, *greeter.HelloRequest) (*greeter.HelloResponse, error) {
	return &greeter.HelloResponse{
		Message: "hello",
	}, nil
}
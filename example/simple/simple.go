package main

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/k8s-practice/octopus"
	greeter "github.com/k8s-practice/octopus/example/simple/proto"
	"github.com/k8s-practice/octopus/pkg/grpcsvc"
	"github.com/k8s-practice/octopus/pkg/httpsvc"
	"google.golang.org/grpc"
	"net/http"
)

func main() {
	app := octopus.NewApplication()

	grpcService := grpcsvc.MustGetService("grpcService1")
	grpcService.Register(registerGreeter)

	httpService1 := httpsvc.MustGetService("httpService1")
	httpService1.Router().GET("/api/v1/greeter", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "this is http service1 greeter.")
	})

	httpService2 := httpsvc.MustGetService("httpService2")
	httpService2.Router().GET("/api/v1/greeter", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "this is http service2 greeter.")
	})

	app.Run()
}

type GreeterServer struct {
	greeter.UnimplementedGreeterServer
}

func registerGreeter(server *grpc.Server) {
	greeter.RegisterGreeterServer(server, &GreeterServer{})
}
func (server *GreeterServer) Hello(context.Context, *greeter.HelloRequest) (*greeter.HelloResponse, error) {
	return &greeter.HelloResponse{
		Message: "hello!",
	}, nil
}

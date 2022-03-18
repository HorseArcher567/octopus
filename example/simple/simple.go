package main

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/k8s-practice/octopus"
	greeter "github.com/k8s-practice/octopus/example/simple/proto"
	"github.com/k8s-practice/octopus/pkg/service/ginsvc"
	"github.com/k8s-practice/octopus/pkg/service/grpcsvc"
	"google.golang.org/grpc"
	"net/http"
)

func init() {
}

func main() {
	octopus.Init(octopus.WithConfigPath("./config/application.yaml"))

	grpcsvc.Register(registerGreeter)

	ginsvc.Router().GET("/api/v1/greeter", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "hello")
	})

	octopus.Run()
}

type GreeterServer struct {
	greeter.UnimplementedGreeterServer
}

func registerGreeter(server *grpc.Server) {
	greeter.RegisterGreeterServer(server, &GreeterServer{})
}
func (server *GreeterServer) Hello(context.Context, *greeter.HelloRequest) (*greeter.HelloResponse, error) {
	return &greeter.HelloResponse{
		Message: "hello",
	}, nil
}

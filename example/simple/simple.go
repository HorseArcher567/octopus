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
	_ "github.com/k8s-practice/octopus/pkg/service/otelsvc"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"net/http"
	"time"
)

func main() {
	octopus.Init(octopus.WithConfigPath("./config/application.yaml"))

	grpcsvc.RegisterServer(func(server *grpc.Server) {
		greeter.RegisterGreeterServer(server, &GreeterServer{})
	})
	httpsvc.HandleFunc("/api/v1/greeter", func(w http.ResponseWriter, r *http.Request) {
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

func (server *GreeterServer) Hello(ctx context.Context, _ *greeter.HelloRequest) (*greeter.HelloResponse, error) {
	_, span := otel.Tracer("greeter-hello").Start(ctx, "workHard",
		trace.WithAttributes(attribute.String("extra.key", "extra.value")))
	defer span.End()

	span.AddEvent("start sleep")
	time.Sleep(time.Millisecond)
	span.AddEvent("wake up")

	return &greeter.HelloResponse{
		Message: "hello",
	}, nil
}

func (server *GreeterServer) Bibi(context.Context, *greeter.HelloRequest) (*greeter.HelloResponse, error) {
	return &greeter.HelloResponse{
		Message: "hello",
	}, nil
}

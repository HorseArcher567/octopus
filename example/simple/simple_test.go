package main

import (
	"context"
	greeter "github.com/k8s-practice/octopus/example/simple/proto"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"io"
	"net/http"
	"testing"
	"time"
)

func TestGreeterServer_Hello(t *testing.T) {
	conn, err := grpc.Dial("localhost:9092", grpc.WithTransportCredentials(insecure.NewCredentials()))
	assert.Nil(t, err)

	client := greeter.NewGreeterClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	reply, err := client.Hello(ctx, &greeter.HelloRequest{})
	assert.Nil(t, err)
	assert.Equal(t, "hello", reply.Message)

	cancel()
}

func TestHttpService(t *testing.T) {
	reply, err := http.Get("http://localhost:9091/api/v1/greeter")
	assert.Nil(t, err)
	body, _ := io.ReadAll(reply.Body)
	assert.Equal(t, "hello", string(body))
}

func TestGinService(t *testing.T) {
	reply, err := http.Get("http://localhost:9090/api/v1/greeter")
	assert.Nil(t, err)
	body, _ := io.ReadAll(reply.Body)
	assert.Equal(t, "hello", string(body))
}

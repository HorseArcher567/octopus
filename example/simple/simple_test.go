package main

import (
	"context"
	greeter "github.com/k8s-practice/octopus/example/simple/proto"
	"github.com/k8s-practice/octopus/pkg/log"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"io"
	"net/http"
	"testing"
	"time"
)

func TestGreeterServer_Hello(t *testing.T) {
	conn, err := grpc.Dial("localhost:8082", grpc.WithInsecure())
	assert.Nil(t, err)

	client := greeter.NewGreeterClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	reply, err := client.Hello(ctx, &greeter.HelloRequest{})
	assert.Nil(t, err)
	log.Infoln(reply.Message)
}

func TestHttpService(t *testing.T) {
	reply, err := http.Get("http://localhost:8080/api/v1/greeter")
	assert.Nil(t, err)
	body, _ := io.ReadAll(reply.Body)
	log.Infoln(string(body))

	reply, err = http.Get("http://localhost:8081/api/v1/greeter")
	assert.Nil(t, err)
	body, _ = io.ReadAll(reply.Body)
	log.Infoln(string(body))
}

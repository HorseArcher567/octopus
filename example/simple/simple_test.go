package main

import (
	"context"
	greeter "github.com/k8s-practice/octopus/example/simple/proto"
	"github.com/k8s-practice/octopus/pkg/grpc/connpool"
	"github.com/k8s-practice/octopus/pkg/log"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/semconv/v1.10.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"io"
	"net/http"
	"testing"
	"time"
)

var connPool = connpool.New(
	grpc.WithTransportCredentials(insecure.NewCredentials()),
	grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()),
	grpc.WithStreamInterceptor(otelgrpc.StreamClientInterceptor()),
)

func init() {
	exporter, err := jaeger.New(jaeger.WithCollectorEndpoint(
		jaeger.WithEndpoint("http://localhost:14268/api/traces")))
	if err != nil {
		log.Panic(err)
	}

	tp := trace.NewTracerProvider(
		trace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNamespaceKey.String("dev-xlive"),
			semconv.ServiceNameKey.String("simple-client"))),
		trace.WithSampler(trace.AlwaysSample()),
		trace.WithSyncer(exporter),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{}))
}

func TestGreeterServer_Hello(t *testing.T) {
	conn := connPool.MustGetConn("localhost:9092")

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

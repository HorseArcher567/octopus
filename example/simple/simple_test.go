package main

import (
	"context"
	greeter "github.com/k8s-practice/octopus/example/simple/proto"
	"github.com/k8s-practice/octopus/pkg/grpc/connpool"
	"github.com/k8s-practice/octopus/pkg/log"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/contrib/instrumentation/net/http/httptrace/otelhttptrace"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdkTrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/semconv/v1.10.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"io"
	"net/http"
	"net/http/httptrace"
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

	tp := sdkTrace.NewTracerProvider(
		sdkTrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNamespaceKey.String("dev_xlive"),
			semconv.ServiceNameKey.String("simple_test"))),
		sdkTrace.WithSampler(sdkTrace.TraceIDRatioBased(0.5)),
		sdkTrace.WithSyncer(exporter),
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
	client := http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport)}
	tracer := otel.Tracer("example_test_http_client")
	ctx, span := tracer.Start(context.Background(), "greeter",
		trace.WithAttributes(semconv.PeerServiceKey.String("ExampleService")))
	defer span.End()

	ctx = httptrace.WithClientTrace(ctx, otelhttptrace.NewClientTrace(ctx))
	request, err := http.NewRequestWithContext(ctx, "GET",
		"http://localhost:9091/api/v1/greeter", nil)
	assert.Nil(t, err)

	reply, err := client.Do(request)
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

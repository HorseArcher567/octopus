package grpcsvc

import (
	"github.com/k8s-practice/octopus/pkg/log"
	"github.com/k8s-practice/octopus/pkg/prometheus/metrics"
	"github.com/k8s-practice/octopus/pkg/service"
	"google.golang.org/grpc"
	"reflect"
)

var (
	singleton *Service
)

func init() {
	// auto register builder
	service.RegisterBuilder(&Builder{})
}

type BeforeServe func()

type Service struct {
	enabled bool
	name    string
	server  *grpc.Server
	address string

	metrics     *metrics.GrpcServerMetrics
	beforeServe []BeforeServe
}

type RegisterFunc func(*grpc.Server)

// Register exposes *grpc.Server
func Register(f RegisterFunc) {
	if singleton != nil {
		f(singleton.server)
	} else {
		log.Panicf("%s uninitialized", reflect.TypeOf(singleton))
	}
}

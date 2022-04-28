package grpcsvc

import (
	"github.com/k8s-practice/octopus/pkg/log"
	"github.com/k8s-practice/octopus/pkg/prometheus/metrics"
	"github.com/k8s-practice/octopus/pkg/service"
	"google.golang.org/grpc"
	"reflect"
)

var (
	defaultBuilder = &builder{}
)

func init() {
	// auto register builder
	service.RegisterBuilder(defaultBuilder)
}

// RegisterServer registers server to *grpc.Server
func RegisterServer(f func(*grpc.Server)) {
	defaultBuilder.service.RegisterServer(f)
}

type Service struct {
	enabled bool
	name    string
	server  *grpc.Server
	address string

	metrics     *metrics.GrpcServerMetrics
	beforeServe []func()
}

func (svc *Service) RegisterServer(fn func(*grpc.Server)) {
	if svc == nil {
		log.Panicf("%s uninitialized", reflect.TypeOf(svc))
		return
	}
	fn(svc.server)
}

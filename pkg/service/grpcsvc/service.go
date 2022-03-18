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

// MustGetService gets service by name, or panic.
func MustGetService(name string) *Service {
	entry := service.GetEntry(name)
	if entry == nil {
		log.Panicf("%s service not exist", name)
		return nil
	}

	svc, ok := entry.(*Service)
	if !ok {
		log.Panicf("%s service isn't a %s", name, reflect.TypeOf(Service{}))
		return nil
	}
	return svc
}

type BeforeServe func()

type Service struct {
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

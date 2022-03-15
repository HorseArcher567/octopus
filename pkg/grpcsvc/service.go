package grpcsvc

import (
	"github.com/k8s-practice/octopus/pkg/log"
	"github.com/k8s-practice/octopus/pkg/service"
	"google.golang.org/grpc"
)

func init() {
	service.RegisterBuilder(&Builder{})
}

// MustGetService get service by name, or panic.
func MustGetService(name string) *Service {
	entry := service.Get(name)
	if entry == nil {
		log.Panicf("%s entry not exist", name)
		return nil
	}

	svc, ok := entry.(*Service)
	if !ok {
		log.Panicf("%s service is not a grpc service", name)
		return nil
	}
	return svc
}

type Service struct {
	name    string
	server  *grpc.Server
	address string
}

type RegisterFunc func(*grpc.Server)

func (svc *Service) Register(f RegisterFunc) {
	f(svc.server)
}

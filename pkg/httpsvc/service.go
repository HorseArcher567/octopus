package httpsvc

import (
	"github.com/k8s-practice/octopus/pkg/log"
	"github.com/k8s-practice/octopus/pkg/service"
	"net/http"
)

func init() {
	service.RegisterBuilder(&Builder{})
}

func MustGetService(name string) *Service {
	entry := service.Get(name)
	if entry == nil {
		log.Panicf("%s entry not exist", name)
		return nil
	}

	svc, ok := entry.(*Service)
	if !ok {
		log.Panicf("%s service is not a http service", name)
		return nil
	}
	return svc
}

type Service struct {
	name    string
	server  *http.Server
	address string
}

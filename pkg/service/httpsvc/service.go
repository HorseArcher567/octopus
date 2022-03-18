package httpsvc

import (
	"github.com/k8s-practice/octopus/pkg/log"
	"github.com/k8s-practice/octopus/pkg/service"
	"net/http"
	"reflect"
)

func init() {
	// auto register builder
	service.RegisterBuilder(&Builder{})
}

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

type Service struct {
	name    string
	server  *http.Server
	address string
}

func (svc *Service) Mux() *http.ServeMux {
	return svc.server.Handler.(*http.ServeMux)
}

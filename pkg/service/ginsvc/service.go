package ginsvc

import (
	"github.com/gin-gonic/gin"
	"github.com/k8s-practice/octopus/pkg/log"
	"github.com/k8s-practice/octopus/pkg/service"
	"net/http"
	"reflect"
)

var (
	singleton *Service
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

func Router() *gin.Engine {
	if singleton != nil {
		return singleton.server.Handler.(*gin.Engine)
	} else {
		log.Panicf("%s uninitialized", reflect.TypeOf(singleton))
		return nil
	}
}

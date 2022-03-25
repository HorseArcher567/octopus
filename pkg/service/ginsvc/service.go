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

type Service struct {
	name    string
	server  *http.Server
	address string
}

func Router() gin.IRouter {
	if singleton != nil {
		return singleton.server.Handler.(*gin.Engine)
	} else {
		log.Panicf("%s uninitialized", reflect.TypeOf(singleton))
		return nil
	}
}

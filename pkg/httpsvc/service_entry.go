package httpsvc

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/k8s-practice/octopus/pkg/log"
	"net"
)

func (svc *Service) Name() string {
	return svc.name
}

func (svc *Service) Run(_ context.Context) {
	listener, err := net.Listen("tcp", svc.address)
	if err != nil {
		log.Panicln(err)
		return
	}
	log.Errorln(svc.server.Serve(listener))
}

func (svc *Service) Stop(ctx context.Context) {
	log.Errorln(svc.server.Shutdown(ctx))
}

func (svc *Service) Router() *gin.Engine {
	return svc.server.Handler.(*gin.Engine)
}

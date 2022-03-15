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

func (svc *Service) Serve(_ context.Context) {
	log.Infoln(svc.Name(), "begin serve")

	listener, err := net.Listen("tcp", svc.address)
	if err != nil {
		log.Panicln(err)
		return
	}

	err = svc.server.Serve(listener)
	log.Errorln(svc.Name(), "end serve,", err)
}

func (svc *Service) Stop(ctx context.Context) {
	log.Errorln(svc.Name(), "stopped", svc.server.Shutdown(ctx))
}

func (svc *Service) Router() *gin.Engine {
	return svc.server.Handler.(*gin.Engine)
}

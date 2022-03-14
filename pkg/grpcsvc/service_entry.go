package grpcsvc

import (
	"context"
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

func (svc *Service) Stop(_ context.Context) {
	svc.server.GracefulStop()
}

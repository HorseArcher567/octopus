package grpcsvc

import (
	"context"
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
	log.Errorln(svc.Name(), "end serve", err)
}

func (svc *Service) Stop(_ context.Context) {
	svc.server.GracefulStop()
	log.Errorln("stopped", svc.Name())
}

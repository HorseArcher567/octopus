package grpcsvc

import (
	"context"
	"github.com/k8s-practice/octopus/pkg/log"
	"net"
)

// Name implements service.Entry interface.
func (svc *Service) Name() string {
	return svc.name
}

// Serve implements service.Serve interface.
func (svc *Service) Serve(_ context.Context) {
	log.Infoln(svc.Name(), "begin serve")

	listener, err := net.Listen("tcp", svc.address)
	if err != nil {
		log.Panicln(err)
		return
	}

	for i := 0; i < len(svc.beforeServe); i++ {
		svc.beforeServe[i]()
	}

	err = svc.server.Serve(listener)
	log.Errorln(svc.Name(), "end serve", err)
}

// Stop implements service.Stop interface.
func (svc *Service) Stop(_ context.Context) {
	svc.server.GracefulStop()
	log.Errorln("stopped", svc.Name())
}

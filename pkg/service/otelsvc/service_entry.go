package otelsvc

import "context"

func (svc *Service) Enabled() bool {
	return svc.enabled
}

func (svc *Service) Name() string {
	return svc.name
}

func (svc *Service) Serve(context.Context) {
}

func (svc *Service) Stop(context.Context) {
}

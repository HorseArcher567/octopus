package otelsvc

import "github.com/k8s-practice/octopus/pkg/service"

var (
	defaultBuilder = &builder{}
)

func init() {
	// auto register builder
	service.RegisterBuilder(defaultBuilder)
}

type Service struct {
	enabled bool
	name    string
}

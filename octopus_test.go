package octopus

import (
	"github.com/k8s-practice/octopus/pkg/service/grpcsvc"
	"github.com/k8s-practice/octopus/pkg/service/httpsvc"
	"github.com/k8s-practice/octopus/pkg/util/structure"
	"github.com/stretchr/testify/assert"
	"testing"
)

const rawConfig = `
grpc:
  name: gateGrpc
  enabled: true
  address: :8081
http:
  name: gateHttp
  enabled: true
  address: :8080
`

func TestWithConfigRawYaml(t *testing.T) {
	app := &application{}
	option := WithConfigRawYaml([]byte(rawConfig))
	option(app)
	app.initBootConfig()

	assert.Equal(t, 2, len(app.bootConfig))
	assert.NotNil(t, app.bootConfig["grpc"])
	assert.NotNil(t, app.bootConfig["http"])

	var grpcConfig grpcsvc.Config
	assert.Nil(t, structure.UnmarshalWithTag(app.bootConfig, &grpcConfig, "yaml"))
	assert.Equal(t, "gateGrpc", grpcConfig.Grpc.Name)
	assert.Equal(t, true, grpcConfig.Grpc.Enabled)
	assert.Equal(t, ":8081", grpcConfig.Grpc.Address)

	var httpConfig httpsvc.Config
	assert.Nil(t, structure.UnmarshalWithTag(app.bootConfig, &httpConfig, "yaml"))
	assert.Equal(t, "gateHttp", httpConfig.Http.Name)
	assert.Equal(t, true, httpConfig.Http.Enabled)
	assert.Equal(t, ":8080", httpConfig.Http.Address)
}

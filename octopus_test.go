package octopus

import (
	grpcsvc2 "github.com/k8s-practice/octopus/pkg/service/grpcsvc"
	httpsvc2 "github.com/k8s-practice/octopus/pkg/service/httpsvc"
	"github.com/k8s-practice/octopus/pkg/util/structure"
	"github.com/stretchr/testify/assert"
	"testing"
)

const rawConfig = `
grpc:
  - name: gateGrpc
    enabled: true
    address: :8081
http:
  - name: gateHttp
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

	var grpcConfig grpcsvc2.Config
	assert.Nil(t, structure.UnmarshalWithTag(app.bootConfig, &grpcConfig, "yaml"))
	assert.Equal(t, 1, len(grpcConfig.Grpc))
	assert.Equal(t, "gateGrpc", grpcConfig.Grpc[0].Name)
	assert.Equal(t, true, grpcConfig.Grpc[0].Enabled)
	assert.Equal(t, ":8081", grpcConfig.Grpc[0].Address)

	var httpConfig httpsvc2.Config
	assert.Nil(t, structure.UnmarshalWithTag(app.bootConfig, &httpConfig, "yaml"))
	assert.Equal(t, 1, len(httpConfig.Http))
	assert.Equal(t, "gateHttp", httpConfig.Http[0].Name)
	assert.Equal(t, true, httpConfig.Http[0].Enabled)
	assert.Equal(t, ":8080", httpConfig.Http[0].Address)
}

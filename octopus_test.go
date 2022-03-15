package octopus

import (
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
}

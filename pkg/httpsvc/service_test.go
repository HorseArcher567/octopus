package httpsvc

import (
	"github.com/k8s-practice/octopus/pkg/log"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
	"testing"
)

const rawYaml = `
http:
  - enabled: true
    name: httpSvc1
    address: 0.0.0.0:8080
  - enabled: false
    name: httpSvc2
    address: 127.0.0.1:8081
`

func TestBuilder_Build(t *testing.T) {
	var bootConfig map[interface{}]interface{}
	assert.Nil(t, yaml.Unmarshal([]byte(rawYaml), &bootConfig))
	log.Infof("%+v", bootConfig)

	var builder Builder
	services := builder.Build(bootConfig, "yaml")
	assert.Equal(t, 1, len(services))
}

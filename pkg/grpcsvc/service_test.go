package grpcsvc

import (
	"github.com/k8s-practice/octopus/pkg/log"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
	"testing"
)

const rawYaml = `
simple:
  - enabled: true
    name: grpcServer1
    address: 0.0.0.0:8081
  - enabled: false
    name: grpcServer2
    address: 127.0.0.1:8082
`

func TestBuilder_Build(t *testing.T) {
	var bootConfig map[interface{}]interface{}
	assert.Nil(t, yaml.Unmarshal([]byte(rawYaml), &bootConfig))
	log.Infof("%+v", bootConfig)

	var builder Builder
	services := builder.Build(bootConfig, "yaml")
	assert.Equal(t, 1, len(services))
}

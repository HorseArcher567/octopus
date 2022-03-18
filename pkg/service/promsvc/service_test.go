package promsvc

import (
	"github.com/k8s-practice/octopus/pkg/log"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
	"testing"
)

const rawYaml = `
prometheus:
  enabled: true
  path: /metrics
  name: promService
  address: :9093
`

func TestBuilder_Build(t *testing.T) {
	var bootConfig map[interface{}]interface{}
	assert.Nil(t, yaml.Unmarshal([]byte(rawYaml), &bootConfig))
	log.Infof("%+v", bootConfig)

	var builder Builder
	assert.NotNil(t, builder.Build(bootConfig, "yaml"))
}

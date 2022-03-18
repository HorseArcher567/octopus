package structure

import (
	"github.com/k8s-practice/octopus/pkg/log"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
	"testing"
)

func TestToSlice(t *testing.T) {
	s1 := []int{1, 2, 3, 4, 5, 6}
	var s2 []string
	assert.Nil(t, Unmarshal(s1, &s2))
}

func TestToArray(t *testing.T) {
	s1 := []int{1, 2, 3, 4, 5, 6}
	var s3 [6]string
	assert.Nil(t, Unmarshal(s1, &s3))
	log.Infoln(s1)
	log.Infoln(s3)
}

func TestToMap(t *testing.T) {
	type Meta struct {
		Name string
		Age  int
	}
	type Student struct {
		Name string
		Age  int
		Meta
	}
	s := &Student{"LiLei", 18, Meta{"HanMei", 16}}
	var m map[string]interface{}
	assert.Nil(t, Unmarshal(s, &m))
	log.Infoln(m)

	var s1 Student
	assert.Nil(t, Unmarshal(m, &s1))
	log.Infoln(s1)
}

const yamlData = `
apiVersion: v1
kind: Pod
metadata:
  annotations:
    prometheus.io/scrape: 'true'
  name: hello
  labels:
    app: hello
spec:
  containers:
  - name: hello
    image: artisan2yp/hello:1.0.3
  - name: world
    image: artisan2yp/world:1.0.1
`

func TestYamlToStruct(t *testing.T) {
	var m map[string]interface{}
	assert.Nil(t, yaml.Unmarshal([]byte(yamlData), &m))

	type PodDeploy struct {
		ApiVersion string `yaml:"apiVersion,omitempty" json:"apiVersion,omitempty"`
		Kind       string `yaml:"kind,omitempty" json:"kind,omitempty"`
		Metadata   struct {
			Name   string `yaml:"name,omitempty" json:"name,omitempty"`
			Labels struct {
				App string `yaml:"app,omitempty" json:"app,omitempty"`
			} `yaml:"labels,omitempty" json:"labels,omitempty"`
		} `yaml:"metadata,omitempty" json:"metadata,omitempty"`
		Spec struct {
			Containers []struct {
				Name  string `yaml:"-,omitempty" json:"-,omitempty"`
				Image string `yaml:"image,omitempty" json:"image,omitempty"`
			} `yaml:"containers,omitempty" json:"containers,omitempty"`
		} `yaml:"spec,omitempty" json:"spec,omitempty"`
	}
	var podDeploy PodDeploy
	assert.Nil(t, UnmarshalWithTag(m, &podDeploy, "yaml"))
	log.Infof("%+v", podDeploy)
}

func TestStructToMap(t *testing.T) {
	type Meta struct {
		Name string `yaml:"name,omitempty" json:"name,omitempty"`
		Age  int    `yaml:"age,omitempty" json:"age,omitempty"`
	}
	type Student struct {
		Meta  `yaml:"meta,omitempty" json:"meta,omitempty"`
		Class string `yaml:"class,omitempty" json:"class,omitempty"`
	}

	s1 := &Student{
		Meta{
			"LiLei",
			18,
		},
		"302",
	}

	var m map[string]interface{}
	assert.Nil(t, UnmarshalWithTag(s1, &m, "yaml"))
	log.Infof("%+v", m)

	var s2 Student
	assert.Nil(t, UnmarshalWithTag(m, &s2, "yaml"))
	log.Infof("%+v", s2)
}

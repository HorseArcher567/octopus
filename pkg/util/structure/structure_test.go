package structure

import (
	"github.com/BurntSushi/toml"
	"github.com/go-redis/redis/v8"
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

const tomlData = `
[database]
host = "172.16.116.50"
port = 3307
user = "tars"
password = "tars2015"
dbname = "k8s_tenants"
driver = "mysql"

[redis]
Address = ["172.16.116.50:9033","172.16.116.50:9034","172.16.116.50:9035"]
MaxRedirects = 8
ReadOnly = false
RouteByLatency = false
RouteRandomly = false

UserName = ""
Password = ""

#Duration,1是纳秒单位，1000是微秒，1000000是毫秒，1000000000是秒
MaxRetries = 0
MinRetryBackoff = 8_000_000
MaxRetryBackoff = 512_000_000

DialTimeout = 5_000_000_000
ReadTimeout = 3_000_000_000
WriteTimeout = 3_000_000_000

PoolSize = 15
MinIdleConns = 10
MaxConnAge = 0
PoolTimeout = 4_000_000_000
IdleTimeout = 300_000_000_000
IdleCheckFrequency = 60_000_000_000
`

func TestTomlToStruct(t *testing.T) {
	var m map[string]interface{}
	_, err := toml.Decode(tomlData, &m)
	assert.Nil(t, err)
	log.Infoln(m)

	type SqlConfig struct {
		Driver   string `mysql:"driver"`
		TenantId int32  `mysql:"tenant_id"`
		Host     string `mysql:"host"`
		Port     uint16 `mysql:"port"`
		User     string `mysql:"user"`
		Password string `mysql:"password"`
		DbName   string `mysql:"db_name"`
	}

	var sc SqlConfig
	assert.Nil(t, Unmarshal(m["database"], &sc))
	log.Infof("%+v", sc)

	var rc redis.ClusterOptions
	assert.Nil(t, Unmarshal(m["redis"], &rc))
	log.Infof("%+v", rc)
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

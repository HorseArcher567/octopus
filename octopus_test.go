package octopus

import "testing"

const rawConfig = `
grpc:
  - name: gateGrpc
    enabled: true
    ip: 0.0.0.0
    port: 8081
http:
  - name: gateHttp
    enabled: true
    ip: 0.0.0.0
    port: 8080
`

func TestWithApplicationConfigRawYaml(t *testing.T) {
	app := NewApplication(WithApplicationConfigRawYaml([]byte(rawConfig)))
	app.Run()
}

func TestWithApplicationConfigPath(t *testing.T) {
	app := NewApplication(WithApplicationConfigPath("./config/application.yaml"))
	app.Run()
}

package octopus

import "testing"

const rawConfig = `
simple-plugin:
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

func TestNewApplication(t *testing.T) {
	app := NewApplication(WithApplicationConfigRawYaml([]byte(rawConfig)))
	app.Run()

	app = NewApplication(WithApplicationConfigPath("./bootConfig/application.yaml"))
	app.Run()
}

package octopus

import (
	"context"
	"github.com/k8s-practice/octopus/pkg/log"
	"github.com/k8s-practice/octopus/pkg/service"
	"gopkg.in/yaml.v2"
	"os"
	"os/signal"
	"time"
)

const (
	defaultApplicationConfigPath = `./config/application.yaml`
)

var (
	defaultApp = &application{}
)

func Init(options ...Option) {
	for _, option := range options {
		option(defaultApp)
	}

	// load boot config from raw yaml or local yaml file
	defaultApp.initBootConfig()
	// init all service by boot config
	defaultApp.initService()
}

func Run() {
	defaultApp.run()
}

type application struct {
	configPath    string
	configRawYaml []byte
	bootConfig    map[interface{}]interface{}
}

type Option func(*application)

func WithConfigRawYaml(rawYaml []byte) Option {
	return func(app *application) {
		app.configRawYaml = rawYaml
	}
}

func WithConfigPath(path string) Option {
	return func(app *application) {
		app.configPath = path
	}
}

func (app *application) run() {
	// start all services
	app.startService()
	// waiting for interrupt signal
	app.listenSignal()
	// stop all services
	app.stopService()
}

func (app *application) initBootConfig() {
	if len(app.configRawYaml) == 0 {
		if len(app.configPath) == 0 {
			app.configPath = defaultApplicationConfigPath
		}

		if rawYaml, err := os.ReadFile(app.configPath); err != nil {
			log.Panicln(err)
		} else {
			app.configRawYaml = rawYaml
		}
	}

	if err := yaml.Unmarshal(app.configRawYaml, &app.bootConfig); err != nil {
		log.Panicln(err)
	}
}

func (app *application) initService() {
	if i, ok := app.bootConfig["service"]; ok {
		if v, ok := i.(map[interface{}]interface{}); ok {
			service.Init(v)
		}
	}
}

func (app *application) startService() {
	service.Start(context.Background())
}

func (app *application) stopService() {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	service.Stop(ctx)
}

func (app *application) listenSignal() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	select {
	case s := <-c:
		log.Errorln("receive", s, "signal")
	}
}

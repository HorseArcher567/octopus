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

type Application struct {
	configPath    string
	configRawYaml []byte
	bootConfig    map[interface{}]interface{}
}

type Option func(*Application)

func WithApplicationConfigRawYaml(rawYaml []byte) Option {
	return func(app *Application) {
		if len(app.configPath) != 0 {
			log.Panicln("bootConfig path already specified.")
		}
		app.configRawYaml = rawYaml
	}
}

func WithApplicationConfigPath(path string) Option {
	return func(app *Application) {
		if len(app.configRawYaml) != 0 {
			log.Panicln("bootConfig raw yaml already specified.")
		}
		app.configPath = path
	}
}

func NewApplication(options ...Option) *Application {
	app := &Application{}
	for _, option := range options {
		option(app)
	}

	app.initBootConfig()
	app.initService()

	return app
}

func (app *Application) Run() {
	app.startService()
	app.listenSignal()
	app.stopService()
}

func (app *Application) initBootConfig() {
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

func (app *Application) initService() {
	service.Init(app.bootConfig)
}

func (app *Application) startService() {
	service.Start(context.Background())
}

func (app *Application) stopService() {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	service.Stop(ctx)
}

func (app *Application) listenSignal() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	select {
	case <-ctx.Done():
		log.Errorln("application will exist,", ctx.Err())
	}
}

package octopus

import (
	"context"
	"github.com/k8s-practice/octopus/pkg/log"
	"github.com/k8s-practice/octopus/pkg/service"
	"gopkg.in/yaml.v2"
	"os"
	"os/signal"
	"syscall"
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

	service.Init(app.bootConfig)
	return app
}

func (app *Application) Run() {
	service.Run(context.Background())
	app.listenSignal()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	service.Stop(ctx)
}

func (app *Application) listenSignal() {
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	select {
	case sig := <-c:
		log.Warnln("receive signal:", sig)
		return
	}
}

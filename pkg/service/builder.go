package service

const (
	defaultConfigTag = "yaml"
)

var (
	registeredBuilders []Builder
)

type Builder interface {
	Build(bootConfig map[interface{}]interface{}, tag string) []Entry
}

func RegisterBuilder(builder Builder) {
	registeredBuilders = append(registeredBuilders, builder)
}

func buildEntries(bootConfig map[interface{}]interface{}) {
	for _, builder := range registeredBuilders {
		services := builder.Build(bootConfig, defaultConfigTag)
		Register(services...)
	}
}

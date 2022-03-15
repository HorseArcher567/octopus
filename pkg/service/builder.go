package service

const (
	defaultConfigTag = "yaml"
)

var (
	registeredBuilders []Builder
)

// Builder build service entries by config.
type Builder interface {
	Build(bootConfig map[interface{}]interface{}, tag string) []Entry
}

// RegisterBuilder register builder, and build service entries while service.Init function called.
func RegisterBuilder(builder Builder) {
	registeredBuilders = append(registeredBuilders, builder)
}

func buildEntries(bootConfig map[interface{}]interface{}) {
	for _, builder := range registeredBuilders {
		entries := builder.Build(bootConfig, defaultConfigTag)
		register(entries...)
	}
}

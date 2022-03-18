package service

const (
	defaultConfigTag = "yaml"
)

var (
	registeredBuilders []Builder
)

// Builder build service entries by config.
type Builder interface {
	Build(bootConfig map[interface{}]interface{}, tag string) Entry
}

// RegisterBuilder registerEntry builder, and build service entries while service.Init
// function called.
func RegisterBuilder(builder Builder) {
	registeredBuilders = append(registeredBuilders, builder)
}

// buildEntry build service entry by invoke registered Builder.Build.
func buildEntry(bootConfig map[interface{}]interface{}) []Entry {
	var entries []Entry
	for _, builder := range registeredBuilders {
		entry := builder.Build(bootConfig, defaultConfigTag)
		entries = append(entries, entry)
	}
	return entries
}

package service

const (
	defaultConfigTag = "yaml"
)

var (
	registeredBuilders []Builder
)

// Builder builds service.Entry by config.
type Builder interface {
	Build(bootConfig map[interface{}]interface{}, tag string) Entry
}

// RegisterBuilder registers service.Entry builder,
// service.Init will build service.Entry.
func RegisterBuilder(builder Builder) {
	registeredBuilders = append(registeredBuilders, builder)
}

// buildEntry builds service.Entry by invoke registered Builder.Build.
func buildEntry(bootConfig map[interface{}]interface{}) []Entry {
	var entries []Entry
	for _, builder := range registeredBuilders {
		entry := builder.Build(bootConfig, defaultConfigTag)
		entries = append(entries, entry)
	}
	return entries
}

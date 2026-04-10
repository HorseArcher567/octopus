package resource

// Runtime is the generic resource access contract.
type Runtime interface {
	Register(kind, name string, value any, closeFn func() error) error
	Get(kind, name string) (any, error)
	MustGet(kind, name string) any
	List(kind string) []string
	Close() error
}

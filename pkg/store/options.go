package store

type SetOption func(*entry) error

func WithClose(fn func() error) SetOption {
	return func(e *entry) error {
		e.closeFn = fn
		return nil
	}
}

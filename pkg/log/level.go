package log

import "fmt"

type Level int32

const (
	DebugLevel = iota
	InfoLevel
	WarnLevel
	ErrorLevel
	PanicLevel
	FatalLevel
)

func (lvl Level) String() string {
	switch lvl {
	case DebugLevel:
		return "DEBUG"
	case InfoLevel:
		return "INFO"
	case WarnLevel:
		return "WARN"
	case ErrorLevel:
		return "ERROR"
	case PanicLevel:
		return "PANIC"
	case FatalLevel:
		return "FATAL"
	default:
		return fmt.Sprintf("Level(%d)", lvl)
	}
}

func (lvl Level) Enabled(l Level) bool {
	return l >= lvl
}

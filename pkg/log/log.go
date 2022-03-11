package log

import (
	"fmt"
	"os"
)

const callDepth = 3

func SetLevel(level Level) {
	stdLogger.SetLevel(level)
}

func GetLevel() Level {
	return stdLogger.GetLevel()
}

func Enabled(level Level) bool {
	return stdLogger.Enabled(level)
}

func SetFlag(flag Flag) {
	stdLogger.SetFlag(flag)
}

func GetFlag() Flag {
	return stdLogger.GetFlag()
}

func Debug(v ...interface{}) {
	if stdLogger.Enabled(DebugLevel) {
		stdLogger.Output(callDepth, DebugLevel, fmt.Sprint(v...))
	}
}

func Debugf(format string, v ...interface{}) {
	if stdLogger.Enabled(DebugLevel) {
		stdLogger.Output(callDepth, DebugLevel, fmt.Sprintf(format, v...))
	}
}

func Debugln(v ...interface{}) {
	if stdLogger.Enabled(DebugLevel) {
		stdLogger.Output(callDepth, DebugLevel, fmt.Sprintln(v...))
	}
}

func Info(v ...interface{}) {
	if stdLogger.Enabled(InfoLevel) {
		stdLogger.Output(callDepth, InfoLevel, fmt.Sprint(v...))
	}
}

func Infof(format string, v ...interface{}) {
	if stdLogger.Enabled(InfoLevel) {
		stdLogger.Output(callDepth, InfoLevel, fmt.Sprintf(format, v...))
	}
}

func Infoln(v ...interface{}) {
	if stdLogger.Enabled(InfoLevel) {
		stdLogger.Output(callDepth, InfoLevel, fmt.Sprintln(v...))
	}
}

func Warn(v ...interface{}) {
	if stdLogger.Enabled(WarnLevel) {
		stdLogger.Output(callDepth, WarnLevel, fmt.Sprint(v...))
	}
}

func Warnf(format string, v ...interface{}) {
	if stdLogger.Enabled(WarnLevel) {
		stdLogger.Output(callDepth, WarnLevel, fmt.Sprintf(format, v...))
	}
}

func Warnln(v ...interface{}) {
	if stdLogger.Enabled(WarnLevel) {
		stdLogger.Output(callDepth, WarnLevel, fmt.Sprintln(v...))
	}
}

func Error(v ...interface{}) {
	if stdLogger.Enabled(ErrorLevel) {
		stdLogger.Output(callDepth, ErrorLevel, fmt.Sprint(v...))
	}
}

func Errorf(format string, v ...interface{}) {
	if stdLogger.Enabled(ErrorLevel) {
		stdLogger.Output(callDepth, ErrorLevel, fmt.Sprintf(format, v...))
	}
}

func Errorln(v ...interface{}) {
	if stdLogger.Enabled(ErrorLevel) {
		stdLogger.Output(callDepth, ErrorLevel, fmt.Sprintln(v...))
	}
}

func Panic(v ...interface{}) {
	if stdLogger.Enabled(PanicLevel) {
		s := fmt.Sprint(v...)
		stdLogger.Output(callDepth, PanicLevel, s)
		panic(s)
	}
}

func Panicf(format string, v ...interface{}) {
	if stdLogger.Enabled(PanicLevel) {
		s := fmt.Sprintf(format, v...)
		stdLogger.Output(callDepth, PanicLevel, s)
		panic(s)
	}
}

func Panicln(v ...interface{}) {
	if stdLogger.Enabled(PanicLevel) {
		s := fmt.Sprintln(v...)
		stdLogger.Output(callDepth, PanicLevel, s)
		panic(s)
	}
}

func Fatal(v ...interface{}) {
	if stdLogger.Enabled(FatalLevel) {
		stdLogger.Output(callDepth, FatalLevel, fmt.Sprint(v...))
		os.Exit(1)
	}
}

func Fatalf(format string, v ...interface{}) {
	if stdLogger.Enabled(FatalLevel) {
		stdLogger.Output(callDepth, FatalLevel, fmt.Sprintf(format, v...))
		os.Exit(1)
	}
}

func Fatalln(v ...interface{}) {
	if stdLogger.Enabled(FatalLevel) {
		stdLogger.Output(callDepth, FatalLevel, fmt.Sprintln(v...))
		os.Exit(1)
	}
}

func Print(incDepth int, level Level, v ...interface{}) {
	if !stdLogger.Enabled(level) {
		return
	}

	body := fmt.Sprint(v...)
	stdLogger.Output(callDepth+incDepth, level, body)
	if level == PanicLevel {
		panic(body)
	} else if level == FatalLevel {
		os.Exit(1)
	}
}

func Println(incDepth int, level Level, v ...interface{}) {
	if !stdLogger.Enabled(level) {
		return
	}

	body := fmt.Sprintln(v...)
	stdLogger.Output(callDepth+incDepth, level, body)
	if level == PanicLevel {
		panic(body)
	} else if level == FatalLevel {
		os.Exit(1)
	}
}

func Printf(incDepth int, level Level, format string, v ...interface{}) {
	if !stdLogger.Enabled(level) {
		return
	}

	body := fmt.Sprintf(format, v...)
	stdLogger.Output(callDepth+incDepth, level, body)
	if level == PanicLevel {
		panic(body)
	} else if level == FatalLevel {
		os.Exit(1)
	}
}

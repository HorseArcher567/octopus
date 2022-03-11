package log

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"sync/atomic"
	"time"
)

var stdLogger = NewLogger(InfoLevel, StdFlags, os.Stderr)

func NewLogger(level Level, flag Flag, out io.Writer) *Logger {
	return &Logger{
		level: int32(level),
		flag:  int32(flag),
		out:   out,
	}
}

type Logger struct {
	level int32
	flag  int32
	out   io.Writer
}

func (l *Logger) SetLevel(level Level) {
	atomic.StoreInt32(&l.level, int32(level))
}

func (l *Logger) GetLevel() Level {
	return Level(atomic.LoadInt32(&l.level))
}

func (l *Logger) Enabled(lvl Level) bool {
	return l.GetLevel().Enabled(lvl)
}

func (l *Logger) SetFlag(flag Flag) {
	atomic.StoreInt32(&l.flag, int32(flag))
}

func (l *Logger) GetFlag() Flag {
	return Flag(atomic.LoadInt32(&l.flag))
}

func (l *Logger) Debugf(format string, v ...interface{}) {
	if l.Enabled(DebugLevel) {
		l.Output(callDepth, DebugLevel, fmt.Sprintf(format, v...))
	}
}

func (l *Logger) Debug(v ...interface{}) {
	if l.Enabled(DebugLevel) {
		l.Output(callDepth, DebugLevel, fmt.Sprint(v...))
	}
}

func (l *Logger) Debugln(v ...interface{}) {
	if l.Enabled(DebugLevel) {
		l.Output(callDepth, DebugLevel, fmt.Sprintln(v...))
	}
}

func (l *Logger) Infof(format string, v ...interface{}) {
	if l.Enabled(InfoLevel) {
		l.Output(callDepth, InfoLevel, fmt.Sprintf(format, v...))
	}
}

func (l *Logger) Info(v ...interface{}) {
	if l.Enabled(InfoLevel) {
		l.Output(callDepth, InfoLevel, fmt.Sprint(v...))
	}
}

func (l *Logger) Infoln(v ...interface{}) {
	if l.Enabled(InfoLevel) {
		l.Output(callDepth, InfoLevel, fmt.Sprintln(v...))
	}
}

func (l *Logger) Warnf(format string, v ...interface{}) {
	if l.Enabled(WarnLevel) {
		l.Output(callDepth, WarnLevel, fmt.Sprintf(format, v...))
	}
}

func (l *Logger) Warn(v ...interface{}) {
	if l.Enabled(WarnLevel) {
		l.Output(callDepth, WarnLevel, fmt.Sprint(v...))
	}
}

func (l *Logger) Warnln(v ...interface{}) {
	if l.Enabled(WarnLevel) {
		l.Output(callDepth, WarnLevel, fmt.Sprintln(v...))
	}
}

func (l *Logger) Errorf(format string, v ...interface{}) {
	if l.Enabled(WarnLevel) {
		l.Output(callDepth, ErrorLevel, fmt.Sprintf(format, v...))
	}
}

func (l *Logger) Error(v ...interface{}) {
	if l.Enabled(ErrorLevel) {
		l.Output(callDepth, ErrorLevel, fmt.Sprint(v...))
	}
}

func (l *Logger) Errorln(v ...interface{}) {
	if l.Enabled(ErrorLevel) {
		l.Output(callDepth, ErrorLevel, fmt.Sprintln(v...))
	}
}

func (l *Logger) Panic(v ...interface{}) {
	if l.Enabled(PanicLevel) {
		s := fmt.Sprint(v...)
		l.Output(callDepth, PanicLevel, s)
		panic(s)
	}
}

func (l *Logger) Panicf(format string, v ...interface{}) {
	if l.Enabled(PanicLevel) {
		s := fmt.Sprintf(format, v...)
		l.Output(callDepth, PanicLevel, s)
		panic(s)
	}
}

func (l *Logger) Panicln(v ...interface{}) {
	if l.Enabled(PanicLevel) {
		s := fmt.Sprintln(v...)
		l.Output(callDepth, PanicLevel, s)
		panic(s)
	}
}

func (l *Logger) Fatal(v ...interface{}) {
	if l.Enabled(FatalLevel) {
		l.Output(callDepth, FatalLevel, fmt.Sprint(v...))
		os.Exit(1)
	}
}

func (l *Logger) Fatalf(format string, v ...interface{}) {
	if l.Enabled(FatalLevel) {
		l.Output(callDepth, FatalLevel, fmt.Sprintf(format, v...))
		os.Exit(1)
	}
}

func (l *Logger) Fatalln(v ...interface{}) {
	if l.Enabled(FatalLevel) {
		l.Output(callDepth, FatalLevel, fmt.Sprintln(v...))
		os.Exit(1)
	}
}

func (l *Logger) Output(callDepth int, level Level, body string) {
	now := time.Now()
	buf := get()

	flag := l.GetFlag()
	if flag&(Caller|ShortFile|LongFile) != 0 {
		pc := make([]uintptr, 1)
		numFrames := runtime.Callers(callDepth, pc)
		if numFrames == 1 {
			frame, _ := runtime.CallersFrames(pc).Next()
			l.formatHeader(&buf, flag, now, level, frame.File, frame.Function, frame.Line)
		} else {
			l.formatHeader(&buf, flag, now, level, "???", "??", 0)
		}
	} else {
		l.formatHeader(&buf, flag, now, level, "", "", 0)
	}

	buf = append(buf, body...)
	if len(body) == 0 || body[len(body)-1] != '\n' {
		buf = append(buf, '\n')
	}

	if _, err := l.out.Write(buf); err != nil {
		fmt.Println(err)
	}

	put(buf)
}

const delimiter = '\t'

func (l *Logger) formatHeader(buf *[]byte, flag Flag, now time.Time, level Level, file, function string, line int) {
	if flag.SpecificTime() {
		*buf = append(*buf, now.Format(time.RFC3339Nano)...)
		*buf = append(*buf, delimiter)
	}

	*buf = append(*buf, level.String()...)
	*buf = append(*buf, delimiter)

	if flag.SpecificShortFile() {
		*buf = append(*buf, filepath.Base(file)...)
		*buf = append(*buf, ':')
		*buf = strconv.AppendInt(*buf, int64(line), 10)
		*buf = append(*buf, delimiter)
	} else if flag.SpecificLongFile() {
		*buf = append(*buf, file...)
		*buf = append(*buf, ':')
		*buf = strconv.AppendInt(*buf, int64(line), 10)
		*buf = append(*buf, delimiter)
	}

	if flag.SpecificCaller() {
		*buf = append(*buf, function...)
		*buf = append(*buf, delimiter)
	}
}

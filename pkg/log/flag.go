package log

type Flag int32

const (
	// Time is the date in the local time zone: 2009/01/23 01:23:23.9999
	Time = 1 << iota
	// LongFile is the full file name and line number: /a/b/c/d.go:23
	LongFile
	// ShortFile si the final file name element and line number: d.go:23. overrides LongFile
	ShortFile
	// Caller is the caller of logger: main.fun1()
	Caller
	// StdFlags is the initial values for the standard logger
	StdFlags = Time | ShortFile | Caller
)

func (f Flag) SpecificTime() bool {
	return f&Time != 0
}

func (f Flag) SpecificCaller() bool {
	return f&Caller != 0
}

func (f Flag) SpecificLongFile() bool {
	return f&LongFile != 0
}

func (f Flag) SpecificShortFile() bool {
	return f&ShortFile != 0
}

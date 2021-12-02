package qulibs

// A LogLevelType defines the level logging should be performed at. Used to instruct
// the SDK which statements should be logged.
type LogLevelType uint

// LogLevel returns the pointer to a LogLevel. Should be used to workaround
// not being able to take the address of a non-composite literal.
func LogLevel(l LogLevelType) *LogLevelType {
	return &l
}

// Value returns the LogLevel value or the default value LogOff if the LogLevel
// is nil. Safe to use on nil value LogLevelTypes.
func (l *LogLevelType) Value() LogLevelType {
	if l != nil {
		return *l
	}
	return LogOff
}

// Matches returns true if the v LogLevel is enabled by this LogLevel. Should be
// used with logging sub levels. Is safe to use on nil value LogLevelTypes. If
// LogLevel is nil, will default to LogOff comparison.
func (l *LogLevelType) Matches(v LogLevelType) bool {
	c := l.Value()
	return c&v == v
}

// AtLeast returns true if this LogLevel is at least high enough to satisfies v.
// Is safe to use on nil value LogLevelTypes. If LogLevel is nil, will default
// to LogOff comparison.
func (l *LogLevelType) AtLeast(v LogLevelType) bool {
	return l.Value() >= v
}

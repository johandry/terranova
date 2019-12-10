package logger

// EmptyLog is an implementation of a Logger that do nothing. It's simmilar to
// Discard.
type EmptyLog struct{}

// NewEmptyLog create a Terranova Logger that do nothing.
func NewEmptyLog() Logger {
	return &EmptyLog{}
}

// Printf implements a standard Printf function of Logger interface
func (l *EmptyLog) Printf(format string, args ...interface{}) {}

// Debugf implements a standard Debugf function of Logger interface
func (l *EmptyLog) Debugf(format string, args ...interface{}) {}

// Infof implements a standard Infof function of Logger interface
func (l *EmptyLog) Infof(format string, args ...interface{}) {}

// Warnf implements a standard Warnf function of Logger interface
func (l *EmptyLog) Warnf(format string, args ...interface{}) {}

// Errorf implements a standard Errorf function of Logger interface
func (l *EmptyLog) Errorf(format string, args ...interface{}) {}

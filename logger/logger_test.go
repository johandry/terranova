/*
Copyright The Terranova Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package logger

import (
	"io"
	"log"
	"strings"
)

// MockLog is an implementation of a Logger that capture the log output and send
// it to a buffer. It's useful for testing
type MockLog struct {
	log *log.Logger
}

// NewMockLog create a Terranova Logger that do nothing.
func NewMockLog(w io.Writer) Logger {
	l := log.New(w, "", log.LstdFlags)

	return &MockLog{
		log: l,
	}
}

// Printf implements a standard Printf function of Logger interface
func (l *MockLog) Printf(format string, args ...interface{}) {
	l.output("     ", format, args...)
}

// Debugf implements a standard Debugf function of Logger interface
func (l *MockLog) Debugf(format string, args ...interface{}) {
	if strings.HasPrefix(format, TracePrefix) {
		l.tracef(format, args...)
		return
	}
	l.output("DEBUG", format, args...)
}

// tracef is a second level of debug, it's call by Debugf when TraceLevel is set
func (l *MockLog) tracef(format string, args ...interface{}) {
	format = strings.TrimPrefix(format, TracePrefix)
	l.output("TRACE", format, args...)
}

// Infof implements a standard Infof function of Logger interface
func (l *MockLog) Infof(format string, args ...interface{}) {
	l.output("INFO ", format, args...)
}

// Warnf implements a standard Warnf function of Logger interface
func (l *MockLog) Warnf(format string, args ...interface{}) {
	l.output("WARN ", format, args...)
}

// Errorf implements a standard Errorf function of Logger interface
func (l *MockLog) Errorf(format string, args ...interface{}) {
	l.output("ERROR", format, args...)
}

func (l *MockLog) output(levelStr string, format string, args ...interface{}) {
	l.log.SetPrefix(levelStr + " [ ")
	l.log.Printf("] "+format, args...)
}

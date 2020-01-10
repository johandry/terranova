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
	"fmt"
	"io"
	"log"
	"strings"
)

// Log is a basic implementation of the Logger interface that wraps the Go log
// package and implements Debugf
type Log struct {
	Prefix string
	Level  Level
	log    *log.Logger
}

// Level is the log level type
type Level uint8

// Different log level from higher to lower.
// Lower log levels to the level set won't be displayed.
// i.e. If LogLevelWarn is set only Warnings and Errors are displayed.
// DefLogLevel defines the log level to assing by default
const (
	LogLevelError Level = iota
	LogLevelWarn
	LogLevelInfo
	LogLevelDebug
	LogLevelTrace

	DefLogLevel = LogLevelInfo
)

// NewLog creates a Log which is an implementation of the Logger interface
func NewLog(w io.Writer, prefix string, level Level) *Log {
	l := log.New(w, prefix, log.LstdFlags)

	return &Log{
		Prefix: prefix,
		Level:  level,
		log:    l,
	}
}

// Printf implements a standard Printf function of Logger interface
func (l *Log) Printf(format string, args ...interface{}) {
	l.output("     ", format, args...)
}

// Debugf implements a standard Debugf function of Logger interface
func (l *Log) Debugf(format string, args ...interface{}) {
	if l.Level < LogLevelDebug {
		return
	}
	if strings.HasPrefix(format, TracePrefix) {
		l.tracef(format, args...)
		return
	}
	l.output("DEBUG", format, args...)
}

// tracef is a second level of debug, it's call by Debugf when TraceLevel is set
func (l *Log) tracef(format string, args ...interface{}) {
	if l.Level < LogLevelTrace {
		return
	}
	format = strings.TrimPrefix(format, TracePrefix)
	l.output("TRACE", format, args...)
}

// Infof implements a standard Infof function of Logger interface
func (l *Log) Infof(format string, args ...interface{}) {
	if l.Level < LogLevelInfo {
		return
	}
	l.output("INFO ", format, args...)
}

// Warnf implements a standard Warnf function of Logger interface
func (l *Log) Warnf(format string, args ...interface{}) {
	if l.Level < LogLevelWarn {
		return
	}
	l.output("WARN ", format, args...)
}

// Errorf implements a standard Errorf function of Logger interface
func (l *Log) Errorf(format string, args ...interface{}) {
	if l.Level < LogLevelError {
		return
	}
	l.output("ERROR", format, args...)
}

func (l *Log) output(levelStr string, format string, args ...interface{}) {
	p := l.Prefix
	if len(p) != 0 {
		p = fmt.Sprintf("] %s: ", p)
	} else {
		p = "] "
	}
	l.log.SetPrefix(levelStr + " [ ")
	l.log.Printf(p+format, args...)
}

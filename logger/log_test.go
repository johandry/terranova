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
	"bytes"
	"log"
	"reflect"
	"regexp"
	"testing"
)

func TestLogOutput(t *testing.T) {
	rDate := regexp.MustCompile(`[0-9][0-9][0-9][0-9]/[0-9][0-9]/[0-9][0-9] [0-9][0-9]:[0-9][0-9]:[0-9][0-9]`)

	type logArgs struct {
		prefix string
		level  Level
	}
	type input struct {
		name   string
		level  Level
		format string
		args   []interface{}
		wantW  string
	}
	tests := []struct {
		name    string
		args    logArgs
		wantLog *Log
		inputs  []input
	}{
		{"log with error level", logArgs{"", LogLevelError}, &Log{"", LogLevelError, log.New(&bytes.Buffer{}, "", log.LstdFlags)}, []input{
			{"simple error message", LogLevelError, "test", []interface{}{}, "ERROR [ <DATE> ] test"},
			{"simple warning message", LogLevelWarn, "test", []interface{}{}, ""},
			{"simple info message", LogLevelInfo, "test", []interface{}{}, ""},
			{"simple debug message", LogLevelDebug, "test", []interface{}{}, ""},
			{"simple trace message", LogLevelTrace, "[TRACE] test", []interface{}{}, ""},
			{"simple message", Level(100), "test", []interface{}{}, "      [ <DATE> ] test"},

			{"formatted error message", LogLevelError, "test %s %d", []interface{}{"this", 1}, "ERROR [ <DATE> ] test this 1"},
			{"formatted message", Level(100), "test %s %d", []interface{}{"me", 0}, "      [ <DATE> ] test me 0"},
		}},
		{"log with warn level", logArgs{"", LogLevelWarn}, &Log{"", LogLevelWarn, log.New(&bytes.Buffer{}, "", log.LstdFlags)}, []input{
			{"simple error message", LogLevelError, "test", []interface{}{}, "ERROR [ <DATE> ] test"},
			{"simple warning message", LogLevelWarn, "test", []interface{}{}, "WARN  [ <DATE> ] test"},
			{"simple info message", LogLevelInfo, "test", []interface{}{}, ""},
			{"simple debug message", LogLevelDebug, "test", []interface{}{}, ""},
			{"simple trace message", LogLevelTrace, "[TRACE] test", []interface{}{}, ""},
			{"simple message", Level(100), "test", []interface{}{}, "      [ <DATE> ] test"},

			{"formatted error message", LogLevelError, "test %s %d", []interface{}{"this", 1}, "ERROR [ <DATE> ] test this 1"},
			{"formatted warn message", LogLevelWarn, "test %s %d", []interface{}{"this", 1}, "WARN  [ <DATE> ] test this 1"},
			{"formatted message", Level(100), "test %s %d", []interface{}{"me", 0}, "      [ <DATE> ] test me 0"},
		}},
		{"log with info level", logArgs{"", LogLevelInfo}, &Log{"", LogLevelInfo, log.New(&bytes.Buffer{}, "", log.LstdFlags)}, []input{
			{"simple error message", LogLevelError, "test", []interface{}{}, "ERROR [ <DATE> ] test"},
			{"simple warning message", LogLevelWarn, "test", []interface{}{}, "WARN  [ <DATE> ] test"},
			{"simple info message", LogLevelInfo, "test", []interface{}{}, "INFO  [ <DATE> ] test"},
			{"simple debug message", LogLevelDebug, "test", []interface{}{}, ""},
			{"simple trace message", LogLevelTrace, "[TRACE] test", []interface{}{}, ""},
			{"simple message", Level(100), "test", []interface{}{}, "      [ <DATE> ] test"},

			{"formatted error message", LogLevelError, "test %s %d", []interface{}{"this", 1}, "ERROR [ <DATE> ] test this 1"},
			{"formatted warn message", LogLevelWarn, "test %s %d", []interface{}{"this", 1}, "WARN  [ <DATE> ] test this 1"},
			{"formatted info message", LogLevelInfo, "test %s %d", []interface{}{"this", 1}, "INFO  [ <DATE> ] test this 1"},
			{"formatted message", Level(100), "test %s %d", []interface{}{"me", 0}, "      [ <DATE> ] test me 0"},
		}},
		{"log with debug level", logArgs{"", LogLevelDebug}, &Log{"", LogLevelDebug, log.New(&bytes.Buffer{}, "", log.LstdFlags)}, []input{
			{"simple error message", LogLevelError, "test", []interface{}{}, "ERROR [ <DATE> ] test"},
			{"simple warning message", LogLevelWarn, "test", []interface{}{}, "WARN  [ <DATE> ] test"},
			{"simple info message", LogLevelInfo, "test", []interface{}{}, "INFO  [ <DATE> ] test"},
			{"simple debug message", LogLevelDebug, "test", []interface{}{}, "DEBUG [ <DATE> ] test"},
			{"simple trace message", LogLevelTrace, "[TRACE] test", []interface{}{}, ""},
			{"simple message", Level(100), "test", []interface{}{}, "      [ <DATE> ] test"},

			{"formatted error message", LogLevelError, "test %s %d", []interface{}{"this", 1}, "ERROR [ <DATE> ] test this 1"},
			{"formatted warn message", LogLevelWarn, "test %s %d", []interface{}{"this", 1}, "WARN  [ <DATE> ] test this 1"},
			{"formatted info message", LogLevelInfo, "test %s %d", []interface{}{"this", 1}, "INFO  [ <DATE> ] test this 1"},
			{"formatted debug message", LogLevelDebug, "test %s %d", []interface{}{"this", 1}, "DEBUG [ <DATE> ] test this 1"},
			{"formatted message", Level(100), "test %s %d", []interface{}{"me", 0}, "      [ <DATE> ] test me 0"},
		}},
		{"log with trace level", logArgs{"", LogLevelTrace}, &Log{"", LogLevelTrace, log.New(&bytes.Buffer{}, "", log.LstdFlags)}, []input{
			{"simple error message", LogLevelError, "test", []interface{}{}, "ERROR [ <DATE> ] test"},
			{"simple warning message", LogLevelWarn, "test", []interface{}{}, "WARN  [ <DATE> ] test"},
			{"simple info message", LogLevelInfo, "test", []interface{}{}, "INFO  [ <DATE> ] test"},
			{"simple debug message", LogLevelDebug, "test", []interface{}{}, "DEBUG [ <DATE> ] test"},
			{"simple trace message", LogLevelTrace, "[TRACE] test", []interface{}{}, "TRACE [ <DATE> ] test"},
			{"simple message", Level(100), "test", []interface{}{}, "      [ <DATE> ] test"},

			{"formatted error message", LogLevelError, "test %s %d", []interface{}{"this", 1}, "ERROR [ <DATE> ] test this 1"},
			{"formatted warn message", LogLevelWarn, "test %s %d", []interface{}{"this", 1}, "WARN  [ <DATE> ] test this 1"},
			{"formatted info message", LogLevelInfo, "test %s %d", []interface{}{"this", 1}, "INFO  [ <DATE> ] test this 1"},
			{"formatted debug message", LogLevelDebug, "test %s %d", []interface{}{"this", 1}, "DEBUG [ <DATE> ] test this 1"},
			{"formatted trace message", LogLevelTrace, "[TRACE] test %s %d", []interface{}{"this", 1}, "TRACE [ <DATE> ] test this 1"},
			{"formatted message", Level(100), "test %s %d", []interface{}{"me", 0}, "      [ <DATE> ] test me 0"},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &bytes.Buffer{}
			gotLog := NewLog(w, tt.args.prefix, tt.args.level)
			if !reflect.DeepEqual(gotLog, tt.wantLog) {
				t.Errorf("NewLog() = %v, want %v", gotLog, tt.wantLog)
			}
			var fn string
			for _, in := range tt.inputs {
				switch in.level {
				case LogLevelError:
					gotLog.Errorf(in.format, in.args...)
					fn = "Errorf"
				case LogLevelWarn:
					gotLog.Warnf(in.format, in.args...)
					fn = "Warnf"
				case LogLevelInfo:
					gotLog.Infof(in.format, in.args...)
					fn = "Infof"
				case LogLevelDebug:
					gotLog.Debugf(in.format, in.args...)
					fn = "Debugf"
				case LogLevelTrace:
					gotLog.Debugf(in.format, in.args...)
					fn = "Trace with Debugf"
				default:
					gotLog.Printf(in.format, in.args...)
					fn = "Printf"
				}
				gotW := w.String()
				gotW = rDate.ReplaceAllString(gotW, "<DATE>")
				if len(gotW) > 0 {
					gotW = gotW[0 : len(gotW)-1]
				}
				if gotW != in.wantW {
					t.Errorf("%s: Log.%s() = %v, want %v", in.name, fn, gotW, in.wantW)
				}
				w.Reset()
			}
		})
	}
}

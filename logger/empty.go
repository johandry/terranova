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

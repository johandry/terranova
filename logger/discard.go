package logger

import "io/ioutil"

// DiscardLog returns a Terranova logger which will discard all the terraform logs.
func DiscardLog() *Log {
	return NewLog(ioutil.Discard, "", LogLevelError)
}

package tlog

import (
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"testing"
)

func TestNewMiddleware(t *testing.T) {
	tnLogger := NewLog(os.Stdout, "", DefLogLevel)
	stdLogWriter := log.Writer()
	discardLogger := NewLog(ioutil.Discard, "DISCARD", LogLevelTrace)

	tests := []struct {
		name string
		l    Logger
		want *Middleware
	}{
		{"no logger", nil, &Middleware{
			log:        tnLogger,
			prevWriter: stdLogWriter,
		}},
		{"std logger", tnLogger, &Middleware{
			log:        tnLogger,
			prevWriter: stdLogWriter,
		}},
		{"discard logger", discardLogger, &Middleware{
			log:        discardLogger,
			prevWriter: stdLogWriter,
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewMiddleware(tt.l)
			defer got.Close()

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewMiddleware() = %v, want %v", got, tt.want)
			}
		})
	}
}

// func TestMiddlewarePrint(t *testing.T) {
// 	tests := []struct {
// 		name string
// 		l    Logger
// 		want *Middleware
// 	}{

// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if got := NewMiddleware(tt.l); !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("NewMiddleware() = %v, want %v", got, tt.want)
// 			}
// 		}
// 	}
// }

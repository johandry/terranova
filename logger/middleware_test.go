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
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"regexp"
	"strings"
	"testing"
)

func TestNewMiddleware(t *testing.T) {
	tnLogger := NewLog(os.Stdout, "", DefLogLevel)
	discardLogger := NewLog(ioutil.Discard, "DISCARD", LogLevelTrace)

	tests := []struct {
		name string
		l    Logger
		want *Middleware
	}{
		{"no logger", nil, &Middleware{
			log: tnLogger,
		}},
		{"std logger", tnLogger, &Middleware{
			log: tnLogger,
		}},
		{"discard logger", discardLogger, &Middleware{
			log: discardLogger,
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewMiddleware(tt.l)
			defer got.Close()

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewMiddleware() = %+v, want %+v", got, tt.want)
			}
		})
	}

	t.Run("sending no parameters", func(t *testing.T) {
		got := NewMiddleware()
		defer got.Close()
		want := &Middleware{
			log: tnLogger,
		}

		if !reflect.DeepEqual(got, want) {
			t.Errorf("NewMiddleware() = %+v, want %+v", got, want)
		}
	})

	t.Run("sending more than one parameters", func(t *testing.T) {
		got := NewMiddleware(discardLogger, tnLogger, nil)
		defer got.Close()
		want := &Middleware{
			log: discardLogger,
		}

		if !reflect.DeepEqual(got, want) {
			t.Errorf("NewMiddleware() = %+v, want %+v", got, want)
		}
	})
}

func TestMiddleware_IgnoreOwnLog(t *testing.T) {
	tfBuf := new(bytes.Buffer)
	logBuf := new(bytes.Buffer)

	check := func(owner, got, pattern string) {
		line := got
		if len(got) > 0 {
			line = got[0 : len(got)-1]
		}

		matched, err := regexp.MatchString(pattern, line)
		if err != nil {
			t.Fatal(owner+" pattern did not compile:", err)
		}
		if !matched {
			t.Errorf(owner+" Print = %v, want pattern %v", line, pattern)
		}
	}
	checkOutput := func(tfPattern, logPattern string) {
		check("Middleware/Terraform", tfBuf.String(), tfPattern)
		check("Log", logBuf.String(), logPattern)
	}

	tests := []struct {
		name       string
		tfPattern  string
		logPattern string
		f          func(lm *Middleware)
	}{
		{"TF captured, no log",
			"^DEBUG \\[ " + Rdate + " " + Rtime + " \\] this line is captured by the Log Middleware$",
			"",
			func(lm *Middleware) {
				lm.Start()

				mockTerraformLog("debug", "this line is captured by the Log Middleware")
			},
		},
		{"start, def log: no TF captured, log printed",
			"",
			"^" + Rdate + " " + Rtime + " \\[DEBUG\\] this line is not captured by the Log Middleware$",
			func(lm *Middleware) {
				lm.Start()
				log := log.New(logBuf, "", log.LstdFlags)

				log.Printf("[DEBUG] this line is not captured by the Log Middleware")
			},
		},
		{"def log, start: no TF captured, log printed",
			"",
			"^" + Rdate + " " + Rtime + " \\[DEBUG\\] this line is not captured by the Log Middleware$",
			func(lm *Middleware) {
				log := log.New(logBuf, "", log.LstdFlags)
				lm.Start()

				log.Printf("[DEBUG] this line is not captured by the Log Middleware")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLog := NewMockLog(tfBuf)
			lm := NewMiddleware(mockLog)
			defer lm.Close()
			defer tfBuf.Reset()
			defer logBuf.Reset()

			tt.f(lm)
			checkOutput(tt.tfPattern, tt.logPattern)
		})
	}
}

func TestMiddleware_IsEnabled(t *testing.T) {
	l := NewLog(ioutil.Discard, "DISCARD", LogLevelTrace)

	tests := []struct {
		name   string
		action string
		want   bool
	}{
		{"is enable before start?", "", false},
		{"is enable after close when not started?", "close", false},
		{"is enable after start?", "start", true},
		{"is enable after close?", "start,close", false},
		{"is enable after start when was closed?", "start,close,start", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lm := NewMiddleware(l)
			defer lm.Close()

			for _, a := range strings.Split(tt.action, ",") {
				switch a {
				case "start":
					lm.Start()
				case "close":
					lm.Close()
				}
			}

			if enabled := lm.IsEnabled(); enabled != tt.want {
				t.Errorf("IsEnabled() = %v, want pattern %v", enabled, tt.want)
			}
		})
	}
}

const (
	Rdate       = `[0-9][0-9][0-9][0-9]/[0-9][0-9]/[0-9][0-9]`
	Rtime       = `[0-9][0-9]:[0-9][0-9]:[0-9][0-9]`
	DatePattern = `${DATE}`
)

func TestMiddlewarePrint(t *testing.T) {
	tests := []struct {
		name           string
		tfLogEntryType string
		tfLogEntry     string
		wantPattern    string
	}{
		{"single line trace", "trace", mockTFLogEntries["trace"], fmt.Sprintf("TRACE [ %s ] %s", DatePattern, mockTFLogEntries["trace"])},
		// {"multiple line trace", "trace-ml", mockTFLogEntries["trace-ml"], fmt.Sprintf("TRACE [ %s ] %s", DatePattern, mockTFLogEntries["trace-ml"])},

		{"single line debug", "debug", mockTFLogEntries["debug"], fmt.Sprintf("DEBUG [ %s ] %s", DatePattern, mockTFLogEntries["debug"])},
		// {"multiple line debug", "debug-ml", mockTFLogEntries["debug-ml"], fmt.Sprintf("DEBUG [ %s ] %s", DatePattern, mockTFLogEntries["debug-ml"])},

		{"single line info", "info", mockTFLogEntries["info"], fmt.Sprintf("INFO  [ %s ] %s", DatePattern, mockTFLogEntries["info"])},
		// {"multiple line info", "info-ml", mockTFLogEntries["info-ml"], fmt.Sprintf("INFO  [ %s ] %s", DatePattern, mockTFLogEntries["info-ml"])},

		{"single line warn", "warn", mockTFLogEntries["warn"], fmt.Sprintf("WARN  [ %s ] %s", DatePattern, mockTFLogEntries["warn"])},
		// {"multiple line warn", "warn-ml", mockTFLogEntries["warn-ml"], fmt.Sprintf("WARN  [ %s ] %s", DatePattern, mockTFLogEntries["warn-ml"])},

		{"single line error", "error", mockTFLogEntries["error"], fmt.Sprintf("ERROR [ %s ] %s", DatePattern, mockTFLogEntries["error"])},
		// {"multiple line error", "error-ml", mockTFLogEntries["error-ml"], fmt.Sprintf("ERROR [ %s ] %s", DatePattern, mockTFLogEntries["error-ml"])},

		{"single line", "line", mockTFLogEntries["line"], fmt.Sprintf("      [ %s ] %s", DatePattern, mockTFLogEntries["line"])},
		// {"multiple line", "mline", mockTFLogEntries["mline"], fmt.Sprintf("      [ %s ] %s", DatePattern, mockTFLogEntries["mline"])},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := new(bytes.Buffer)

			mockLog := NewMockLog(buf)
			lm := NewMiddleware(mockLog)
			defer lm.Close()

			lm.Start()

			mockTerraformLog(tt.tfLogEntryType, tt.tfLogEntry)
			got := buf.String()

			line := got[0 : len(got)-1]
			pattern := "^" + tt.wantPattern + "$"
			pattern = strings.Replace(pattern, "[", `\[`, -1)
			pattern = strings.Replace(pattern, "]", `\]`, -1)
			pattern = strings.Replace(pattern, "(", `\(`, -1)
			pattern = strings.Replace(pattern, ")", `\)`, -1)
			pattern = strings.Replace(pattern, "*", `\*`, -1)
			pattern = strings.Replace(pattern, ".", `\.`, -1)
			pattern = strings.Replace(pattern, DatePattern, Rdate+" "+Rtime, -1)
			matched, err := regexp.MatchString(pattern, line)
			if err != nil {
				t.Fatal("pattern did not compile:", err)
			}
			if !matched {
				t.Errorf("Middleware Print = %v, want pattern %v", line, pattern)
			}
		})
	}
}

func mockTerraformLog(entryType, entry string) {
	switch entryType {
	case "trace":
		log.Printf("[TRACE] %s\n", entry)

	case "debug":
		log.Printf("[DEBUG] %s\n", entry)

	case "info":
		log.Printf("[INFO] %s\n", entry)

	case "warn":
		log.Printf("[WARN] %s\n", entry)

	case "error":
		log.Printf("[ERROR] %s\n", entry)

	default:
		log.Println(entry)
	}
}

var mockTFLogEntries = map[string]string{
	"trace": `terraform.NewContext: starting`,
	"trace-ml": `Completed graph transform *terraform.ConfigTransformer with new graph:
aws_instance.server - *terraform.NodeValidatableResource
------`,
	"debug": `ProviderTransformer: "aws_instance.server" (*terraform.NodeValidatableResource) needs provider.aws`,
	"debug-ml": `(graphTransformerMulti) Completed graph transform *terraform.ProviderConfigTransformer with new graph:
aws_instance.server - *terraform.NodeValidatableResource
provider.aws - *terraform.NodeApplyableProvider
var.c - *terraform.NodeRootVariable
var.key_name - *terraform.NodeRootVariable
------`,
	"info": `terraform: building graph: GraphTypeValidate`,
	"info-ml": `Completed graph transform *terraform.RootVariableTransformer with new graph:
aws_instance.server - *terraform.NodeValidatableResource
var.c - *terraform.NodeRootVariable
var.key_name - *terraform.NodeRootVariable
------`,
	"warn": `no schema is attached to aws_instance.server[0], so config references cannot be detected`,
	"warn-ml": `Provider "aws" produced an invalid plan for aws_instance.server[0], but we are tolerating it because it is using the legacy plugin SDK.
    The following problems may be the cause of any confusing errors from downstream operations:
      - .source_dest_check: planned value cty.True does not match config value cty.NullVal(cty.Bool)
      - .get_password_data: planned value cty.False does not match config value cty.NullVal(cty.Bool)
      - .ephemeral_block_device: attribute representing nested block must not be unknown itself; set nested attribute values to unknown instead
      - .network_interface: attribute representing nested block must not be unknown itself; set nested attribute values to unknown instead
      - .root_block_device: attribute representing nested block must not be unknown itself; set nested attribute values to unknown instead
			- .ebs_block_device: attribute representing nested block must not be unknown itself; set nested attribute values to unknown instead`,
	"error": `this is a fake error in one line`,
	"error-ml": `Provider "aws" failed to do something that is fake for aws_instance.server[0], here are a few lines just for testing.
    The following problems may be the cause of any confusing errors from downstream operations:
      - .ephemeral_block_device: attribute representing nested block must not be unknown itself; set nested attribute values to unknown instead
      - .network_interface: attribute representing nested block must not be unknown itself; set nested attribute values to unknown instead
      - .root_block_device: attribute representing nested block must not be unknown itself; set nested attribute values to unknown instead
			- .ebs_block_device: attribute representing nested block must not be unknown itself; set nested attribute values to unknown instead`,
	"line": `this is a single line with no tag, something unusual in Terraform`,
	"ml": `The following lines are an unusual example of multiple lines without tag:
      - .source_dest_check: planned value cty.True does not match config value cty.NullVal(cty.Bool)
      - .ebs_block_device: attribute representing nested block must not be unknown itself; set nested attribute values to unknown instead`,
}

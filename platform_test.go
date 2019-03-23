package terranova_test

import "github.com/johandry/terranova"

func newPlatform() (*terranova.Platform, error) {
	code := `
  variable "testing" {
    description = "Testing variable"
    default     = "testing"
  }`
	return terranova.New("", code)
}

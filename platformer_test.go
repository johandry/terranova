package platformer_test

import "github.com/johandry/platformer"

func newPlatform() (*platformer.Platformer, error) {
	code := `
  variable "testing" {
    description = "Testing variable"
    default     = "testing"
  }`
	return platformer.New("", code)
}

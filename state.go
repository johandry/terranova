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

package terranova

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"

	"github.com/hashicorp/terraform/states/statefile"
)

// WriteState takes a io.Writer as input to write the Terraform state
func (p *Platform) WriteState(w io.Writer) (*Platform, error) {
	sf := statefile.New(p.State, "", 0)
	return p, statefile.Write(sf, w)
}

// ReadState takes a io.Reader as input to read from it the Terraform state
func (p *Platform) ReadState(r io.Reader) (*Platform, error) {
	sf, err := statefile.Read(r)
	if err != nil {
		return p, err
	}
	p.State = sf.State
	return p, nil
}

// WriteStateToFile save the state of the Terraform state to a file
func (p *Platform) WriteStateToFile(filename string) (*Platform, error) {
	var state bytes.Buffer
	if _, err := p.WriteState(&state); err != nil {
		return p, err
	}
	return p, ioutil.WriteFile(filename, state.Bytes(), 0644)
}

// ReadStateFromFile will load the Terraform state from a file and assign it to the
// Platform state.
func (p *Platform) ReadStateFromFile(filename string) (*Platform, error) {
	file, err := os.Open(filename)
	defer file.Close()
	if err != nil {
		return p, err
	}
	return p.ReadState(file)
}

// PersistStateTo reads the state from the given file, if exists. Then will save
// the current state to the given file every time it changes during the Terraform
// actions.
func (p *Platform) PersistStateTo(filename string) (*Platform, error) {
	return p, nil
}

package terranova

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"

	"github.com/hashicorp/terraform/terraform"
)

// State return the Terraform state or return a new one if it doesn't exists
func (p *Platform) State() *terraform.State {
	if p.state == nil {
		p.state = terraform.NewState()
	}
	return p.state
}

// WriteState takes a io.Writer as input to write the Terraform state
func (p *Platform) WriteState(w io.Writer) error {
	if err := terraform.WriteState(p.State(), w); err != nil {
		return err
	}
	return nil
}

// ReadState takes a io.Reader as input to read from it the Terraform state
func (p *Platform) ReadState(r io.Reader) error {
	state, err := terraform.ReadState(r)
	if err != nil {
		return err
	}
	p.state = state
	return nil
}

// WriteStateFile save the state of the Terraform state to a file
func (p *Platform) WriteStateFile(filename string) error {
	var state bytes.Buffer
	if err := terraform.WriteState(p.State(), &state); err != nil {
		return err
	}
	if err := ioutil.WriteFile(filename, state.Bytes(), 0644); err != nil {
		return err
	}
	return nil
}

// ExistStateFile return true if the Terraform state file exists
func (p *Platform) ExistStateFile(filename string) bool {
	_, err := os.Stat(filename)
	return os.IsExist(err)
}

// ReadStateFile will load the Terraform state from a file and assign it to the
// Platform state. If the file is empty or fail to read it, and the current
// state is nil, it will assign a new empty state
func (p *Platform) ReadStateFile(filename string) ([]byte, error) {
	state, err := ioutil.ReadFile(filename)
	if err != nil {
		p.State()
		return state, err
	}
	if len(state) > 0 {
		buf := bytes.NewBuffer(state)
		tfState, err := terraform.ReadState(buf)
		if err != nil {
			return state, err
		}
		p.state = tfState
	} else {
		// State() will create a new empty state if it's nil
		p.State()
	}
	return state, nil
}

// WriteStateBuf save the state of the Terraform state to a buffer of bytes.
func (p *Platform) WriteStateBuf(state []byte) error {
	var buf bytes.Buffer
	if err := terraform.WriteState(p.State(), &buf); err != nil {
		return err
	}
	copy(state, buf.Bytes())
	return nil
}

// ReadStateBuf load the Terraform state from a buffer of bytes.
func (p *Platform) ReadStateBuf(state []byte) error {
	if len(state) > 0 {
		buf := bytes.NewBuffer(state)
		tfState, err := terraform.ReadState(buf)
		if err != nil {
			return err
		}
		p.state = tfState
	} else {
		// State() will create a new empty state if it's nil
		p.State()
	}
	return nil
}

package terranova

// BindVars binds the map of variables to the Platform variables, to be used
// by Terraform
func (p *Platform) BindVars(vars map[string]string) {
	for name, value := range vars {
		p.Var(name, value)
	}
}

// Var set a variable with it's value
func (p *Platform) Var(name, value string) {
	if len(p.vars) == 0 {
		p.vars = make(map[string]interface{})
	}
	p.vars[name] = value
}

// GetVar returns the value of the variable
func (p *Platform) GetVar(name string) string {
	return p.vars[name].(string)
}

// IsVarSet return true if the variable is set
func (p *Platform) IsVarSet(name string) bool {
	_, ok := p.vars[name]
	return ok
}

// Vars return all the Platform variables
func (p *Platform) Vars() map[string]interface{} {
	return p.vars
}

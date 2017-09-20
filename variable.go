package platformer

// BindVars binds the map of variables to the Platformer variables, to be used
// by Terraform
func (p *Platformer) BindVars(vars map[string]string) {
	for name, value := range vars {
		p.Var(name, value)
	}
}

// Var set a variable with it's value
func (p *Platformer) Var(name, value string) {
	if len(p.vars) == 0 {
		p.vars = make(map[string]interface{})
	}
	p.vars[name] = value
}

// GetVar returns the value of the variable
func (p *Platformer) GetVar(name string) string {
	return p.vars[name].(string)
}

// IsVarSet return true if the variable is set
func (p *Platformer) IsVarSet(name string) bool {
	_, ok := p.vars[name]
	return ok
}

// Vars return all the Platformer variables
func (p *Platformer) Vars() map[string]interface{} {
	return p.vars
}

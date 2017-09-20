# Platformer

## Introduction

Platformer is a Go Package that uses Terraform Go Packages instead of using the binary.

*IMPORTANT*: It have not been used in production yet, so use it _only_ for educational purposes to create your own Go library or application.

This project is still in research and I use it for education purposes only, some explanations here may not be accurate so, if you find something wrong, please, [email me](johandry@gmail.com) or create a Pull Request.

## Compatibility

One of the biggest issues with this library is the version of the vendors. This code uses `terraform` package version `0.10.3` (not the latest version) but the latest version of a provisioner may not be the correct for this terraform version. To avoid this versioning issues use a vendoring Go tool (i.e. `govendor`, `glide` or `dep`) to make sure you have the right version of all the imported packages.

## How to use the Terraform library

Make sure you have installed the package:

    go get -u github.com/hashicorp/terraform/terraform

The high level flow is like follows, and it is coded in `func (p *Platformer) Apply(destroy bool) error` in the file `platformer.go`

1. Create the Terraform Context with some configuration parameters
2. Create the execution plan for the previous context and refresh it.
3. Apply the changes

The configuration parameters (on step #1) are the following:

- List of providers
- List of provisioners
- Current State
- Variables
- Assign a storage module or directory where all the Terraform templates are

### Providers

The default list of providers I have identified are: `template` (github.com/terraform-providers/terraform-provider-template/template) and `null` (github.com/terraform-providers/terraform-provider-null/null). Check the function `func (p *Platformer) updateProviders() terraform.ResourceProviderResolver` on `provider.go` and also the entire file.

It's up to you to include more providers, all that your code requires. For example, if you will create an AWS platform or using the AWS resource in a Terraform template, you need to import `github.com/terraform-providers/terraform-provider-aws/aws` and add the provider in the same way as it's done on function `func (p *Platformer) AddProvisioner(name string, provisioner terraform.ResourceProvisioner) *Platformer`

### Provisioners

The default list of provisioners I've found are: `local-exec`, `remote-exec` and `file`, all of them are located in the Terraform library (github.com/hashicorp/terraform/builtin/provisioners). It may be required to add more provisioners, to do so run the code as in the function `func (p *Platformer) AddProvisioner(name string, provisioner terraform.ResourceProvisioner) *Platformer` on `provisioner.go` as well.

### Variables

Optionally, if your Terraform template contains them, you can add variables to the Terraform context. Provide a map of interfaces `map[string]interface{}` with all the variable/value pairs and assign it to the Terraform context to the variable `Variables`.

### Storage Module

To specify the Terraform template you can provide the path where all these templates (or maybe just one) are. Or, you can have the Terraform template embedded in your code. The former is done by passing the directory where all the templates are to `github.com/hashicorp/terraform/config/module`.`func NewTreeModule(name, dir string) (*Tree, error)`, create a `github.com/hashicorp/go-getter`.`FolderStorage` struct with the same directory and assign it to `StorageDir`, finally load the templates with the method `func (t *Tree) Load(s getter.Storage, mode GetMode) error` of the instance of `*Tree` returned from the previous `NewTreeModule`.

The second option is to embed the Terraform template code in your Go code. Save the template to a temporal file in a temporal directory. The next steps is the same as explained above. Do not forget to delete the temporal file and directory (you may use `defer` for this).

All these code is located in the function `func (p *Platformer) setModule() (*module.Tree, error)` on file `platformer.go`

### Current State

The current state need to be assigned to the Terraform context into the field `State` before any action with that context (i.e. Plan)

After the plan is applied the output is an state file. This state file need to be saved into memory or a file so the next time you apply another change to the platform, Terraform would know in what state it is.

Initially (if no previous state is provided) need to be created an empty state. That's done with the function `github.com/hashicorp/terraform/terraform`.`func NewState() *State`. If a state is provided you can load/read it with the function `func ReadState(src io.Reader) (*State, error)`.

After the changes are done, get the final state using the function `func WriteState(d *State, dst io.Writer) error` so you can assign it to a variable (save it to memory) or save it to a file or database.

## Examples

The git repository https://github.com/johandry/platformer-examples contain an example of how to create a few EC2 instances in AWS using this library.

More examples for other infrastructures/clouds will be added to the same directory.

## Sources

All this research was done initially form the Gist of Greg Osuri (https://gist.github.com/gosuri/a1233ad6197e45d670b3), then a lot was obtained from the Terraform documentation (https://godoc.org/github.com/hashicorp/terraform) and source code (https://github.com/hashicorp/terraform).

This is project is still on research so feel free to comment on [email](johandry@gmail.com) or open a Pull Request.

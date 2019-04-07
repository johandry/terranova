# Terranova

Terranova is a Go package that allow you to use the Terraform Go Packages instead of using the binary.

For more information about Terranova, refer to the blog post [Terranova: Using Terraform from Go](http://blog.johandry.com/post/terranova-terraform-from-go/)

## How to use the Terraform library

Make sure you have installed the package:

    go get -u github.com/hashicorp/terraform/terraform

The high level flow is like follows:

1. Create the Platform instance with the Terraform code to apply
2. Add the Terraform Providers used in the code
3. Add the Terraform Provisioner (if any) used in the code
4. Add the variables used in the code
5. Load the previous state of the infrastructure
6. Apply the changes
7. Save the final state of the infrastructure

The following short example is to create, scale or terminate AWS EC2 instances:

```go
package main

import (
  "log"
  "os"
  "strconv"

  "github.com/hashicorp/terraform/builtin/provisioners/file"
  "github.com/johandry/terranova"
  "github.com/terraform-providers/terraform-provider-aws/aws"
)

var code string

const stateFilename = "aws-ec2-ubuntu.tfstate"

func main() {
  count := 1
  keyName := "username"

  platform, err := terranova.NewPlatform(code).
    AddProvider("aws", aws.Provider()).
    AddProvisioner("file", file.Provisioner()).
    Var("count", strconv.Itoa(count)).
    Var("key_name", keyName).
    ReadStateFromFile(stateFilename)

  if err != nil {
    if os.IsNotExist(err) {
      log.Printf("[DEBUG] state file %s does not exists", stateFilename)
    } else {
      log.Fatalf("Fail to load the initial state of the platform from file %s. %s", stateFilename, err)
    }
  }

  terminate := (count == 0)
  if err := platform.Apply(terminate); err != nil {
    log.Fatalf("Fail to apply the changes to the platform. %s", err)
  }

  if _, err := platform.WriteStateFile(stateFilename); err != nil {
    log.Fatalf("Fail to save the final state of the platform to file %s. %s", stateFilename, err)
  }
}

func init() {
  code = `
  variable "count"    { default = 2 }
  variable "key_name" {}
  provider "aws" {
    region        = "us-west-2"
  }
  resource "aws_instance" "server" {
    instance_type = "t2.micro"
    ami           = "ami-6e1a0117"
    count         = "${var.count}"
    key_name      = "${var.key_name}"
  }
  provisioner "file" {
    content     = "ami used: ${self.ami}"
    destination = "/tmp/file.log"
  }
`
}

```

Read the same example with some improvements in [example/main.go](example/main.go) and the blog post [Terranova: Using Terraform from Go](http://blog.johandry.com/post/terranova-terraform-from-go/) for a detail explanation of the code.

The git repository https://github.com/johandry/terranova-examples contain more examples of how to Terranova with different clouds or providers.

## Sources

All this research was done reading the Terraform documentation (https://godoc.org/github.com/hashicorp/terraform) and source code (https://github.com/hashicorp/terraform).

Feel free to comment or open a Pull Request.

## TODO

- [ ] Hooks
- [ ] Logs
- [ ] Output
- [ ] Testing
- [ ] CI/CD
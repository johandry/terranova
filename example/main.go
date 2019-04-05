package main

import (
  "log"
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

	platform := terranova.NewPlatform(code)
							.AddProvider("aws", aws.Provider())
							.AddProvisioner("file", file.Provisioner())
							.Var("count", strconv.Itoa(count))
							.Var("key_name", keyName)
  
  if _, err := platform.ReadStateFile(stateFilename); err != nil {
    log.Panicf("Fail to load the initial state of the platform from file %s. %s", stateFilename, err)
  }

  terminate := (count == 0)
  if err := platform.Apply(terminate); err != nil {
    log.Fatalf("Fail to apply the changes to the platform. %s", err)
  }

  if err := platform.WriteStateFile(stateFilename); err != nil {
    log.Fatalf("Fail to save the final state of the platform to file %s. %s", stateFilename, err)
  }
}

func init() {
  code = `
  variable "count" 		{ default = 2 }
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
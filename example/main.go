package main

import (
	"flag"
	"log"
	"strconv"

	"github.com/hashicorp/terraform/builtin/provisioners/file"
	"github.com/johandry/terranova"
	"github.com/terraform-providers/terraform-provider-aws/aws"
)

var (
	code    string
	count   int
	keyName string
)

const stateFilename = "aws-ec2-ubuntu.tfstate"

func main() {
	flag.Parse()

	if len(keyName) == 0 {
		log.Fatalf("key name is required. Create or use an existing key name assigned to your AWS account")
	}

	if count < 0 {
		log.Fatalf("count cannot be negative. It has to be '0' to terminate all the creted instances or the desired number of instances")
	}

	platform := terranova.NewPlatform(code).
		AddProvider("aws", aws.Provider()).
		AddProvisioner("file", file.Provisioner()).
		Var("count", strconv.Itoa(count)).
		Var("key_name", keyName).
		ReadStateFromFile(stateFilename)

	if err != nil {
		log.Fatalf("Fail to load the initial state of the platform from file %s. %s", stateFilename, err)
	}

	if err := platform.Apply((count == 0)); err != nil {
		log.Fatalf("Fail to apply the changes to the platform. %s", err)
	}

	if err := platform.WriteStateToFile(stateFilename); err != nil {
		log.Fatalf("Fail to save the final state of the platform to file %s. %s", stateFilename, err)
	}
}

func init() {
	flag.IntVar(&count, "count", 2, "number of instances to create. Set to '0' to terminate them all.")
	flag.StringVar(&keyName, "keyname", "", "keyname to access the instances.")

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

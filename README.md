# Terranova

Terranova is a Go package that allow you to use the Terraform Go Packages instead of using the binary. It works with the Terraform version `0.11.9`.

For more information about Terranova and how to use use it, refer to the blog post [Terranova: Using Terraform from Go](http://blog.johandry.com/post/terranova-terraform-from-go/)

## How to use the Terranova package

Terranova works better as a Go module, if you don't have a `go.mod` file in your project, create it with `go mod init [package full name]`. Import Terranova in the Go code:

```go
import (
  "github.com/johandry/terranova"
)
```

As soon as you execute a Go command such as `go build` or `go test` it will be included in your `go.mod` file and downloaded.

If you are not using modules yet, using vendors or having the packages in `$GOPATH`, please, `git clone` the repository and create the vendor directory:

```bash
mkdir -p $GOPATH/src/github.com/johandry/
cd $GOPATH/src/github.com/johandry/
git clone --depth=1 https://github.com/johandry/terranova.git
GO111MODULE=on go mod vendor
```

After having the package, the high level use of Terranova is like follows:

1. Create a *Platform* instance with the Terraform *code* to apply
2. Get (`go get`), import and add (`AddProvider()`) the Terraform *Provider(s)* used in the code
3. Get (`go get`), import and add (`AddProvisioner()`) the  the Terraform *Provisioner* (if any) used in the Terraform code
4. Add (`Var()`) the *variables* used in the Terraform code
5. Load the previous *state* of the infrastructure using `ReadStateFromFile()` or `ReadState()` methods.
6. *Apply* the changes using the method `Apply()`
7. Save the final *state* of the infrastructure using `WriteStateToFile()` or `WriteState()` methods.

The following example shows how to create, scale or terminate AWS EC2 instances:

```go
package main

import (
  "log"
  "os"

  "github.com/johandry/terranova"
  "github.com/terraform-providers/terraform-provider-aws/aws"
)

var code string

const stateFilename = "simple.tfstate"

func main() {
  count := 1
  keyName := "demo"

  platform, err := terranova.NewPlatform(code).
    AddProvider("aws", aws.Provider()).
    Var("count", count).
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

  if _, err := platform.WriteStateToFile(stateFilename); err != nil {
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
`
}
```

Read the same example at [terranova-examples/aws/simple/main.go](https://github.com/johandry/terranova-examples/blob/master/aws/simple/main.go) and the blog post [Terranova: Using Terraform from Go](http://blog.johandry.com/post/terranova-terraform-from-go/) for a detail explanation of the code.

The git repository [terranova-examples](https://github.com/johandry/terranova-examples) contain more examples of how to use Terranova with different clouds or providers.

## Sources

All this research was done reading the [Terraform documentation](https://godoc.org/github.com/hashicorp/terraform) and [source code](https://github.com/hashicorp/terraform).

Please, feel free to comment or open Pull Requests, help us to improve Terranova.

## TODO

- [ ] Create testing files
- [ ] Create CI/CD pipeline
- [ ] Implement Terraform Hooks
- [ ] Implement Logs
- [ ] Implement Stats
- [ ] Implement Code Validation
- [ ] Release Go module version `0.1.0`
- [ ] Implement Output interface

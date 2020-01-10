[![Actions Status](https://github.com/johandry/terranova/workflows/Unit%20Test/badge.svg)](https://github.com/johandry/terranova/actions) [![Build Status](https://travis-ci.org/johandry/terranova.svg?branch=master)](https://travis-ci.org/johandry/terranova) [![codecov](https://codecov.io/gh/johandry/terranova/branch/master/graph/badge.svg)](https://codecov.io/gh/johandry/terranova) [![GoDoc](https://godoc.org/github.com/johandry/terranova?status.svg)](https://godoc.org/github.com/johandry/terranova) [![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

# Terranova

Terranova is a Go package that allows you to easily use the Terraform Go Packages instead of executing the Terraform binary. The version `v0.0.2` of Terranova works with Terraform version `v0.12.17`.

For more information about Terranova and how to use use it, refer to the blog post [Terranova: Using Terraform from Go](http://blog.johandry.com/post/terranova-terraform-from-go/).

## How to use the Terranova package

Terranova works better as a Go module, if you don't have a `go.mod` file in your project, create it with `go mod init [package full name]`. Import Terranova in the Go code:

```go
import (
  "github.com/johandry/terranova"
)
```

As soon as you execute a Go command such as `go build`, `go test` or `go mod tidy` it will be included in your `go.mod` file and downloaded.

If you are not using modules yet, using vendors or having the packages in `$GOPATH`, please, `git clone` the repository and create the vendor directory:

```bash
mkdir -p $GOPATH/src/github.com/johandry/
cd $GOPATH/src/github.com/johandry/
git clone --depth=1 https://github.com/johandry/terranova.git
GO111MODULE=on go mod vendor
```

After having the package, the high level use of Terranova is like follows:

1. Create a *Platform* instance with the Terraform *code* to apply
2. Add to the `go.mod` file, import and add (`AddProvider()`) the Terraform *Provider(s)* used in the code
3. Add to the `go.mod` file, import and add (`AddProvisioner()`) the Terraform *Provisioner* (if any) used in the Terraform code.
4. Add (`Var()` or `BindVars()`) the *variables* used in the Terraform code.
5. (*optional*) Create (`NewMiddleware()`) a logger middleware with the default logger or a custom logger that implements the `Logger` interface.
6. (*optional*) Create your custom Terraform Hooks and assign them to the *Platform* instance.
7. Load the previous *state* of the infrastructure and keep it updated using `PersistStateToFile()`.
8. *Apply* the changes using the method `Apply()`.

The following example shows how to create, scale or terminate AWS EC2 instances:

```go
package main

import (
  "log"
  "os"

  "github.com/johandry/terranova"
  "github.com/johandry/terranova/logger"
  "github.com/terraform-providers/terraform-provider-aws/aws"
)

var code string

const stateFilename = "simple.tfstate"

func main() {
  count := 0
  keyName := "demo"

  log := log.New(os.Stderr, "", log.LstdFlags)
  logMiddleware := logger.NewMiddleware()
  defer logMiddleware.Close()

  platform, err := terranova.NewPlatform(code).
    SetMiddleware(logMiddleware).
    AddProvider("aws", aws.Provider()).
    Var("c", count).
    Var("key_name", keyName).
    PersistStateToFile(stateFilename)

  if err != nil {
    log.Fatalf("Fail to create the platform using state file %s. %s", stateFilename, err)
  }

  terminate := (count == 0)
  if err := platform.Apply(terminate); err != nil {
    log.Fatalf("Fail to apply the changes to the platform. %s", err)
  }
}

func init() {
  code = `
  variable "c"    { default = 2 }
  variable "key_name" {}
  provider "aws" {
    region        = "us-west-2"
  }
  resource "aws_instance" "server" {
    instance_type = "t2.micro"
    ami           = "ami-6e1a0117"
    count         = "${var.c}"
    key_name      = "${var.key_name}"
  }
`
}
```

Read the same example at [terranova-examples/aws/simple/main.go](https://github.com/johandry/terranova-examples/blob/master/aws/simple/main.go) and the blog post [Terranova: Using Terraform from Go](http://blog.johandry.com/post/terranova-terraform-from-go/) for a detail explanation of the code.

The git repository [terranova-examples](https://github.com/johandry/terranova-examples) contain more examples of how to use Terranova with different clouds or providers.

## Providers version

Terranova works with the latest version of Terraform (`v0.12.12`) but requires Terraform providers using the Legacy Terraform Plugin SDK instead of the newer Terraform Plugin SDK. If the required provider still uses the Legacy Terraform Plugin SDK select the latest release using the Terraform Plugin SDK. For more information read the [Terraform Plugin SDK page in the Extending Terraform documentation](https://www.terraform.io/docs/extend/plugin-sdk.html).

These are the latest versions supported for some providers:

- **AWS**:   `github.com/terraform-providers/terraform-provider-aws v1.60.1-0.20191003145700-f8707a46c6ec`
- **OpenStack**: `github.com/terraform-providers/terraform-provider-openstack v1.23.0`
- **vSphere**: `github.com/terraform-providers/terraform-provider-vsphere v1.13.0`
- **Azure**: `github.com/terraform-providers/terraform-provider-azurerm v1.34.0`
- **Null**: `github.com/terraform-providers/terraform-provider-null v1.0.1-0.20190430203517-8d3d85a60e20`
- **Template**: `github.com/terraform-providers/terraform-provider-template v1.0.1-0.20190501175038-5333ad92003c`
- **TLS**: `github.com/terraform-providers/terraform-provider-tls v1.2.1-0.20190816230231-0790c4b40281`

 Include the code line in the `require` section of your `go.mod` file for the provider you are importing, for example:

```go
require (
  github.com/hashicorp/terraform v0.12.12
  github.com/johandry/terranova v0.0.3
  github.com/terraform-providers/terraform-provider-openstack v1.23.0
  github.com/terraform-providers/terraform-provider-vsphere v1.13.0
)
```

If you get an error for a provider that you are not directly using (i.e. TLS provider) include the required version in the `replace` section of the `go.mod` file, for example:

```go
replace (
  github.com/terraform-providers/terraform-provider-tls => github.com/terraform-providers/terraform-provider-tls v1.2.1-0.20190816230231-0790c4b40281
)
```

If the required provider is not in this list, you can identify the version like this:

1. Go to the GitHub project of the provider
2. Locate the `provider.go` file located in the directory named like the provider name (i.e. `was/provider.go`)
3. If the file imports `terraform-plugin-sdk/terraform` go to the previous git tag.
4. Repeat step #3 until you find the latest tag importing `terraform/terraform` (i.e. the latest tag for AWS is `v2.31.0`)
5. If the tag can be used in the `go.mod` file, just include it after the module name, for example: `terraform-provider-vsphere v1.13.0`
6. If the tag cannot be used in the `go.mod` file, for example `v2.31.0` for AWS:
   1. Find the release generated from the git tag (i.e. v2.31.0)
   2. Locate the hash number assigned to the release (i.e. `f8707a4`)
   3. Include the number in the `go.mod` file and execute `go mod tidy` or any other go command such as `go test` or `go build`.

If this is not working and need help to identify the right version, open an issue and you'll be help as soon as possible.

## Logs

Using Terranova without a Log Middleware cause to print to Stderr all the Terraform logs, a lot, a lot of lines including traces. This may be uncomfortable or not needed. To discard the Terraform logs or filter them (print only the required log entries) you have to use the Log Middleware and (optionally) a custom Logger.

To create a Log Middleware use:

```go
logMiddleware := logger.NewMiddleware()
defer logMiddleware.Close()

...

platform.SetMiddleware(logMiddleware)
```

You can decide when the Log Middleware starts intercepting the standard `log` with `logMiddleware.Start()`, if you don't the Log Middleware will start intercepting every line printed by the standard `log` when Terranova execute an action that makes Terraform to print something to the standard `log`. Every line intercepted by the Log Middleware is printed by the provided logger. This hijack ends when the Log Middleware is closed. To make the platform use the middleware, add it with `SetMiddleware()` to the platform.

A logger is an instance of the interface `Logger`. If the Log Middleware is created without parameter the default logger will be used, it prints the INFO, WARN and ERROR log entries of Terraform. To create your own logger check the examples in the [Terranova Examples](https://github.com/johandry/terranova-examples/tree/master/custom-logs) repository.

**IMPORTANT**: It's recommended to create your own instance of `log` to not use the standard log when the Log Middleware is in use. Everything that is printed using the standard log will be intercepted by the Log Middleware and processed by the Logger. So, use your own custom log or do something like this before creating the Log Middleware:

```go
log := log.New(os.Stderr, "", log.LstdFlags)
...
log.Printf("[DEBUG] this line is not captured by the Log Middleware")
```

## Sources

All this research was done reading the [Terraform documentation](https://godoc.org/github.com/hashicorp/terraform) and [source code](https://github.com/hashicorp/terraform).

Please, feel free to comment, open Issues and Pull Requests, help us to improve Terranova.

## Contribute and Support

A great way to contribute to the project is to send a detailed report when you encounter an issue/bug or when you are requesting a new feature. We always appreciate a well-written, thorough bug report, and will thank you for it!

The [GitHub issue](https://github.com/johandry/terranova/issues) tracker is for bug reports and feature requests. If you already forked and modified Terranova with a new feature or fix, please, open a [Pull Request at GitHub](https://github.com/johandry/terranova/pulls)

General support can be found at Gopher's Slack [#terraform](slack://channel?team=T029RQSE6&id=CS75NMT9P) channel or message me directly [@johandry](slack://channel?team=T029RQSE6&id=D5W0Q8CKW)

## Stars over time

[![Stargazers over time](https://starchart.cc/johandry/terranova.svg)](https://starchart.cc/johandry/terranova)

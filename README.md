# Terranova

Terranova is a Go package that allow you to use the Terraform Go Packages instead of using the binary.

For more information about Terranova, refer to this [blog post](http://blog.johandry.com/post/terranova-terraform-from-go/)

## Compatibility

One of the biggest issues with this library is the version of the vendors. This code uses `terraform` package version `0.10.3` (not the latest version) but the latest version of a provisioner may not be the correct for this terraform version. To avoid this versioning issues use a vendoring Go tool (i.e. `govendor`, `glide` or `mod`) to make sure you have the right version of all the imported packages.

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

Read the example located in [example/main.go](example/main.go) and the [blog post](http://blog.johandry.com/post/terranova-terraform-from-go/)

## Examples

The git repository https://github.com/johandry/terranova-examples contain more examples of how to Terranova.

## Sources

All this research was done from the Terraform documentation (https://godoc.org/github.com/hashicorp/terraform) and source code (https://github.com/hashicorp/terraform).

Feel free to comment or open a Pull Request.

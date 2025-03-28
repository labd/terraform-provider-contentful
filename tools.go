//go:build tools
// +build tools

package main

import (
	// document generation

	_ "github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs"
	_ "github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen"
)

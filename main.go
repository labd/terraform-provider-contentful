package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/flaconi/terraform-provider-contentful/contentful"
	"github.com/flaconi/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6/tf6server"
	"github.com/hashicorp/terraform-plugin-mux/tf5to6server"
	"github.com/hashicorp/terraform-plugin-mux/tf6muxserver"
	"log"
)

// Run "go generate" to format example terraform files and generate the docs for the registry/website

// If you do not have terraform installed, you can remove the formatting command, but its suggested to
// ensure the documentation is formatted properly.
//go:generate terraform fmt -recursive ./examples/

// Run the docs generation tool, check its repository for more information on how it works and how docs
// can be customized.
//
//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs

var (
	// these will be set by the goreleaser configuration
	// to appropriate values for the compiled binary
	version string = "dev"
	commit  string = "snapshot"
)

func main() {

	debugFlag := flag.Bool("debug", false, "Start provider in debug mode.")
	flag.Parse()

	fullVersion := fmt.Sprintf("%s (%s)", version, commit)

	sdkProvider := contentful.Provider()

	ctx := context.Background()

	upgradedSdkProvider, err := tf5to6server.UpgradeServer(
		ctx,
		sdkProvider().GRPCProvider,
	)

	if err != nil {
		log.Fatal(err)
	}

	providers := []func() tfprotov6.ProviderServer{
		providerserver.NewProtocol6(provider.New(fullVersion, *debugFlag)),
		func() tfprotov6.ProviderServer {
			return upgradedSdkProvider
		},
	}

	muxServer, err := tf6muxserver.NewMuxServer(ctx, providers...)

	if err != nil {
		log.Fatal(err)
	}

	var serveOpts []tf6server.ServeOpt
	if *debugFlag {
		serveOpts = append(serveOpts, tf6server.WithManagedDebug())
	}

	err = tf6server.Serve(
		"registry.terraform.io/flaconi/contentful",
		muxServer.ProviderServer,
		serveOpts...,
	)
	if err != nil {
		log.Fatal(err)
	}
}

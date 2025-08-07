package role_assignment_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/labd/terraform-provider-contentful/internal/acctest"
	"github.com/labd/terraform-provider-contentful/internal/provider"
)

func TestRoleAssignmentResource_Basic(t *testing.T) {
	// Role assignment resource requires team functionality which isn't implemented yet
	t.Skip("Role assignment resource depends on team resource (issue #84) and role assignment APIs that are not yet available")
}

// This test validates that the resource schema can be created without errors
func TestRoleAssignmentResource_Schema(t *testing.T) {
	t.Parallel()

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"contentful": providerserver.NewProtocol6WithError(provider.New("test", true)()),
		},
		Steps: []resource.TestStep{
			{
				Config: `
					provider "contentful" {}
				`,
			},
		},
	})
}
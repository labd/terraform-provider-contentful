package environment_alias_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	hashicor_acctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/assert"

	"github.com/labd/terraform-provider-contentful/internal/acctest"
	"github.com/labd/terraform-provider-contentful/internal/provider"
	"github.com/labd/terraform-provider-contentful/internal/sdk"
)

type assertFunc func(*testing.T, *sdk.EnvironmentAlias)

func TestEnvironmentAliasResource_Basic(t *testing.T) {
	aliasID := fmt.Sprintf("alias-%s", hashicor_acctest.RandString(3))
	resourceName := "contentful_environment_alias.test"
	spaceID := os.Getenv("CONTENTFUL_SPACE_ID")

	// Create two test environments to switch between
	env1Name := fmt.Sprintf("env1-%s", hashicor_acctest.RandString(3))
	env2Name := fmt.Sprintf("env2-%s", hashicor_acctest.RandString(3))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.TestAccPreCheck(t) },
		CheckDestroy: testAccCheckContentfulEnvironmentAliasDestroy,
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"contentful": providerserver.NewProtocol6WithError(provider.New("test", true)()),
		},
		Steps: []resource.TestStep{
			{
				Config: testEnvironmentAliasConfig(spaceID, aliasID, env1Name, env2Name, "contentful_environment.env1.id"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "id", aliasID),
					resource.TestCheckResourceAttrPair(resourceName, "environment_id", "contentful_environment.env1", "id"),
					testAccCheckContentfulEnvironmentAliasExists(t, resourceName, func(t *testing.T, alias *sdk.EnvironmentAlias) {
						assert.EqualValues(t, aliasID, alias.Sys.Id)
					}),
				),
			},
			{
				Config: testEnvironmentAliasConfig(spaceID, aliasID, env1Name, env2Name, "contentful_environment.env2.id"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "id", aliasID),
					resource.TestCheckResourceAttrPair(resourceName, "environment_id", "contentful_environment.env2", "id"),
					testAccCheckContentfulEnvironmentAliasExists(t, resourceName, func(t *testing.T, alias *sdk.EnvironmentAlias) {
						assert.EqualValues(t, aliasID, alias.Sys.Id)
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     fmt.Sprintf("%s:ALIAS_ID_PLACEHOLDER", spaceID),
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources[resourceName]
					if !ok {
						return "", fmt.Errorf("not found: %s", resourceName)
					}
					return fmt.Sprintf("%s:%s", rs.Primary.Attributes["space_id"], rs.Primary.ID), nil
				},
			},
		},
	})
}

func testAccCheckContentfulEnvironmentAliasExists(t *testing.T, resourceName string, assertFunc assertFunc) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		alias, err := getEnvironmentAliasFromState(s, resourceName)
		if err != nil {
			return err
		}

		assertFunc(t, alias)
		return nil
	}
}

func getEnvironmentAliasFromState(s *terraform.State, resourceName string) (*sdk.EnvironmentAlias, error) {
	rs, ok := s.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("environment alias not found in state: %s", resourceName)
	}

	if rs.Primary.ID == "" {
		return nil, fmt.Errorf("no environment alias ID found")
	}

	spaceID := rs.Primary.Attributes["space_id"]
	if spaceID == "" {
		return nil, fmt.Errorf("no space_id is set")
	}

	aliasID := rs.Primary.Attributes["id"]
	if aliasID == "" {
		return nil, fmt.Errorf("no alias_id is set")
	}

	client := acctest.GetClient()
	resp, err := client.GetEnvironmentAliasWithResponse(context.Background(), spaceID, aliasID)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("environment alias not found: %s", aliasID)
	}

	return resp.JSON200, nil
}

func testAccCheckContentfulEnvironmentAliasDestroy(s *terraform.State) error {
	client := acctest.GetClient()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "contentful_environment_alias" {
			continue
		}

		spaceID := rs.Primary.Attributes["space_id"]
		if spaceID == "" {
			return fmt.Errorf("no space_id is set")
		}

		aliasID := rs.Primary.Attributes["id"]
		if aliasID == "" {
			return fmt.Errorf("no alias_id is set")
		}

		resp, err := client.GetEnvironmentAliasWithResponse(context.Background(), spaceID, aliasID)
		if err != nil {
			return err
		}

		if resp.StatusCode() == 404 {
			return nil
		}

		return fmt.Errorf("environment alias still exists with id: %s", aliasID)
	}

	return nil
}

func testEnvironmentAliasConfig(spaceID, aliasID, env1Name, env2Name, environmentIDRef string) string {
	return fmt.Sprintf(`
resource "contentful_environment" "env1" {
  space_id = "%s"
  name     = "%s"
}

resource "contentful_environment" "env2" {
  space_id = "%s"
  name     = "%s"
}

resource "contentful_environment_alias" "test" {
  space_id       = "%s"
  alias_id       = "%s"
  environment_id = %s
}
`, spaceID, env1Name, spaceID, env2Name, spaceID, aliasID, environmentIDRef)
}

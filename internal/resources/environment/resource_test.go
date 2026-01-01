package environment_test

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

type assertFunc func(*testing.T, *sdk.Environment)

func TestEnvironmentResource_Basic(t *testing.T) {
	name := fmt.Sprintf("env-%s", hashicor_acctest.RandString(3))
	updatedName := fmt.Sprintf("env-updated-%s", hashicor_acctest.RandString(3))
	resourceName := "contentful_environment.myenvironment"
	spaceID := os.Getenv("CONTENTFUL_SPACE_ID")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.TestAccPreCheck(t) },
		CheckDestroy: testAccCheckContentfulEnvironmentDestroy,
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"contentful": providerserver.NewProtocol6WithError(provider.New("test", true)()),
		},
		Steps: []resource.TestStep{
			{
				Config: testEnvironmentConfig(spaceID, name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", name),
					testAccCheckContentfulEnvironmentExists(t, resourceName, func(t *testing.T, env *sdk.Environment) {
						assert.EqualValues(t, name, env.Name)
					}),
				),
			},
			{
				Config: testEnvironmentConfig(spaceID, updatedName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", updatedName),
					testAccCheckContentfulEnvironmentExists(t, resourceName, func(t *testing.T, env *sdk.Environment) {
						assert.EqualValues(t, updatedName, env.Name)
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     fmt.Sprintf("%s:ENVIRONMENT_ID_PLACEHOLDER", spaceID),
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

func testAccCheckContentfulEnvironmentExists(t *testing.T, resourceName string, assertFunc assertFunc) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		env, err := getEnvironmentFromState(s, resourceName)
		if err != nil {
			return err
		}

		assertFunc(t, env)
		return nil
	}
}

func getEnvironmentFromState(s *terraform.State, resourceName string) (*sdk.Environment, error) {
	rs, ok := s.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("environment not found in state: %s", resourceName)
	}

	if rs.Primary.ID == "" {
		return nil, fmt.Errorf("no environment ID found")
	}

	spaceID := rs.Primary.Attributes["space_id"]
	if spaceID == "" {
		return nil, fmt.Errorf("no space_id is set")
	}

	client := acctest.GetClient()
	resp, err := client.GetEnvironmentWithResponse(context.Background(), spaceID, rs.Primary.ID)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("environment not found: %s", rs.Primary.ID)
	}

	return resp.JSON200, nil
}

func testAccCheckContentfulEnvironmentDestroy(s *terraform.State) error {
	client := acctest.GetClient()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "contentful_environment" {
			continue
		}

		spaceID := rs.Primary.Attributes["space_id"]
		if spaceID == "" {
			return fmt.Errorf("no space_id is set")
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no environment ID is set")
		}

		resp, err := client.GetEnvironmentWithResponse(context.Background(), spaceID, rs.Primary.ID)
		if err != nil {
			return err
		}

		if resp.StatusCode() == 404 {
			return nil
		}

		return fmt.Errorf("environment still exists with id: %s", rs.Primary.ID)
	}

	return nil
}

func testEnvironmentConfig(spaceID string, name string) string {
	return fmt.Sprintf(`
resource "contentful_environment" "myenvironment" {
  space_id = "%s"
  name = "%s"
}
`, spaceID, name)
}

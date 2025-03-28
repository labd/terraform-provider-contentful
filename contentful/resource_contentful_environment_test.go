package contentful

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/labd/terraform-provider-contentful/internal/acctest"
	"github.com/labd/terraform-provider-contentful/internal/sdk"
)

func TestAccContentfulEnvironment_Basic(t *testing.T) {
	var environment sdk.Environment

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccContentfulEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccContentfulEnvironmentConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContentfulEnvironmentExists("contentful_environment.myenvironment", &environment),
					testAccCheckContentfulEnvironmentAttributes(&environment, map[string]interface{}{
						"space_id": spaceID,
						"name":     "provider-test",
					}),
				),
			},
			{
				Config: testAccContentfulEnvironmentUpdateConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContentfulEnvironmentExists("contentful_environment.myenvironment", &environment),
					testAccCheckContentfulEnvironmentAttributes(&environment, map[string]interface{}{
						"space_id": spaceID,
						"name":     "provider-test-updated",
					}),
				),
			},
		},
	})
}

func testAccCheckContentfulEnvironmentExists(n string, environment *sdk.Environment) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ctx := context.Background()
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not Found: %s", n)
		}

		spaceID := rs.Primary.Attributes["space_id"]
		if spaceID == "" {
			return fmt.Errorf("no space_id is set")
		}

		name := rs.Primary.Attributes["name"]
		if name == "" {
			return fmt.Errorf("no name is set")
		}

		client := acctest.GetClient()

		resp, err := client.GetEnvironmentWithResponse(ctx, spaceID, rs.Primary.ID)
		if err != nil {
			return err
		}

		if resp.StatusCode() != 200 {
			return fmt.Errorf("environment not found: %s", rs.Primary.ID)
		}

		*environment = *resp.JSON200

		return nil
	}
}

func testAccCheckContentfulEnvironmentAttributes(environment *sdk.Environment, attrs map[string]interface{}) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		name := attrs["name"].(string)
		if environment.Name != name {
			return fmt.Errorf("environment name does not match: %s, %s", environment.Name, name)
		}

		return nil
	}
}

func testAccContentfulEnvironmentDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "contentful_environment" {
			continue
		}
		spaceID := rs.Primary.Attributes["space_id"]
		if spaceID == "" {
			return fmt.Errorf("no space_id is set")
		}

		environmentID := rs.Primary.ID
		if environmentID == "" {
			return fmt.Errorf("no environment ID is set")
		}

		client := acctest.GetClient()

		resp, err := client.GetEnvironmentWithResponse(context.Background(), spaceID, rs.Primary.ID)
		if err != nil {
			return err
		}

		if resp.StatusCode() == 404 {
			return nil
		}

		return fmt.Errorf("environment still exists with id: %s", environmentID)
	}

	return nil
}

var testAccContentfulEnvironmentConfig = `
resource "contentful_environment" "myenvironment" {
  space_id = "` + spaceID + `"
  name = "provider-test"
}
`

var testAccContentfulEnvironmentUpdateConfig = `
resource "contentful_environment" "myenvironment" {
  space_id = "` + spaceID + `"
  name = "provider-test-updated"
}
`

package contentful

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	contentful "github.com/labd/contentful-go"
)

func TestAccContentfulEnvironment_Basic(t *testing.T) {
	var environment contentful.Environment

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

func testAccCheckContentfulEnvironmentExists(n string, environment *contentful.Environment) resource.TestCheckFunc {
	return func(s *terraform.State) error {
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

		client := testAccProvider.Meta().(*contentful.Client)

		contentfulEnvironment, err := client.Environments.Get(spaceID, rs.Primary.ID)
		if err != nil {
			return err
		}

		*environment = *contentfulEnvironment

		return nil
	}
}

func testAccCheckContentfulEnvironmentAttributes(environment *contentful.Environment, attrs map[string]interface{}) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		name := attrs["name"].(string)
		if environment.Name != name {
			return fmt.Errorf("locale name does not match: %s, %s", environment.Name, name)
		}

		return nil
	}
}

func testAccContentfulEnvironmentDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "contentful_locale" {
			continue
		}
		spaceID := rs.Primary.Attributes["space_id"]
		if spaceID == "" {
			return fmt.Errorf("no space_id is set")
		}

		localeID := rs.Primary.ID
		if localeID == "" {
			return fmt.Errorf("no locale ID is set")
		}

		client := testAccProvider.Meta().(*contentful.Client)

		_, err := client.Locales.Get(spaceID, localeID)
		if _, ok := err.(contentful.NotFoundError); ok {
			return nil
		}

		return fmt.Errorf("locale still exists with id: %s", localeID)
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

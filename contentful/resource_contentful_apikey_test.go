package contentful

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	contentful "github.com/labd/contentful-go"
)

func TestAccContentfulAPIKey_Basic(t *testing.T) {
	var apiKey contentful.APIKey

	name := fmt.Sprintf("apikey-name-%s", acctest.RandString(3))
	description := fmt.Sprintf("apikey-description-%s", acctest.RandString(3))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccContentfulAPIKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccContentfulAPIKeyConfig(name, description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContentfulAPIKeyExists("contentful_apikey.myapikey", &apiKey),
					testAccCheckContentfulAPIKeyAttributes(&apiKey, map[string]interface{}{
						"space_id":    spaceID,
						"name":        name,
						"description": description,
					}),
				),
			},
			{
				Config: testAccContentfulAPIKeyUpdateConfig(name, description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContentfulAPIKeyExists("contentful_apikey.myapikey", &apiKey),
					testAccCheckContentfulAPIKeyAttributes(&apiKey, map[string]interface{}{
						"space_id":    spaceID,
						"name":        fmt.Sprintf("%s-updated", name),
						"description": fmt.Sprintf("%s-updated", description),
					}),
				),
			},
		},
	})
}

func testAccCheckContentfulAPIKeyExists(n string, apiKey *contentful.APIKey) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not Found: %s", n)
		}

		spaceID := rs.Primary.Attributes["space_id"]
		if spaceID == "" {
			return fmt.Errorf("no space_id is set")
		}

		apiKeyID := rs.Primary.ID
		if apiKeyID == "" {
			return fmt.Errorf("no api key ID is set")
		}

		client := testAccProvider.Meta().(*contentful.Client)

		contentfulAPIKey, err := client.APIKeys.Get(spaceID, apiKeyID)
		if err != nil {
			return err
		}

		*apiKey = *contentfulAPIKey

		return nil
	}
}

func testAccCheckContentfulAPIKeyAttributes(apiKey *contentful.APIKey, attrs map[string]interface{}) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		name := attrs["name"].(string)
		if apiKey.Name != name {
			return fmt.Errorf("APIKey name does not match: %s, %s", apiKey.Name, name)
		}

		description := attrs["description"].(string)
		if apiKey.Description != description {
			return fmt.Errorf("APIKey description does not match: %s, %s", apiKey.Description, description)
		}

		return nil
	}
}

func testAccContentfulAPIKeyDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "contentful_apikey" {
			continue
		}

		// get space id from resource data
		spaceID := rs.Primary.Attributes["space_id"]
		if spaceID == "" {
			return fmt.Errorf("no space_id is set")
		}

		apiKeyID := rs.Primary.ID
		if apiKeyID == "" {
			return fmt.Errorf("no apikey ID is set")
		}

		client := testAccProvider.Meta().(*contentful.Client)

		_, err := client.APIKeys.Get(spaceID, apiKeyID)
		if _, ok := err.(contentful.NotFoundError); ok {
			return nil
		}

		return fmt.Errorf("api Key still exists with id: %s", rs.Primary.ID)
	}

	return nil
}

func testAccContentfulAPIKeyConfig(name, description string) string {
	return fmt.Sprintf(`
resource "contentful_apikey" "myapikey" {
  space_id = "%s"

  name = "%s"
  description = "%s"
}
`, spaceID, name, description)
}

func testAccContentfulAPIKeyUpdateConfig(name, description string) string {
	return fmt.Sprintf(`
resource "contentful_apikey" "myapikey" {
  space_id = "%s"

  name = "%s-updated"
  description = "%s-updated"
}
`, spaceID, name, description)
}

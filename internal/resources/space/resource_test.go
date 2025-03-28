package space_test

import (
	"context"
	"fmt"
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

type assertFunc func(*testing.T, *sdk.Space)

func TestSpaceResource_Create(t *testing.T) {
	// Space resource requires organization admin permissions
	t.Skip("Space resource can only be tested when user has organization admin rights")

	name := fmt.Sprintf("space-test-%s", hashicor_acctest.RandString(5))
	updatedName := fmt.Sprintf("space-test-updated-%s", hashicor_acctest.RandString(5))
	resourceName := "contentful_space.myspace"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.TestAccPreCheck(t) },
		CheckDestroy: testAccCheckContentfulSpaceDestroy,
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"contentful": providerserver.NewProtocol6WithError(provider.New("test", true)()),
		},
		Steps: []resource.TestStep{
			{
				Config: testSpaceConfig(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "default_locale", "en"),
					testAccCheckContentfulSpaceExists(t, resourceName, func(t *testing.T, space *sdk.Space) {
						assert.EqualValues(t, name, space.Name)
					}),
				),
			},
			{
				Config: testSpaceConfig(updatedName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", updatedName),
					resource.TestCheckResourceAttr(resourceName, "default_locale", "en"),
					testAccCheckContentfulSpaceExists(t, resourceName, func(t *testing.T, space *sdk.Space) {
						assert.EqualValues(t, updatedName, space.Name)
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				// default_locale isn't returned in the API response, so we need to ignore it during import verification
				ImportStateVerifyIgnore: []string{"default_locale"},
			},
		},
	})
}

func testAccCheckContentfulSpaceExists(t *testing.T, resourceName string, assertFunc assertFunc) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		space, err := getSpaceFromState(s, resourceName)
		if err != nil {
			return err
		}

		assertFunc(t, space)
		return nil
	}
}

func getSpaceFromState(s *terraform.State, resourceName string) (*sdk.Space, error) {
	rs, ok := s.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("space not found in state: %s", resourceName)
	}

	if rs.Primary.ID == "" {
		return nil, fmt.Errorf("no space ID found")
	}

	client := acctest.GetClient()
	resp, err := client.GetSpaceWithResponse(context.Background(), rs.Primary.ID)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("space not found: %s", rs.Primary.ID)
	}

	return resp.JSON200, nil
}

func testAccCheckContentfulSpaceDestroy(s *terraform.State) error {
	client := acctest.GetClient()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "contentful_space" {
			continue
		}

		resp, err := client.GetSpaceWithResponse(context.Background(), rs.Primary.ID)
		if err != nil {
			// API error, likely the resource is already gone
			return nil
		}

		if resp.StatusCode() == 404 {
			// Resource properly deleted
			return nil
		}

		return fmt.Errorf("space %s still exists after destroy", rs.Primary.ID)
	}

	return nil
}

func testSpaceConfig(name string) string {
	return fmt.Sprintf(`
resource "contentful_space" "myspace" {
  name = "%s"
  default_locale = "en"
}
`, name)
}

func testSpaceConfigWithLocale(name string, locale string) string {
	return fmt.Sprintf(`
resource "contentful_space" "myspace" {
  name = "%s"
  default_locale = "%s"
}
`, name, locale)
}

package locale_test

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

type assertFunc func(*testing.T, *sdk.Locale)

func TestLocaleResource_Basic(t *testing.T) {
	name := fmt.Sprintf("locale-name-%s", hashicor_acctest.RandString(3))
	code := fmt.Sprintf("l%s", hashicor_acctest.RandString(2))
	resourceName := "contentful_locale.mylocale"
	spaceID := os.Getenv("CONTENTFUL_SPACE_ID")
	environment := "master"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.TestAccPreCheck(t) },
		CheckDestroy: testAccCheckContentfulLocaleDestroy,
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"contentful": providerserver.NewProtocol6WithError(provider.New("test", true)()),
		},
		Steps: []resource.TestStep{
			{
				Config: testLocaleConfig(spaceID, environment, name, code),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "code", code),
					resource.TestCheckResourceAttr(resourceName, "fallback_code", "en-US"),
					resource.TestCheckResourceAttr(resourceName, "optional", "false"),
					resource.TestCheckResourceAttr(resourceName, "cda", "false"),
					resource.TestCheckResourceAttr(resourceName, "cma", "true"),
					testAccCheckContentfulLocaleExists(t, resourceName, func(t *testing.T, locale *sdk.Locale) {
						assert.EqualValues(t, name, locale.Name)
						assert.EqualValues(t, code, locale.Code)
						assert.EqualValues(t, "en-US", *locale.FallbackCode)
						assert.EqualValues(t, false, locale.Optional)
						assert.EqualValues(t, false, locale.ContentDeliveryApi)
						assert.EqualValues(t, true, locale.ContentManagementApi)
					}),
				),
			},
			{
				Config: testLocaleUpdateConfig(spaceID, environment, name, code),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("%s-updated", name)),
					resource.TestCheckResourceAttr(resourceName, "code", code), // Code remains the same
					resource.TestCheckResourceAttr(resourceName, "fallback_code", "en-US"),
					resource.TestCheckResourceAttr(resourceName, "optional", "true"),
					resource.TestCheckResourceAttr(resourceName, "cda", "true"),
					resource.TestCheckResourceAttr(resourceName, "cma", "false"),
					testAccCheckContentfulLocaleExists(t, resourceName, func(t *testing.T, locale *sdk.Locale) {
						assert.EqualValues(t, fmt.Sprintf("%s-updated", name), locale.Name)
						assert.EqualValues(t, code, locale.Code)
						assert.EqualValues(t, "en-US", *locale.FallbackCode)
						assert.EqualValues(t, true, locale.Optional)
						assert.EqualValues(t, true, locale.ContentDeliveryApi)
						assert.EqualValues(t, false, locale.ContentManagementApi)
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     fmt.Sprintf("%s:%s:%s", "LOCALE_ID_PLACEHOLDER", environment, spaceID),
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources[resourceName]
					if !ok {
						return "", fmt.Errorf("not found: %s", resourceName)
					}
					return fmt.Sprintf("%s:%s:%s", rs.Primary.ID, rs.Primary.Attributes["environment"], rs.Primary.Attributes["space_id"]), nil
				},
			},
		},
	})
}

func testAccCheckContentfulLocaleExists(t *testing.T, resourceName string, assertFunc assertFunc) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		locale, err := getLocaleFromState(s, resourceName)
		if err != nil {
			return err
		}

		assertFunc(t, locale)
		return nil
	}
}

func getLocaleFromState(s *terraform.State, resourceName string) (*sdk.Locale, error) {
	rs, ok := s.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("locale not found in state: %s", resourceName)
	}

	if rs.Primary.ID == "" {
		return nil, fmt.Errorf("no locale ID found")
	}

	spaceID := rs.Primary.Attributes["space_id"]
	if spaceID == "" {
		return nil, fmt.Errorf("no space_id is set")
	}

	environment := rs.Primary.Attributes["environment"]
	if environment == "" {
		return nil, fmt.Errorf("no environment is set")
	}

	client := acctest.GetClient()
	resp, err := client.GetLocaleWithResponse(context.Background(), spaceID, environment, rs.Primary.ID)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("locale not found: %s", rs.Primary.ID)
	}

	return resp.JSON200, nil
}

func testAccCheckContentfulLocaleDestroy(s *terraform.State) error {
	client := acctest.GetClient()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "contentful_locale" {
			continue
		}

		spaceID := rs.Primary.Attributes["space_id"]
		if spaceID == "" {
			return fmt.Errorf("no space_id is set")
		}

		environment := rs.Primary.Attributes["environment"]
		if environment == "" {
			return fmt.Errorf("no environment is set")
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no locale ID is set")
		}

		resp, err := client.GetLocaleWithResponse(context.Background(), spaceID, environment, rs.Primary.ID)
		if err != nil {
			return err
		}

		if resp.StatusCode() == 404 {
			return nil
		}

		return fmt.Errorf("locale still exists with id: %s", rs.Primary.ID)
	}

	return nil
}

func testLocaleConfig(spaceID, environment, name, code string) string {
	return fmt.Sprintf(`
resource "contentful_locale" "mylocale" {
  space_id = "%s"
  environment = "%s"
  name = "%s"
  code = "%s"
  fallback_code = "en-US"
  optional = false
  cda = false
  cma = true
}
`, spaceID, environment, name, code)
}

func testLocaleUpdateConfig(spaceID, environment, name, code string) string {
	return fmt.Sprintf(`
resource "contentful_locale" "mylocale" {
  space_id = "%s"
  environment = "%s"
  name = "%s-updated"
  code = "%s"
  fallback_code = "en-US"
  optional = true
  cda = true
  cma = false
}
`, spaceID, environment, name, code)
}

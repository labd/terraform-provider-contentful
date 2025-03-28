package app_definition_test

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/assert"

	"github.com/labd/terraform-provider-contentful/internal/acctest"
	"github.com/labd/terraform-provider-contentful/internal/provider"
	"github.com/labd/terraform-provider-contentful/internal/sdk"
	"github.com/labd/terraform-provider-contentful/internal/utils"
)

type assertFunc func(*testing.T, *sdk.AppDefinition)

func TestAppDefinitionResource_Create(t *testing.T) {
	resourceName := "contentful_app_definition.acctest_app_definition"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.TestAccPreCheck(t) },
		CheckDestroy: testAccCheckContentfulAppDefinitionDestroy,
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"contentful": providerserver.NewProtocol6WithError(provider.New("test", true)),
		},
		Steps: []resource.TestStep{
			{
				Config: testAppDefinition("acctest_app_definition"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "tf_test1"),
					resource.TestMatchResourceAttr(resourceName, "id", regexp.MustCompile(`^[a-zA-Z0-9-_.]{1,64}$`)),
					resource.TestMatchResourceAttr(resourceName, "bundle_id", regexp.MustCompile(`^[a-zA-Z0-9-_.]{1,64}$`)),
					testAccCheckContentfulAppDefinitionExists(t, resourceName, func(t *testing.T, definition *sdk.AppDefinition) {
						assert.Nil(t, definition.Src)
						assert.EqualValues(t, definition.Name, "tf_test1")
						assert.Len(t, definition.Locations, 1)
						assert.EqualValues(t, definition.Locations[0].Location, "entry-field")
						assert.Len(t, definition.Locations[0].FieldTypes, 1)
						assert.EqualValues(t, definition.Locations[0].FieldTypes[0].Type, "Symbol")
						assert.NotNil(t, definition.Bundle)
						assert.Regexp(t, regexp.MustCompile(`^[a-zA-Z0-9-_.]{1,64}$`), definition.Bundle.Sys.Id)
					}),
				),
			},
			{
				Config: testAppDefinitionUpdateLocation("acctest_app_definition"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "tf_test1"),
					resource.TestMatchResourceAttr(resourceName, "id", regexp.MustCompile(`^[a-zA-Z0-9-_.]{1,64}$`)),
					resource.TestMatchResourceAttr(resourceName, "bundle_id", regexp.MustCompile(`^[a-zA-Z0-9-_.]{1,64}$`)),
					testAccCheckContentfulAppDefinitionExists(t, resourceName, func(t *testing.T, definition *sdk.AppDefinition) {
						assert.Nil(t, definition.Src)
						assert.EqualValues(t, definition.Name, "tf_test1")
						assert.Len(t, definition.Locations, 2)
						assert.EqualValues(t, definition.Locations[0].Location, "entry-field")
						assert.Len(t, definition.Locations[0].FieldTypes, 1)
						assert.EqualValues(t, definition.Locations[1].Location, "dialog")
						assert.EqualValues(t, definition.Locations[0].FieldTypes[0].Type, "Symbol")
						assert.NotNil(t, definition.Bundle)
						assert.Regexp(t, regexp.MustCompile(`^[a-zA-Z0-9-_.]{1,64}$`), definition.Bundle.Sys.Id)
					}),
				),
			},
			{
				Config: testAppDefinitionUpdateToSrc("acctest_app_definition"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "tf_test1"),
					resource.TestCheckResourceAttr(resourceName, "src", "http://localhost"),
					resource.TestCheckNoResourceAttr(resourceName, "bundle_id"),
					testAccCheckContentfulAppDefinitionExists(t, resourceName, func(t *testing.T, definition *sdk.AppDefinition) {
						assert.Equal(t, *definition.Src, "http://localhost")
						assert.EqualValues(t, definition.Name, "tf_test1")
						assert.Len(t, definition.Locations, 1)
						assert.EqualValues(t, definition.Locations[0].Location, "entry-field")
						assert.Len(t, definition.Locations[0].FieldTypes, 1)
						assert.EqualValues(t, definition.Locations[0].FieldTypes[0].Type, "Symbol")
						assert.Nil(t, definition.Bundle)
					}),
				),
			},
		},
	})
}

func testAccCheckContentfulAppDefinitionExists(t *testing.T, resourceName string, assertFunc assertFunc) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		definition, err := getAppDefinitionFromState(s, resourceName)
		if err != nil {
			return err
		}

		assertFunc(t, definition)
		return nil
	}
}

func getAppDefinitionFromState(s *terraform.State, resourceName string) (*sdk.AppDefinition, error) {
	rs, ok := s.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("app definition not found")
	}

	if rs.Primary.ID == "" {
		return nil, fmt.Errorf("no app definition ID found")
	}

	return getRemoteResource(rs.Primary.ID, 200)
}

func getRemoteResource(id string, expectStatusCode int) (*sdk.AppDefinition, error) {
	client := acctest.GetClient()

	resp, err := client.GetAppDefinitionWithResponse(context.Background(), os.Getenv("CONTENTFUL_ORGANIZATION_ID"), id)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != expectStatusCode {
		return nil, fmt.Errorf("status code %d, expected %d", resp.StatusCode(), expectStatusCode)
	}

	if resp.StatusCode() == 200 {
		return resp.JSON200, nil
	}
	return nil, nil
}

func testAccCheckContentfulAppDefinitionDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "contentful_app_definition" {
			continue
		}

		_, err := getRemoteResource(rs.Primary.ID, 404)
		if err != nil {
			return err
		}
	}

	return nil
}

func testAppDefinition(identifier string) string {
	return utils.HCLTemplateFromPath("test_resources/create.tf", map[string]any{
		"identifier": identifier,
	})
}

func testAppDefinitionUpdateLocation(identifier string) string {
	return utils.HCLTemplateFromPath("test_resources/update.tf", map[string]any{
		"identifier": identifier,
	})
}

func testAppDefinitionUpdateToSrc(identifier string) string {
	return utils.HCLTemplateFromPath("test_resources/update_to_src.tf", map[string]any{
		"identifier": identifier,
	})
}

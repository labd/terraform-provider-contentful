package app_installation_test

import (
	"context"
	"errors"
	"fmt"
	"github.com/flaconi/contentful-go/pkgs/common"
	"github.com/flaconi/contentful-go/pkgs/model"
	"github.com/flaconi/terraform-provider-contentful/internal/acctest"
	"github.com/flaconi/terraform-provider-contentful/internal/provider"
	"github.com/flaconi/terraform-provider-contentful/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

type assertFunc func(*testing.T, *model.AppInstallation)

func TestAppInstallation_Create(t *testing.T) {
	resourceName := "contentful_app_installation.acctest_app_installation"
	// merge app
	appInstallationId := "cQeaauOu1yUCYVhQ00atE"
	//graphql playground
	otherId := "66frtrAqmWSowDJzQNDiD"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.TestAccPreCheck(t) },
		CheckDestroy: testAccCheckContentfulAppInstallationDestroy,
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"contentful": providerserver.NewProtocol6WithError(provider.New("test", true)),
		},
		Steps: []resource.TestStep{
			{
				Config: testAppInstallation("acctest_app_installation", os.Getenv("CONTENTFUL_SPACE_ID"), "master", appInstallationId),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "app_definition_id", appInstallationId),
					testAccCheckContentfulAppInstallationExists(t, resourceName, func(t *testing.T, appInstallation *model.AppInstallation) {
						assert.IsType(t, map[string]any{}, appInstallation.Parameters)
						assert.Len(t, appInstallation.Parameters, 0)
						assert.EqualValues(t, appInstallationId, appInstallation.Sys.AppDefinition.Sys.ID)
					}),
				),
			},
			{
				Config: testAppInstallationWithParameter("acctest_app_installation_2", os.Getenv("CONTENTFUL_SPACE_ID"), "master", otherId),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("contentful_app_installation.acctest_app_installation_2", "app_definition_id", otherId),
					testAccCheckContentfulAppInstallationExists(t, "contentful_app_installation.acctest_app_installation_2", func(t *testing.T, appInstallation *model.AppInstallation) {
						assert.IsType(t, map[string]any{}, appInstallation.Parameters)
						assert.Len(t, appInstallation.Parameters, 1)
						assert.EqualValues(t, "not-working-ever", appInstallation.Parameters["cpaToken"])
						assert.EqualValues(t, otherId, appInstallation.Sys.AppDefinition.Sys.ID)
					}),
				),
			},
		},
	})
}

func getAppInstallationFromState(s *terraform.State, resourceName string) (*model.AppInstallation, error) {
	rs, ok := s.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("app installation not found")
	}

	if rs.Primary.ID == "" {
		return nil, fmt.Errorf("no app installation ID found")
	}

	client := acctest.GetCMA()

	return client.WithSpaceId(os.Getenv("CONTENTFUL_SPACE_ID")).WithEnvironment("master").AppInstallations().Get(context.Background(), rs.Primary.ID)
}

func testAccCheckContentfulAppInstallationExists(t *testing.T, resourceName string, assertFunc assertFunc) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		result, err := getAppInstallationFromState(s, resourceName)
		if err != nil {
			return err
		}

		assertFunc(t, result)
		return nil
	}
}

func testAccCheckContentfulAppInstallationDestroy(s *terraform.State) error {
	client := acctest.GetCMA()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "contentful_app_installation" {
			continue
		}

		_, err := client.WithSpaceId(os.Getenv("CONTENTFUL_SPACE_ID")).WithEnvironment("master").AppInstallations().Get(context.Background(), rs.Primary.ID)
		var notFoundError common.NotFoundError
		if errors.As(err, &notFoundError) {
			return nil
		}

		return fmt.Errorf("app installation still exists with id: %s", rs.Primary.ID)
	}

	return nil
}

func testAppInstallation(identifier string, spaceId string, environment string, appDefinitionId string) string {
	return utils.HCLTemplateFromPath("test_resources/without_terms.tf", map[string]any{
		"identifier":      identifier,
		"spaceId":         spaceId,
		"environment":     environment,
		"appDefinitionId": appDefinitionId,
	})
}

func testAppInstallationWithParameter(identifier string, spaceId string, environment string, appDefinitionId string) string {
	return utils.HCLTemplateFromPath("test_resources/with_terms.tf", map[string]any{
		"identifier":      identifier,
		"spaceId":         spaceId,
		"environment":     environment,
		"appDefinitionId": appDefinitionId,
	})
}

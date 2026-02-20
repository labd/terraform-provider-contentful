package preview_environment_test

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/labd/terraform-provider-contentful/internal/acctest"
	"github.com/labd/terraform-provider-contentful/internal/provider"
	"github.com/labd/terraform-provider-contentful/internal/sdk"
	"github.com/labd/terraform-provider-contentful/internal/utils"
)

func TestAccPreviewEnvironmentResource_Basic(t *testing.T) {
	spaceID := os.Getenv("CONTENTFUL_SPACE_ID")
	resourceName := "contentful_preview_environment.preview_environment"

	var previewEnvironment sdk.PreviewEnvironment

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.TestAccPreCheck(t)
			acctest.TestHasNoContentTypes(t)
		},
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"contentful": providerserver.NewProtocol6WithError(provider.New("test", false)()),
		},
		CheckDestroy: testAccCheckContentfulPreviewEnvironmentDestroy,
		Steps: []resource.TestStep{
			// Create and test initial setup
			{
				Config: testAccPreviewEnvironmentConfig(spaceID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPreviewEnvironmentExists(resourceName, &previewEnvironment),
					resource.TestCheckResourceAttr(resourceName, "space_id", spaceID),
					resource.TestCheckNoResourceAttr(resourceName, "description"),
				),
			},
			// Test updates
			{
				Config: testAccEditorInterfaceConfigUpdate(spaceID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPreviewEnvironmentExists(resourceName, &previewEnvironment),
					resource.TestCheckResourceAttr(resourceName, "description", "This is a test preview environment"),
				),
			},
			// Test import
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources[resourceName]
					if !ok {
						return "", fmt.Errorf("not found: %s", resourceName)
					}
					return fmt.Sprintf(
						"%s:%s",
						rs.Primary.ID,
						rs.Primary.Attributes["space_id"],
					), nil
				},
			},
		},
	})
}

func testAccCheckPreviewEnvironmentExists(n string, previewEnvironment *sdk.PreviewEnvironment) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		spaceID := rs.Primary.Attributes["space_id"]

		client := acctest.GetClient()
		resp, err := client.GetPreviewEnvironmentWithResponse(
			context.Background(),
			spaceID,
			rs.Primary.ID,
		)

		if err != nil {
			return err
		}

		if resp.StatusCode() != 200 {
			return fmt.Errorf("error getting editor interface: %d", resp.StatusCode())
		}

		*previewEnvironment = *resp.JSON200
		return nil
	}
}

func testAccCheckContentfulPreviewEnvironmentDestroy(s *terraform.State) error {
	client := acctest.GetClient()
	ctx := context.Background()

	for _, rs := range s.RootModule().Resources {
		if rs.Type == "contentful_contenttype" {
			spaceID := rs.Primary.Attributes["space_id"]
			environment := rs.Primary.Attributes["environment"]

			resp, err := client.GetContentTypeWithResponse(ctx, spaceID, environment, rs.Primary.ID)
			if err := utils.CheckClientResponse(resp, err, http.StatusNotFound); err != nil {
				return fmt.Errorf("error checking content type exists: %w", err)
			}
		}

		if rs.Type == "contentful_preview_environment" {
			spaceID := rs.Primary.Attributes["space_id"]

			resp, err := client.GetPreviewEnvironmentWithResponse(ctx, spaceID, rs.Primary.ID)
			if err := utils.CheckClientResponse(resp, err, http.StatusNotFound); err != nil {
				return fmt.Errorf("error checking preview environment exists: %w", err)
			}
		}
	}

	return nil
}

func testAccPreviewEnvironmentConfig(spaceID string) string {
	return fmt.Sprintf(`
resource "contentful_contenttype" "example_contenttype" {
  space_id      = "%s"
  environment   = "master-2026-02-20"
  id            = "example_contenttype"
  name          = "name"
  description   = "content type description"
  display_field = "name"

  fields = [
    {
      id       = "name"
      name     = "Name"
      type     = "Text"
      required = true
    }
  ]
}
	
resource "contentful_preview_environment" "preview_environment" {
  name     = "Test"
  space_id = "%s"
  configuration = [
	{
	  content_type = contentful_contenttype.example_contenttype.id
      enabled      = true
      url          = "http://www.example.com"
	}
  ]	
}
`, spaceID, spaceID)
}

func testAccEditorInterfaceConfigUpdate(spaceID string) string {
	return fmt.Sprintf(`
resource "contentful_contenttype" "example_contenttype" {
  space_id      = "%s"
  environment   = "master-2026-02-20"
  id            = "example_contenttype"
  name          = "name"
  description   = "content type description"
  display_field = "name"

  fields = [
    {
      id       = "name"
      name     = "Name"
      type     = "Text"
      required = true
    }
  ]
}	
	
resource "contentful_preview_environment" "preview_environment" {
  name     = "Test"
  space_id = "%s"
  description = "This is a test preview environment"
  configuration = [
	{
	  content_type = contentful_contenttype.example_contenttype.id
      enabled      = true
      url          = "http://www.example.com"
	}
  ]	
}
`, spaceID, spaceID)
}

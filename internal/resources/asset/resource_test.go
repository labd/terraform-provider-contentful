package asset_test

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

type assertFunc func(*testing.T, *sdk.Asset)

func TestAssetResource_Basic(t *testing.T) {
	assetName := fmt.Sprintf("asset-%s", hashicor_acctest.RandString(3))
	resourceName := "contentful_asset.myasset"
	spaceID := os.Getenv("CONTENTFUL_SPACE_ID")
	environment := "master"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.TestAccPreCheck(t) },
		CheckDestroy: testAccCheckContentfulAssetDestroy,
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"contentful": providerserver.NewProtocol6WithError(provider.New("test", true)()),
		},
		Steps: []resource.TestStep{
			{
				Config: testAssetConfig(spaceID, environment, assetName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "fields.title.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "fields.title.0.content", "Asset title"),
					resource.TestCheckResourceAttr(resourceName, "fields.description.0.content", "Asset description"),
					resource.TestCheckResourceAttr(resourceName, "published", "true"),
					resource.TestCheckResourceAttr(resourceName, "archived", "false"),
					testAccCheckContentfulAssetExists(t, resourceName, func(t *testing.T, asset *sdk.Asset) {
						assert.NotNil(t, asset.Sys.PublishedAt, "Asset should be published")
						assert.Equal(t, "Asset title", asset.Fields.Title["en-US"])
						assert.Equal(t, "Asset description", asset.Fields.Description["en-US"])
					}),
				),
			},
			{
				Config: testAssetUpdateConfig(spaceID, environment, assetName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "fields.title.0.content", "Updated asset title"),
					resource.TestCheckResourceAttr(resourceName, "fields.description.0.content", "Updated asset description"),
					resource.TestCheckResourceAttr(resourceName, "published", "false"),
					testAccCheckContentfulAssetExists(t, resourceName, func(t *testing.T, asset *sdk.Asset) {
						assert.Nil(t, asset.Sys.PublishedAt, "Asset should not be published")
						assert.Equal(t, "Updated asset title", asset.Fields.Title["en-US"])
						assert.Equal(t, "Updated asset description", asset.Fields.Description["en-US"])
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: false,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources[resourceName]
					if !ok {
						return "", fmt.Errorf("not found: %s", resourceName)
					}
					return fmt.Sprintf("%s:%s:%s",
						rs.Primary.ID,
						rs.Primary.Attributes["space_id"],
						rs.Primary.Attributes["environment"]), nil
				},
			},
		},
	})
}

func testAccCheckContentfulAssetExists(t *testing.T, resourceName string, assertFunc assertFunc) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		asset, err := getAssetFromState(s, resourceName)
		if err != nil {
			return err
		}

		assertFunc(t, asset)
		return nil
	}
}

func getAssetFromState(s *terraform.State, resourceName string) (*sdk.Asset, error) {
	rs, ok := s.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("asset not found in state: %s", resourceName)
	}

	if rs.Primary.ID == "" {
		return nil, fmt.Errorf("no asset ID found")
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
	resp, err := client.GetAssetWithResponse(context.Background(), spaceID, environment, rs.Primary.ID)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("asset not found: %s", rs.Primary.ID)
	}

	return resp.JSON200, nil
}

func testAccCheckContentfulAssetDestroy(s *terraform.State) error {
	client := acctest.GetClient()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "contentful_asset" {
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

		resp, err := client.GetAssetWithResponse(context.Background(), spaceID, environment, rs.Primary.ID)
		if err != nil {
			return err
		}

		if resp.StatusCode() == 404 {
			return nil
		}

		return fmt.Errorf("asset still exists with id: %s", rs.Primary.ID)
	}

	return nil
}

func testAssetConfig(spaceID, environment, name string) string {
	return fmt.Sprintf(`
resource "contentful_asset" "myasset" {
  asset_id = "%s"
  environment = "%s"
  space_id = "%s"
  fields {
    title {
      locale = "en-US"
      content = "Asset title"
    }
    description {
      locale = "en-US"
      content = "Asset description"
    }
    file {
      upload = "https://images.ctfassets.net/fo9twyrwpveg/2VQx7vz73aMEYi20MMgCk0/66e502115b1f1f973a944b4bd2cc536f/IC-1H_Modern_Stack_Website.svg"
      file_name = "example.jpeg"
      content_type = "image/jpeg"
      locale = "en-US"
    }
  }
  published = true
  archived = false
}
`, name, environment, spaceID)
}

func testAssetUpdateConfig(spaceID, environment, name string) string {
	return fmt.Sprintf(`
resource "contentful_asset" "myasset" {
  asset_id = "%s"
  environment = "%s"
  space_id = "%s"
  fields {
    title {
      locale = "en-US"
      content = "Updated asset title"
    }
    description {
      locale = "en-US"
      content = "Updated asset description"
    }
    file {
      upload = "https://images.ctfassets.net/fo9twyrwpveg/2VQx7vz73aMEYi20MMgCk0/66e502115b1f1f973a944b4bd2cc536f/IC-1H_Modern_Stack_Website.svg"
      file_name = "example.jpeg"
      content_type = "image/jpeg"
      locale = "en-US"
    }
  }
  published = false
  archived = false
}
`, name, environment, spaceID)
}

package contentful

import (
	"context"
	"fmt"
	"github.com/flaconi/contentful-go/pkgs/model"
	"github.com/flaconi/terraform-provider-contentful/internal/acctest"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccContentfulAsset_Basic(t *testing.T) {
	var asset model.Asset

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccContentfulAssetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccContentfulAssetConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContentfulAssetExists("contentful_asset.myasset", &asset),
					testAccCheckContentfulAssetAttributes(&asset, map[string]interface{}{
						"space_id": spaceID,
					}),
				),
			},
			{
				Config: testAccContentfulAssetUpdateConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContentfulAssetExists("contentful_asset.myasset", &asset),
					testAccCheckContentfulAssetAttributes(&asset, map[string]interface{}{
						"space_id": spaceID,
					}),
				),
			},
		},
	})
}

func testAccCheckContentfulAssetExists(n string, asset *model.Asset) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not Found: %s", n)
		}

		spaceID := rs.Primary.Attributes["space_id"]
		if spaceID == "" {
			return fmt.Errorf("no space_id is set")
		}

		client := acctest.GetCMA()

		contentfulAsset, err := client.WithSpaceId(spaceID).WithEnvironment("master").Assets().Get(context.Background(), rs.Primary.ID)
		if err != nil {
			return err
		}

		*asset = *contentfulAsset

		return nil
	}
}

func testAccCheckContentfulAssetAttributes(asset *model.Asset, attrs map[string]interface{}) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		spaceIDCheck := attrs["space_id"].(string)
		if asset.Sys.Space.Sys.ID != spaceIDCheck {
			return fmt.Errorf("space id  does not match: %s, %s", asset.Sys.Space.Sys.ID, spaceIDCheck)
		}

		return nil
	}
}

func testAccContentfulAssetDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "contentful_entry" {
			continue
		}

		// get space id from resource data
		spaceID := rs.Primary.Attributes["space_id"]
		if spaceID == "" {
			return fmt.Errorf("no space_id is set")
		}

		// check webhook resource id
		if rs.Primary.ID == "" {
			return fmt.Errorf("no asset ID is set")
		}

		// sdk client
		client := acctest.GetCMA()

		asset, _ := client.WithSpaceId(spaceID).WithEnvironment("master").Assets().Get(context.Background(), rs.Primary.ID)
		if asset == nil {
			return nil
		}

		return fmt.Errorf("asset still exists with id: %s", rs.Primary.ID)
	}

	return nil
}

var testAccContentfulAssetConfig = `
resource "contentful_asset" "myasset" {
  asset_id = "test_asset"
  
  environment = "master"
  space_id = "` + spaceID + `"
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
`

var testAccContentfulAssetUpdateConfig = `
resource "contentful_asset" "myasset" {
  asset_id = "test_asset"
  environment = "master"
  space_id = "` + spaceID + `"
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
`

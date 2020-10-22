package contentful

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	contentful "github.com/labd/contentful-go"
)

func TestAccContentfulUpload_Basic(t *testing.T) {
	var upload contentful.Resource

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccContentfulUploadDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccContentfulUploadConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContentfulUploadExists("contentful_asset.myupload", &upload),
					testAccCheckContentfulUploadAttributes(&upload, map[string]interface{}{
						"space_id": spaceID,
					}),
				),
			},
		},
	})
}

func testAccCheckContentfulUploadExists(n string, upload *contentful.Resource) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not Found: %s", n)
		}

		spaceID := rs.Primary.Attributes["space_id"]
		if spaceID == "" {
			return fmt.Errorf("no space_id is set")
		}

		client := testAccProvider.Meta().(*contentful.Client)

		contentfulAsset, err := client.Resources.Get(spaceID, rs.Primary.ID)
		if err != nil {
			return err
		}

		*upload = *contentfulAsset

		return nil
	}
}

func testAccCheckContentfulUploadAttributes(upload *contentful.Resource, attrs map[string]interface{}) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		spaceIDCheck := attrs["space_id"].(string)
		if upload.Sys.Space.Sys.ID != spaceIDCheck {
			return fmt.Errorf("space id  does not match: %s, %s", upload.Sys.Space.Sys.ID, spaceIDCheck)
		}

		return nil
	}
}

func testAccContentfulUploadDestroy(s *terraform.State) error {
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
		client := testAccProvider.Meta().(*contentful.Client)

		asset, _ := client.Assets.Get(spaceID, rs.Primary.ID)
		if asset == nil {
			return nil
		}

		return fmt.Errorf("asset still exists with id: %s", rs.Primary.ID)
	}

	return nil
}

var testAccContentfulUploadConfig = `
resource "contentful_upload" "myupload" {
  space_id = "` + spaceID + `"
  file_path =  "/home/kantoor/go/src/github.com/labd/terraform-provider-contentful/local/upload_test.png"
  asset_id = "upload_test"
  locale = "en-US"
  title = "This is an asset"
  description = "Uploaded asset!"
}
`

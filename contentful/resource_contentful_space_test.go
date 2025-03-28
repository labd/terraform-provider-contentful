package contentful

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/labd/terraform-provider-contentful/internal/sdk"
)

func TestAccContentfulSpace_Basic(t *testing.T) {
	t.Skip() // Space resource can only be tested when user has the rights to do so, if not, skip this test!
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckContentfulSpaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccContentfulSpaceConfig,
				Check: resource.TestCheckResourceAttr(
					"contentful_space.myspace", "name", "TF Acc Test Space"),
			},
			{
				Config: testAccContentfulSpaceUpdateConfig,
				Check: resource.TestCheckResourceAttr(
					"contentful_space.myspace", "name", "TF Acc Test Changed Space"),
			},
		},
	})
}

func testAccCheckContentfulSpaceDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*sdk.ClientWithResponses)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "contentful_space" {
			continue
		}

		resp, err := client.GetSpaceWithResponse(context.Background(), rs.Primary.ID)
		if err == nil {
			space := resp.JSON200
			return fmt.Errorf("space %s still exists after destroy", space.Sys.Id)
		}
	}

	return nil
}

var testAccContentfulSpaceConfig = `
resource "contentful_space" "myspace" {
  name = "Playground"
}
`

var testAccContentfulSpaceUpdateConfig = `
resource "contentful_space" "myspace" {
  name = "TF Acc Test Changed Space"
}
`

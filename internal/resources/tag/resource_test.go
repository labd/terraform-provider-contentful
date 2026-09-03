package tag_test

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	hashicorp_acctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/assert"

	"github.com/labd/terraform-provider-contentful/internal/acctest"
	"github.com/labd/terraform-provider-contentful/internal/provider"
	"github.com/labd/terraform-provider-contentful/internal/sdk"
)

type assertFunc func(*testing.T, *sdk.Tag)

func TestTagResource_Basic(t *testing.T) {
	tagID := fmt.Sprintf("tag-%s", hashicorp_acctest.RandString(8))
	name := fmt.Sprintf("Tag %s", hashicorp_acctest.RandString(8))
	resourceName := "contentful_tag.test"
	spaceID := os.Getenv("CONTENTFUL_SPACE_ID")
	environment := os.Getenv("CONTENTFUL_ENVIRONMENT")
	if environment == "" {
		environment = "master"
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.TestAccPreCheck(t) },
		CheckDestroy: testAccCheckContentfulTagDestroy,
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"contentful": providerserver.NewProtocol6WithError(provider.New("test", true)()),
		},
		Steps: []resource.TestStep{
			{
				Config: testTagConfig(spaceID, environment, tagID, name, "public"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "id", tagID),
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "visibility", "public"),
					testAccCheckContentfulTagExists(t, resourceName, func(t *testing.T, tag *sdk.Tag) {
						assert.Equal(t, name, tag.Name)
						assert.Equal(t, sdk.TagVisibilityPublic, tag.Sys.Visibility)
					}),
				),
			},
			{
				Config: testTagConfig(spaceID, environment, tagID, name+" updated", "public"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", name+" updated"),
					testAccCheckContentfulTagExists(t, resourceName, func(t *testing.T, tag *sdk.Tag) {
						assert.Equal(t, name+" updated", tag.Name)
					}),
				),
			},
			{
				ResourceName:       resourceName,
				ImportState:        true,
				ImportStateVerify:  true,
				ImportStatePersist: true,
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					rs, ok := state.RootModule().Resources[resourceName]
					if !ok {
						return "", fmt.Errorf("not found: %s", resourceName)
					}
					return fmt.Sprintf("%s:%s:%s", rs.Primary.ID, rs.Primary.Attributes["environment"], rs.Primary.Attributes["space_id"]), nil
				},
			},
			{
				Config:   testTagConfigWithoutVisibility(spaceID, environment, tagID, name+" updated"),
				PlanOnly: true,
			},
		},
	})
}

func testAccCheckContentfulTagExists(t *testing.T, resourceName string, check assertFunc) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		tag, err := getTagFromState(state, resourceName)
		if err != nil {
			return err
		}

		check(t, tag)
		return nil
	}
}

func getTagFromState(state *terraform.State, resourceName string) (*sdk.Tag, error) {
	rs, ok := state.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("tag not found in state: %s", resourceName)
	}

	client := acctest.GetClient()
	resp, err := client.GetTagWithResponse(
		context.Background(),
		rs.Primary.Attributes["space_id"],
		rs.Primary.Attributes["environment"],
		rs.Primary.ID,
	)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("tag not found: %s", rs.Primary.ID)
	}

	return resp.JSON200, nil
}

func testAccCheckContentfulTagDestroy(state *terraform.State) error {
	client := acctest.GetClient()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "contentful_tag" {
			continue
		}

		resp, err := client.GetTagWithResponse(
			context.Background(),
			rs.Primary.Attributes["space_id"],
			rs.Primary.Attributes["environment"],
			rs.Primary.ID,
		)
		if err != nil {
			return err
		}
		if resp.StatusCode() == http.StatusNotFound {
			return nil
		}

		return fmt.Errorf("tag still exists with id: %s", rs.Primary.ID)
	}

	return nil
}

func testTagConfig(spaceID, environment, tagID, name, visibility string) string {
	return fmt.Sprintf(`
resource "contentful_tag" "test" {
  space_id    = %q
  environment = %q
  id          = %q
  name        = %q
  visibility  = %q
}
`, spaceID, environment, tagID, name, visibility)
}

func testTagConfigWithoutVisibility(spaceID, environment, tagID, name string) string {
	return fmt.Sprintf(`
resource "contentful_tag" "test" {
  space_id    = %q
  environment = %q
  id          = %q
  name        = %q
}
`, spaceID, environment, tagID, name)
}

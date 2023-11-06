package contenttype_test

import (
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/labd/contentful-go"
	"github.com/labd/terraform-provider-contentful/internal/acctest"
	"github.com/labd/terraform-provider-contentful/internal/provider"
	"github.com/labd/terraform-provider-contentful/internal/utils"
	"os"

	"testing"
)

func TestContentTypeResource_Create(t *testing.T) {
	resourceName := "contentful_contenttype.acctest_content_type"
	linkedResourceName := "contentful_contenttype.linked_content_type"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.TestAccPreCheck(t) },
		CheckDestroy: testAccCheckContentfulContentTypeDestroy,
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"contentful": providerserver.NewProtocol6WithError(provider.New("test", false)),
		},
		Steps: []resource.TestStep{
			{
				Config: testContentType("acctest_content_type", os.Getenv("CONTENTFUL_SPACE_ID")),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "tf_test1"),
					resource.TestCheckResourceAttr(resourceName, "id", "tf_test1"),
					resource.TestCheckResourceAttr(resourceName, "version", "2"),
					resource.TestCheckResourceAttr(resourceName, "version_controls", "0"),
					//todo check in contentful directly if the type looks like this
				),
			},
			{
				Config: testContentTypeUpdate("acctest_content_type", os.Getenv("CONTENTFUL_SPACE_ID")),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "tf_test1"),
					resource.TestCheckResourceAttr(resourceName, "version", "2"),
					resource.TestCheckResourceAttr(resourceName, "version_controls", "2"),
					//todo check in contentful directly if the type looks like this
				),
			},
			{
				Config: testContentTypeLinkConfig("acctest_content_type", os.Getenv("CONTENTFUL_SPACE_ID"), "linked_content_type"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(linkedResourceName, "name", "tf_linked"),
					//todo check in contentful directly if the type looks like this
				),
			},
			{
				Config: testContentTypeWithId("acctest_content_type", os.Getenv("CONTENTFUL_SPACE_ID")),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "tf_test1"),
					resource.TestCheckResourceAttr(resourceName, "id", "tf_test2"),
					//todo check in contentful directly if the type looks like this
				),
			},
		},
	})
}

func testAccCheckContentfulContentTypeDestroy(s *terraform.State) (err error) {
	client := acctest.GetClient()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "contentful_contenttype" {
			continue
		}

		spaceID := rs.Primary.Attributes["space_id"]
		if spaceID == "" {
			return fmt.Errorf("no space_id is set")
		}

		_, err := client.ContentTypes.Get(spaceID, rs.Primary.ID)
		var notFoundError contentful.NotFoundError
		if errors.As(err, &notFoundError) {
			return nil
		}

		return fmt.Errorf("content type still exists with id: %s", rs.Primary.ID)
	}

	return nil
}

func testContentType(identifier string, spaceId string) string {
	return utils.HCLTemplate(`
		resource "contentful_contenttype" "{{ .identifier }}" {
  space_id = "{{ .spaceId }}"
  name = "tf_test1"
  description = "Terraform Acc Test Content Type description change"
  display_field = "field1"
  fields = [{
    id        = "field1"
    name      = "Field 1 name change"
    required  = true
    type      = "Text"
  }, {
    id        = "field3"
    name      = "Field 3 new field"
    required  = true
    type      = "Integer"
  }]
}`, map[string]any{
		"identifier": identifier,
		"spaceId":    spaceId,
	})
}

func testContentTypeWithId(identifier string, spaceId string) string {
	return utils.HCLTemplate(`
		resource "contentful_contenttype" "{{ .identifier }}" {
  space_id = "{{ .spaceId }}"
  id = "tf_test2"
  name = "tf_test1"
  description = "Terraform Acc Test Content Type description change"
  display_field = "field1"
  fields = [{
    id        = "field1"
    name      = "Field 1 name change"
    required  = true
    type      = "Text"
  }, {
    id        = "field3"
    name      = "Field 3 new field"
    required  = true
    type      = "Integer"
  }]
}`, map[string]any{
		"identifier": identifier,
		"spaceId":    spaceId,
	})
}

func testContentTypeUpdate(identifier string, spaceId string) string {
	return utils.HCLTemplate(`
		resource "contentful_contenttype" "{{ .identifier }}" {
  space_id = "{{ .spaceId }}"
  name = "tf_test1"
  description = "Terraform Acc Test Content Type description change"
  display_field = "field1"
manage_field_controls = true
  fields = [{
    id        = "field1"
    name      = "Field 1 name change"
    required  = true
    type      = "Text"
  }, {
    id        = "field3"
    name      = "Field 3 new field"
    required  = true
    type      = "Integer"
    control = {
		widget_id = "numberEditor"
    	widget_namespace = "builtin"
	}
  }]
}`, map[string]any{
		"identifier": identifier,
		"spaceId":    spaceId,
	})
}

func testContentTypeLinkConfig(identifier string, spaceId string, linkIdentifier string) string {
	return utils.HCLTemplate(`resource "contentful_contenttype" "{{ .identifier }}" {
  space_id = "{{ .spaceId }}"
  name = "tf_test1"
  description = "Terraform Acc Test Content Type description change"
  display_field = "field1"
manage_field_controls = true
  fields = [{
    id        = "field1"
    name      = "Field 1 name change"
    required  = true
    type      = "Text"
  }, {
    id        = "field3"
    name      = "Field 3 new field"
    required  = true
    type      = "Integer"
    control = {
		widget_id = "numberEditor"
    	widget_namespace = "builtin"
	}
  }]
}

resource "contentful_contenttype" "{{ .linkIdentifier }}" {
   space_id = "{{ .spaceId }}"
  name          = "tf_linked"
  description   = "Terraform Acc Test Content Type with links"
  display_field = "asset_field"
  fields =[{
    id   = "asset_field"
    name = "Asset Field"
    type = "Array"
    items = {
      type      = "Link"
      link_type = "Asset"
    }
    required = true
  },{
    id        = "entry_link_field"
    name      = "Entry Link Field"
    type      = "Link"
    link_type = "Entry"
    validations = [{
		link_content_type = [contentful_contenttype.{{ .identifier }}.id ]
    }]
  }]
}
`, map[string]any{
		"identifier":     identifier,
		"linkIdentifier": linkIdentifier,
		"spaceId":        spaceId,
	})
}

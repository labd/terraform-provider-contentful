package contentful

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	contentful "github.com/labd/contentful-go"
)

func TestAccContentfulContentType_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckContentfulContentTypeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccContentfulContentTypeConfig,
				Check: resource.TestCheckResourceAttr(
					"contentful_contenttype.basic", "name", "tf_test1"),
			},
			{
				Config: testAccContentfulContentTypeUpdateConfig,
				Check: resource.TestCheckResourceAttr(
					"contentful_contenttype.basic", "name", "tf_test1"),
			},
			{
				Config: testAccContentfulContentTypeLinkConfig,
				Check: resource.TestCheckResourceAttr(
					"contentful_contenttype.basicmylinked_contenttype", "name", "tf_linked"),
			},
		},
	})
}

var testAccContentfulContentTypeConfig = `
resource "contentful_contenttype" "basic" {
  space_id = "` + spaceID + `"
  name = "tf_test1"
  description = "Terraform Acc Test Content Type"
  display_field = "field1"
  field {
	disabled  = false
	id        = "field1"
	localized = false
	name      = "Field 1"
	omitted   = false
	required  = true
	type      = "Text"
  }
  field {
	disabled  = false
	id        = "field2"
	localized = false
	name      = "Field 2"
	omitted   = false
	required  = true
	type      = "Integer"
  }
}
`

var testAccContentfulContentTypeUpdateConfig = `
resource "contentful_contenttype" "basic" {
  space_id = "` + spaceID + `"
  name = "tf_test1"
  description = "Terraform Acc Test Content Type description change"
  display_field = "field1"
  field {
    disabled  = false
    id        = "field1"
    localized = false
    name      = "Field 1 name change"
    omitted   = false
    required  = true
    type      = "Text"
  }
  field {
    disabled  = false
    id        = "field3"
    localized = false
    name      = "Field 3 new field"
    omitted   = false
    required  = true
    type      = "Integer"
  }
}
`

var testAccContentfulContentTypeLinkConfig = `
resource "contentful_contenttype" "basic" {
  space_id = "` + spaceID + `"
  name = "tf_test1"
  description = "Terraform Acc Test Content Type description change"
  display_field = "field1"
  field {
    disabled  = false
    id        = "field1"
    localized = false
    name      = "Field 1 name change"
    omitted   = false
    required  = true
    type      = "Text"
  }
  field {
    disabled  = false
    id        = "field3"
    localized = false
    name      = "Field 3 new field"
    omitted   = false
    required  = true
    type      = "Integer"
  }	
}

resource "contentful_contenttype" "basicmylinked_contenttype" {
  space_id = "` + spaceID + `"
  name          = "tf_linked"
  description   = "Terraform Acc Test Content Type with links"
  display_field = "asset_field"
  field {
    id   = "asset_field"
    name = "Asset Field"
    type = "Array"
    items {
      type      = "Link"
      link_type = "Asset"
    }
    required = true
  }
  field {
    id        = "entry_link_field"
    name      = "Entry Link Field"
    type      = "Link"
    link_type = "Entry"
    validations = [
	  jsonencode({
		linkContentType = [
			contentful_contenttype.basic.id
		]
	  })
	]
    required = false
  }
}
`

func TestAccContentfulContentType_WithEnv(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckContentfulContentTypeDestroy,
		// CheckDestroy: func(s *terraform.State) (err error) {
		// if err = testAccCheckContentfulContentTypeDestroy(s); err != nil {
		// 	return
		// } else if err = testAccContentfulEnvironmentDestroy(s); err != nil {
		// 	return
		// }
		// return nil
		// },
		Steps: []resource.TestStep{
			{

				Config: testAccContentfulContentTypeConfigWithEnv,
				Check: resource.TestCheckResourceAttr(
					"contentful_contenttype.env-contenttype", "name", "tf_test2"),
			},
			{
				Config: testAccContentfulContentTypeUpdateConfigWithEnv,
				Check: resource.TestCheckResourceAttr(
					"contentful_contenttype.env-contenttype", "name", "tf_test2"),
			},
			{
				Config: testAccContentfulContentTypeLinkConfigWithEnv,
				Check: resource.TestCheckResourceAttr(
					"contentful_contenttype.env-linked_contenttype", "name", "tf_linked"),
			},
		},
	})
}

// noinspection GoUnusedFunction
func testAccCheckContentfulContentTypeExists(n string, contentType *contentful.ContentType) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no content type ID is set")
		}

		spaceID := rs.Primary.Attributes["space_id"]
		if spaceID == "" {
			return fmt.Errorf("no space_id is set")
		}

		client := testAccProvider.Meta().(*contentful.Client)

		ct, err := client.ContentTypes.Get(spaceID, rs.Primary.ID)
		if err != nil {
			return err
		}

		*contentType = *ct

		return nil
	}
}

func testAccCheckContentfulContentTypeDestroy(s *terraform.State) (err error) {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "contentful_contenttype" {
			continue
		}

		spaceID := rs.Primary.Attributes["space_id"]
		if spaceID == "" {
			return fmt.Errorf("no space_id is set")
		}

		client := testAccProvider.Meta().(*contentful.Client)

		_, err := client.ContentTypes.Get(spaceID, rs.Primary.ID)
		if _, ok := err.(contentful.NotFoundError); ok {
			return nil
		}

		return fmt.Errorf("content type still exists with id: %s", rs.Primary.ID)
	}

	return nil
}

var testAccContentfulContentTypeConfigWithEnv = `
resource "contentful_environment" "testenvironment" {
  space_id = "` + spaceID + `"
  name = "provider-test"
}

resource "contentful_contenttype" "env-contenttype" {
  space_id = "` + spaceID + `"
  environment_id = contentful_environment.testenvironment.id
  name = "tf_test2"
  description = "Terraform Acc Test Content Type in Environment"
  display_field = "field1"
  field {
	disabled  = false
	id        = "field1"
	localized = false
	name      = "Field 1"
	omitted   = false
	required  = true
	type      = "Text"
  }
  field {
	disabled  = false
	id        = "field2"
	localized = false
	name      = "Field 2"
	omitted   = false
	required  = true
	type      = "Integer"
  }
}
`

var testAccContentfulContentTypeUpdateConfigWithEnv = `
resource "contentful_environment" "testenvironment" {
  space_id = "` + spaceID + `"
  name = "provider-test"
}

resource "contentful_contenttype" "env-contenttype" {
  space_id = "` + spaceID + `"
  environment_id = contentful_environment.testenvironment.id
  name = "tf_test2"
  description = "Terraform Acc Test Content Type in Environment description change"
  display_field = "field1"
  field {
    disabled  = false
    id        = "field1"
    localized = false
    name      = "Field 1 name change"
    omitted   = false
    required  = true
    type      = "Text"
  }
  field {
    disabled  = false
    id        = "field3"
    localized = false
    name      = "Field 3 new field"
    omitted   = false
    required  = true
    type      = "Integer"
  }	
}
`

var testAccContentfulContentTypeLinkConfigWithEnv = `
resource "contentful_environment" "testenvironment" {
  space_id = "` + spaceID + `"
  name = "provider-test"
}

resource "contentful_contenttype" "env-contenttype" {
  space_id = "` + spaceID + `"
  environment_id = contentful_environment.testenvironment.id
  name = "tf_test2"
  description = "Terraform Acc Test Content Type in Environment description change"
  display_field = "field1"
  field {
    disabled  = false
    id        = "field1"
    localized = false
    name      = "Field 1 name change"
    omitted   = false
    required  = true
    type      = "Text"
  }
  field {
    disabled  = false
    id        = "field3"
    localized = false
    name      = "Field 3 new field"
    omitted   = false
    required  = true
    type      = "Integer"
  }	
}

resource "contentful_contenttype" "env-linked_contenttype" {
  space_id = "` + spaceID + `"
  environment_id = contentful_environment.testenvironment.id
  name          = "tf_linked"
  description   = "Terraform Acc Test Content Type with links"
  display_field = "asset_field"
  field {
    id   = "asset_field"
    name = "Asset Field"
    type = "Array"
    items {
      type      = "Link"
      link_type = "Asset"
    }
    required = true
  }
  field {
    id        = "entry_link_field"
    name      = "Entry Link Field"
    type      = "Link"
    link_type = "Entry"
    validations = [
	  jsonencode({
		linkContentType = [
			contentful_contenttype.env-contenttype.id
		]
	  })
	]
    required = false
  }
}
`

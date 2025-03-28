package contentful

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"

	"github.com/labd/terraform-provider-contentful/internal/acctest"
	"github.com/labd/terraform-provider-contentful/internal/sdk"
)

func TestParseContentValue_String(t *testing.T) {
	value := "hello"
	assert.Equal(t, parseContentValue(value), value)
}

func TestParseContentValue_Json(t *testing.T) {
	value := `{"foo": "bar", "baz": [1, 2, 3]}`
	assert.Equal(t, parseContentValue(value), map[string]interface{}{"foo": "bar", "baz": []interface{}{float64(1), float64(2), float64(3)}})
}

func TestAccContentfulEntry_Basic(t *testing.T) {
	//todo remove skip when entry is moved to new sdk style as content type already moved
	t.Skip()
	var entry sdk.Entry

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccContentfulEntryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccContentfulEntryConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContentfulEntryExists("contentful_entry.myentry", &entry),
					testAccCheckContentfulEntryAttributes(&entry, map[string]interface{}{
						"space_id": spaceID,
					}),
				),
			},
			{
				Config: testAccContentfulEntryUpdateConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContentfulEntryExists("contentful_entry.myentry", &entry),
					testAccCheckContentfulEntryAttributes(&entry, map[string]interface{}{
						"space_id": spaceID,
					}),
				),
			},
		},
	})
}

func testAccCheckContentfulEntryExists(n string, entry *sdk.Entry) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not Found: %s", n)
		}

		spaceID := rs.Primary.Attributes["space_id"]
		if spaceID == "" {
			return fmt.Errorf("no space_id is set")
		}

		contenttypeID := rs.Primary.Attributes["contenttype_id"]
		if contenttypeID == "" {
			return fmt.Errorf("no contenttype_id is set")
		}

		client := acctest.GetClient()
		resp, err := client.GetEntryWithResponse(context.Background(), spaceID, "master", rs.Primary.ID)
		if err != nil {
			return err
		}

		if resp.StatusCode() != 200 {
			return fmt.Errorf("entry not found: %s", rs.Primary.ID)
		}

		*entry = *resp.JSON200

		return nil
	}
}

func testAccCheckContentfulEntryAttributes(entry *sdk.Entry, attrs map[string]interface{}) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		spaceIDCheck := attrs["space_id"].(string)
		if entry.Sys.Space.Sys.Id != spaceIDCheck {
			return fmt.Errorf("space id  does not match: %s, %s", entry.Sys.Space.Sys.Id, spaceIDCheck)
		}

		return nil
	}
}

func testAccContentfulEntryDestroy(s *terraform.State) error {
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
			return fmt.Errorf("no entry ID is set")
		}

		client := acctest.GetClient()
		resp, err := client.GetEntryWithResponse(context.Background(), spaceID, "master", rs.Primary.ID)
		if err != nil {
			return err
		}

		if resp.StatusCode() == 404 {
			return nil
		}

		return fmt.Errorf("entry still exists with id: %s", rs.Primary.ID)
	}

	return nil
}

var testAccContentfulEntryConfig = `
resource "contentful_contenttype" "mycontenttype" {
  space_id = "` + spaceID + `"
  name = "tf_test_1"
  environment = "master"
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
	type      = "Text"
  }
  field {
    id       = "field3"
    name     = "Field 3"
    type     = "RichText"
  }
}

resource "contentful_entry" "myentry" {
  entry_id = "mytestentry"
  space_id = "` + spaceID + `"
  environment = "master"
  contenttype_id = "tf_test_1"
  locale = "en-US"
  field {
    id = "field1"
    content = "Hello, World!"
    locale = "en-US"
  }
  field {
    id = "field2"
    content = "Bacon is healthy!"
    locale = "en-US"
  }

  field {
    id = "field3"
    locale = "en-US"
    content = jsonencode({
      data= {},
      content= [
        {
          nodeType= "paragraph",
          content= [
            {
              nodeType= "text",
              marks= [],
              value= "This is another paragraph.",
              data= {},
            },
          ],
          data= {},
        }
      ],
      nodeType= "document"
    })
  }
  published = true
  archived  = false
  depends_on = [contentful_contenttype.mycontenttype]
}
`

var testAccContentfulEntryUpdateConfig = `
resource "contentful_contenttype" "mycontenttype" {
  space_id = "` + spaceID + `"
  environment = "master"
  name = "tf_test_1"
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
	type      = "Text"
  }
  field {
    id       = "field3"
    name     = "Field 3"
    type     = "RichText"
  }
}

resource "contentful_entry" "myentry" {
  entry_id = "mytestentry"
  space_id = "` + spaceID + `"
  environment = "master"
  contenttype_id = "tf_test_1"
  locale = "en-US"
  field {
    id = "field1"
    content = "Hello, World!"
    locale = "en-US"
  }
  field {
    id = "field2"
    content = "Bacon is healthy!"
    locale = "en-US"
  }
  field {
    id = "field3"
    locale = "en-US"
    content = jsonencode({
      data= {},
      content= [
        {
          nodeType= "paragraph",
          content= [
            {
              nodeType= "text",
              marks= [],
              value= "This is another paragraph.",
              data= {},
            },
          ],
          data= {},
        }
      ],
      nodeType= "document"
    })
  }
  published = false
  archived  = false
  depends_on = [contentful_contenttype.mycontenttype]
}
`

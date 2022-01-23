package contentful

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	contentful "github.com/labd/contentful-go"
)

func TestAccContentfulEntry_Basic(t *testing.T) {
	var entry contentful.Entry

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

func testAccCheckContentfulEntryExists(n string, entry *contentful.Entry) resource.TestCheckFunc {
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

		client := testAccProvider.Meta().(*contentful.Client)

		contentfulEntry, err := client.Entries.Get(spaceID, rs.Primary.ID)
		if err != nil {
			return err
		}

		*entry = *contentfulEntry

		return nil
	}
}

func testAccCheckContentfulEntryAttributes(entry *contentful.Entry, attrs map[string]interface{}) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		spaceIDCheck := attrs["space_id"].(string)
		if entry.Sys.Space.Sys.ID != spaceIDCheck {
			return fmt.Errorf("space id  does not match: %s, %s", entry.Sys.Space.Sys.ID, spaceIDCheck)
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

		// sdk client
		client := testAccProvider.Meta().(*contentful.Client)

		entry, _ := client.Entries.Get(spaceID, rs.Primary.ID)
		if entry == nil {
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
}

resource "contentful_entry" "myentry" {
  entry_id = "mytestentry"
  space_id = "` + spaceID + `"
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
  published = true
  archived  = false
  depends_on = [contentful_contenttype.mycontenttype]
}
`

var testAccContentfulEntryUpdateConfig = `
resource "contentful_contenttype" "mycontenttype" {
  space_id = "` + spaceID + `"
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
}

resource "contentful_entry" "myentry" {
  entry_id = "mytestentry"
  space_id = "` + spaceID + `"
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
  published = false
  archived  = false
  depends_on = [contentful_contenttype.mycontenttype]
}
`

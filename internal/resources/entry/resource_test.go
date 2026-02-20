package entry_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/assert"

	"github.com/labd/terraform-provider-contentful/internal/acctest"
	"github.com/labd/terraform-provider-contentful/internal/provider"
	"github.com/labd/terraform-provider-contentful/internal/resources/entry"
	"github.com/labd/terraform-provider-contentful/internal/sdk"
)

func TestParseContentValue_String(t *testing.T) {
	value := `"hello"`
	parsed := entry.ParseContentValue(value)
	assert.Equal(t, "hello", parsed)
}

func TestParseContentValue_Json(t *testing.T) {
	value := `{"foo": "bar", "baz": [1, 2, 3]}`
	parsed := entry.ParseContentValue(value)
	assert.Equal(t, map[string]interface{}{"foo": "bar", "baz": []interface{}{float64(1), float64(2), float64(3)}}, parsed)
}

func TestEntryResource_Basic(t *testing.T) {
	spaceID := os.Getenv("CONTENTFUL_SPACE_ID")
	resourceName := "contentful_entry.myentry"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.TestAccPreCheck(t) },
		CheckDestroy: testAccCheckContentfulEntryDestroy,
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"contentful": providerserver.NewProtocol6WithError(provider.New("test", true)()),
		},
		Steps: []resource.TestStep{
			{
				Config: testEntryConfig(spaceID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContentfulEntryExists(t, resourceName, func(t *testing.T, entry *sdk.Entry) {
						assert.Equal(t, "mytestentry", entry.Sys.Id)
						assert.Equal(t, "tf_test_1", entry.Sys.ContentType.Sys.Id)
						assert.Equal(t, spaceID, entry.Sys.Space.Sys.Id)
						assert.NotNil(t, entry.Sys.PublishedAt)
					}),
				),
			},
			{
				Config: testEntryUpdateConfig(spaceID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContentfulEntryExists(t, resourceName, func(t *testing.T, entry *sdk.Entry) {
						assert.Equal(t, "mytestentry", entry.Sys.Id)
						assert.Equal(t, "tf_test_1", entry.Sys.ContentType.Sys.Id)
						assert.Equal(t, spaceID, entry.Sys.Space.Sys.Id)
						assert.Nil(t, entry.Sys.PublishedAt)
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
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

type assertFunc func(*testing.T, *sdk.Entry)

func testAccCheckContentfulEntryExists(t *testing.T, resourceName string, assertFunc assertFunc) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		entry, err := getEntryFromState(s, resourceName)
		if err != nil {
			return err
		}

		assertFunc(t, entry)
		return nil
	}
}

func getEntryFromState(s *terraform.State, resourceName string) (*sdk.Entry, error) {
	rs, ok := s.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("entry not found in state: %s", resourceName)
	}

	if rs.Primary.ID == "" {
		return nil, fmt.Errorf("no entry ID found")
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
	resp, err := client.GetEntryWithResponse(context.Background(), spaceID, environment, rs.Primary.ID)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("entry not found: %s", rs.Primary.ID)
	}

	return resp.JSON200, nil
}

func testAccCheckContentfulEntryDestroy(s *terraform.State) error {
	client := acctest.GetClient()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "contentful_entry" {
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

		resp, err := client.GetEntryWithResponse(context.Background(), spaceID, environment, rs.Primary.ID)
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

func testEntryConfig(spaceID string) string {
	return fmt.Sprintf(`
resource "contentful_contenttype" "mycontenttype" {
  space_id = "%s"
  name = "tf_test_1"
  environment = "master-2026-02-20"
  description = "Terraform Acc Test Content Type"
  display_field = "field1"

  fields = [
		{
			disabled  = false
			id        = "field1"
			localized = false
			name      = "Field 1"
			omitted   = false
			required  = true
			type      = "Text"
		},
		{
			disabled  = false
			id        = "field2"
			localized = false
			name      = "Field 2"
			omitted   = false
			required  = true
			type      = "Text"
		},
		{
			id       = "field3"
			name     = "Field 3"
			type     = "RichText"
		}
	]
}

resource "contentful_entry" "myentry" {
  entry_id = "mytestentry"
  space_id = "%s"
  environment = "master-2026-02-20"
  contenttype_id = "tf_test_1"
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
`, spaceID, spaceID)
}

func testEntryUpdateConfig(spaceID string) string {
	return fmt.Sprintf(`
resource "contentful_contenttype" "mycontenttype" {
  space_id = "%s"
  environment = "master-2026-02-20"
  name = "tf_test_1"
  description = "Terraform Acc Test Content Type"
  display_field = "field1"

	fields = [
		{
			disabled  = false
			id        = "field1"
			localized = false
			name      = "Field 1"
			omitted   = false
			required  = true
			type      = "Text"
		},
		{
			disabled  = false
			id        = "field2"
			localized = false
			name      = "Field 2"
			omitted   = false
			required  = true
			type      = "Text"
		},
		{
			id       = "field3"
			name     = "Field 3"
			type     = "RichText"
		}
	]
}

resource "contentful_entry" "myentry" {
  entry_id = "mytestentry"
  space_id = "%s"
  environment = "master-2026-02-20"
  contenttype_id = "tf_test_1"
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
`, spaceID, spaceID)
}

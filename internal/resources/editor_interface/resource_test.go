package editor_interface_test

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/labd/terraform-provider-contentful/internal/acctest"
	"github.com/labd/terraform-provider-contentful/internal/provider"
	"github.com/labd/terraform-provider-contentful/internal/sdk"
	"github.com/labd/terraform-provider-contentful/internal/utils"
)

func TestAccEditorInterfaceResource_Basic(t *testing.T) {
	spaceID := os.Getenv("CONTENTFUL_SPACE_ID")
	resourceName := "contentful_editor_interface.test_editor_interface"
	resourceNameDatePicker := "contentful_editor_interface.test_editor_interface_datepicker"

	var editorInterface sdk.EditorInterface

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.TestAccPreCheck(t)
			acctest.TestHasNoContentTypes(t)
		},
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"contentful": providerserver.NewProtocol6WithError(provider.New("test", false)()),
		},
		CheckDestroy: testAccCheckContentfulEditorInterfaceDestroy,
		Steps: []resource.TestStep{
			// Create and test initial setup
			{
				Config: testAccEditorInterfaceConfig(spaceID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContentfulEditorInterfaceExists(resourceName, &editorInterface),
					resource.TestCheckResourceAttr(resourceName, "space_id", spaceID),
					resource.TestCheckResourceAttr(resourceName, "environment", "master-2026-02-20"),
					resource.TestCheckResourceAttrSet(resourceName, "content_type"),
					resource.TestCheckResourceAttr(resourceName, "controls.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "controls.0.field_id", "title"),
					resource.TestCheckResourceAttr(resourceName, "controls.0.widget_id", "singleLine"),
					resource.TestCheckResourceAttr(resourceName, "controls.0.widget_namespace", "builtin"),
					resource.TestCheckResourceAttr(resourceName, "controls.1.field_id", "description"),
					resource.TestCheckResourceAttr(resourceName, "controls.1.widget_id", "markdown"),
					resource.TestCheckResourceAttr(resourceName, "controls.1.widget_namespace", "builtin"),
					resource.TestCheckResourceAttr(resourceName, "controls.1.settings.help_text", "Markdown content"),
				),
			},
			// Test updates
			{
				Config: testAccEditorInterfaceConfigUpdate(spaceID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContentfulEditorInterfaceExists(resourceName, &editorInterface),
					resource.TestCheckResourceAttr(resourceName, "controls.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "controls.0.field_id", "title"),
					resource.TestCheckResourceAttr(resourceName, "controls.0.widget_id", "singleLine"),
					resource.TestCheckResourceAttr(resourceName, "controls.1.field_id", "description"),
					resource.TestCheckResourceAttr(resourceName, "controls.1.widget_id", "richTextEditor"),
					resource.TestCheckResourceAttr(resourceName, "controls.1.settings.help_text", "Rich text content"),
					resource.TestCheckResourceAttr(resourceName, "sidebar.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "sidebar.0.widget_id", "content-preview-widget"),
					resource.TestCheckResourceAttr(resourceName, "sidebar.1.widget_id", "translation-widget"),
					resource.TestCheckResourceAttr(resourceName, "editors.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "editors.0.widget_id", "default-editor"),
				),
			},
			{
				Config: testAccEditorInterfaceConfigUpdateWithEmptySideBarDefaultEditorEnabled(spaceID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContentfulEditorInterfaceExists(resourceName, &editorInterface),
					resource.TestCheckResourceAttr(resourceName, "sidebar.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "editors.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "editors.0.disabled", "false"),
				),
			},
			// Test import
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources[resourceName]
					if !ok {
						return "", fmt.Errorf("not found: %s", resourceName)
					}
					return fmt.Sprintf(
						"%s:%s:%s",
						rs.Primary.Attributes["space_id"],
						rs.Primary.Attributes["environment"],
						rs.Primary.Attributes["content_type"],
					), nil
				},
			},

			// Test DatePicker widget
			{
				Config: testAccEditorInterfaceConfig_DatePicker(spaceID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContentfulEditorInterfaceExists(resourceNameDatePicker, &editorInterface),
					resource.TestCheckResourceAttr(resourceNameDatePicker, "space_id", spaceID),
					resource.TestCheckResourceAttr(resourceNameDatePicker, "environment", "master-2026-02-20"),
					resource.TestCheckResourceAttrSet(resourceNameDatePicker, "content_type"),
					resource.TestCheckResourceAttr(resourceNameDatePicker, "controls.#", "1"),
					resource.TestCheckResourceAttr(resourceNameDatePicker, "controls.0.field_id", "date"),
					resource.TestCheckResourceAttr(resourceNameDatePicker, "controls.0.widget_id", "datePicker"),
					resource.TestCheckResourceAttr(resourceNameDatePicker, "controls.0.widget_namespace", "builtin"),
					resource.TestCheckResourceAttr(resourceNameDatePicker, "controls.0.settings.ampm", "24"),
					resource.TestCheckResourceAttr(resourceNameDatePicker, "controls.0.settings.format", "dateonly"),
				),
			},
		},
	})
}

func testAccCheckContentfulEditorInterfaceExists(n string, editorInterface *sdk.EditorInterface) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		spaceID := rs.Primary.Attributes["space_id"]
		environment := rs.Primary.Attributes["environment"]
		contentType := rs.Primary.Attributes["content_type"]

		client := acctest.GetClient()
		resp, err := client.GetEditorInterfaceWithResponse(
			context.Background(),
			spaceID,
			environment,
			contentType,
		)

		if err != nil {
			return err
		}

		if resp.StatusCode() != 200 {
			return fmt.Errorf("error getting editor interface: %d", resp.StatusCode())
		}

		*editorInterface = *resp.JSON200
		return nil
	}
}

func testAccCheckContentfulEditorInterfaceDestroy(s *terraform.State) error {
	client := acctest.GetClient()
	ctx := context.Background()

	for _, rs := range s.RootModule().Resources {
		if rs.Type == "contentful_contenttype" {
			spaceID := rs.Primary.Attributes["space_id"]
			environment := rs.Primary.Attributes["environment"]

			resp, err := client.GetContentTypeWithResponse(ctx, spaceID, environment, rs.Primary.ID)
			if err := utils.CheckClientResponse(resp, err, http.StatusNotFound); err != nil {
				return fmt.Errorf("error checking content type exists: %w", err)
			}
		}

		// Should be removed, since we removed the content type
		if rs.Type == "contentful_editor_interface" {
			spaceID := rs.Primary.Attributes["space_id"]
			environment := rs.Primary.Attributes["environment"]
			contentType := rs.Primary.Attributes["content_type"]

			resp, err := client.GetEditorInterfaceWithResponse(ctx, spaceID, environment, contentType)
			if err := utils.CheckClientResponse(resp, err, http.StatusNotFound); err != nil {
				return fmt.Errorf("error checking editor interface exists: %w", err)
			}
		}
	}

	return nil
}

func testAccEditorInterfaceConfig(spaceID string) string {
	return fmt.Sprintf(`
resource "contentful_contenttype" "test_contenttype" {
  space_id     = "%s"
  environment  = "master-2026-02-20"
  id         	 = "test-content-type"
  name         = "test content type"
  description  = "Test Content Type for Editor Interface"
  display_field = "title"

	fields  = [
		{
			id        = "title"
			name      = "Title"
			type      = "Symbol"
			required  = true
		},
		{
			id        = "description"
			name      = "Description"
			type      = "Text"
			required  = false
		}
	]
}

resource "contentful_editor_interface" "test_editor_interface" {
  space_id     = contentful_contenttype.test_contenttype.space_id
  environment  = contentful_contenttype.test_contenttype.environment
  content_type = contentful_contenttype.test_contenttype.id

  controls = [
    {
      field_id = "title"
      widget_id = "singleLine"
      widget_namespace = "builtin"
    },
    {
      field_id = "description"
      widget_id = "markdown"
      widget_namespace = "builtin"
      settings = {
        help_text = "Markdown content"
      }
    }
  ]
}
`, spaceID)
}

func testAccEditorInterfaceConfigUpdate(spaceID string) string {
	return fmt.Sprintf(`
resource "contentful_contenttype" "test_contenttype" {
  space_id     = "%s"
  environment  = "master-2026-02-20"
  name         = "test-content-type"
  description  = "Test Content Type for Editor Interface"
  display_field = "title"

	fields = [
		{
			id        = "title"
			name      = "Title"
			type      = "Symbol"
			required  = true
		},
		{
			id        = "description"
			name      = "Description"
			type      = "Text"
			required  = false
		}
	]
}

resource "contentful_editor_interface" "test_editor_interface" {
  space_id     = contentful_contenttype.test_contenttype.space_id
  environment  = contentful_contenttype.test_contenttype.environment
  content_type = contentful_contenttype.test_contenttype.id

  controls = [
    {
      field_id = "title"
      widget_id = "singleLine"
      widget_namespace = "builtin"
    },
    {
      field_id = "description"
      widget_id = "richTextEditor"
      widget_namespace = "builtin"
      settings = {
        help_text = "Rich text content"
      }
    }
  ]
	
	sidebar = [
    {
      widget_id        = "content-preview-widget"
      widget_namespace = "sidebar-builtin"
    },
    {
      widget_id        = "translation-widget"
      widget_namespace = "sidebar-builtin"
    }
  ]
	
	  editors = [
    {
      widget_namespace = "editor-builtin",
      widget_id = "default-editor",
      disabled = true
    }
  ]
}
`, spaceID)
}

func testAccEditorInterfaceConfigUpdateWithEmptySideBarDefaultEditorEnabled(spaceID string) string {
	return fmt.Sprintf(`
resource "contentful_contenttype" "test_contenttype" {
  space_id     = "%s"
  environment  = "master-2026-02-20"
  name         = "test-content-type"
  description  = "Test Content Type for Editor Interface"
  display_field = "title"

	fields = [
		{
			id        = "title"
			name      = "Title"
			type      = "Symbol"
			required  = true
		},
		{
			id        = "description"
			name      = "Description"
			type      = "Text"
			required  = false
		}
	]
}

resource "contentful_editor_interface" "test_editor_interface" {
  space_id     = contentful_contenttype.test_contenttype.space_id
  environment  = contentful_contenttype.test_contenttype.environment
  content_type = contentful_contenttype.test_contenttype.id

  controls = [
    {
      field_id = "title"
      widget_id = "singleLine"
      widget_namespace = "builtin"
    },
    {
      field_id = "description"
      widget_id = "richTextEditor"
      widget_namespace = "builtin"
      settings = {
        help_text = "Rich text content"
      }
    }
  ]
	
	sidebar = []
	
	  editors = [
    {
      widget_namespace = "editor-builtin",
      widget_id = "default-editor"
    }
  ]
}
`, spaceID)
}

func testAccEditorInterfaceConfig_DatePicker(spaceID string) string {
	return fmt.Sprintf(`
resource "contentful_contenttype" "test_contenttype_datepicker" {
  space_id     = "%s"
  environment  = "master-2026-02-20"
  id         	 = "test-content-type-datepicker"
  name         = "test content type date picker"
  description  = "Test Content Type for Editor Interface with Date Picker"
  display_field = "date"

	fields  = [
		{
			id        = "date"
			name      = "Date"
			type      = "Date"
		}
	]
}

resource "contentful_editor_interface" "test_editor_interface_datepicker" {
  space_id     = contentful_contenttype.test_contenttype_datepicker.space_id
  environment  = contentful_contenttype.test_contenttype_datepicker.environment
  content_type = contentful_contenttype.test_contenttype_datepicker.id

  controls = [
    {
      field_id         = "date"
      widget_id        = "datePicker"
      widget_namespace = "builtin",
      settings         = { ampm : "24", format : "dateonly" }
    }
  ]
}
`, spaceID)
}

package contenttype_test

import (
	"context"
	"errors"
	"fmt"
	"github.com/flaconi/contentful-go"
	"github.com/flaconi/contentful-go/pkgs/common"
	"github.com/flaconi/contentful-go/pkgs/model"
	"github.com/flaconi/terraform-provider-contentful/internal/acctest"
	"github.com/flaconi/terraform-provider-contentful/internal/provider"
	"github.com/flaconi/terraform-provider-contentful/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
	"os"
	"regexp"

	"testing"
)

type assertFunc func(*testing.T, *model.ContentType)
type assertEditorInterfaceFunc func(*testing.T, *contentful.EditorInterface)

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
					testAccCheckContentfulContentTypeExists(t, resourceName, func(t *testing.T, contentType *model.ContentType) {
						assert.EqualValues(t, "tf_test1", contentType.Name)
						assert.Equal(t, 2, contentType.Sys.Version)
						assert.EqualValues(t, "tf_test1", contentType.Sys.ID)
						assert.EqualValues(t, "none", *contentType.Description)
						assert.EqualValues(t, "field1", contentType.DisplayField)
						assert.Len(t, contentType.Fields, 2)
						assert.Equal(t, &model.Field{
							ID:           "field1",
							Name:         "Field 1 name change",
							Type:         "Text",
							LinkType:     "",
							Items:        nil,
							Required:     true,
							Localized:    false,
							Disabled:     false,
							Omitted:      false,
							Validations:  nil,
							DefaultValue: nil,
						}, contentType.Fields[0])
						assert.Equal(t, &model.Field{
							ID:           "field3",
							Name:         "Field 3 new field",
							Type:         "Integer",
							LinkType:     "",
							Items:        nil,
							Required:     true,
							Localized:    false,
							Disabled:     false,
							Omitted:      false,
							Validations:  nil,
							DefaultValue: nil,
						}, contentType.Fields[1])
					}),
				),
			},
			{
				Config: testContentTypeUpdateWithDifferentOrderOfFields("acctest_content_type", os.Getenv("CONTENTFUL_SPACE_ID")),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "tf_test1"),
					resource.TestCheckResourceAttr(resourceName, "version", "4"),
					resource.TestCheckResourceAttr(resourceName, "version_controls", "0"),
					testAccCheckContentfulContentTypeExists(t, resourceName, func(t *testing.T, contentType *model.ContentType) {
						assert.EqualValues(t, "tf_test1", contentType.Name)
						assert.Equal(t, 4, contentType.Sys.Version)
						assert.EqualValues(t, "tf_test1", contentType.Sys.ID)
						assert.EqualValues(t, "Terraform Acc Test Content Type description change", *contentType.Description)
						assert.EqualValues(t, "field1", contentType.DisplayField)
						assert.Len(t, contentType.Fields, 2)
						assert.Equal(t, &model.Field{
							ID:        "field1",
							Name:      "Field 1 name change",
							Type:      "Text",
							LinkType:  "",
							Required:  true,
							Localized: false,
							Disabled:  false,
							Omitted:   false,
						}, contentType.Fields[1])
						assert.Equal(t, &model.Field{
							ID:        "field3",
							Name:      "Field 3 new field",
							Type:      "Integer",
							LinkType:  "",
							Required:  true,
							Localized: false,
							Disabled:  false,
							Omitted:   false,
						}, contentType.Fields[0])
					}),
				),
			},
			{
				Config: testContentTypeUpdate("acctest_content_type", os.Getenv("CONTENTFUL_SPACE_ID")),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "tf_test1"),
					resource.TestCheckResourceAttr(resourceName, "version", "6"),
					resource.TestCheckResourceAttr(resourceName, "version_controls", "4"),
					testAccCheckContentfulContentTypeExists(t, resourceName, func(t *testing.T, contentType *model.ContentType) {
						assert.EqualValues(t, "tf_test1", contentType.Name)
						assert.Equal(t, 6, contentType.Sys.Version)
						assert.EqualValues(t, "tf_test1", contentType.Sys.ID)
						assert.EqualValues(t, "Terraform Acc Test Content Type description change", *contentType.Description)
						assert.EqualValues(t, "field1", contentType.DisplayField)
						assert.Len(t, contentType.Fields, 2)
						assert.Equal(t, &model.Field{
							ID:        "field1",
							Name:      "Field 1 name change",
							Type:      "Text",
							LinkType:  "",
							Required:  true,
							Localized: false,
							Disabled:  false,
							Omitted:   false,
						}, contentType.Fields[0])
						assert.Equal(t, &model.Field{
							ID:        "field3",
							Name:      "Field 3 new field",
							Type:      "Integer",
							LinkType:  "",
							Required:  true,
							Localized: false,
							Disabled:  false,
							Omitted:   false,
						}, contentType.Fields[1])
					}),
					testAccCheckEditorInterfaceExists(t, "tf_test1", func(t *testing.T, editorInterface *contentful.EditorInterface) {
						assert.Len(t, editorInterface.Controls, 2)
						assert.Equal(t, contentful.Controls{
							FieldID: "field1",
						}, editorInterface.Controls[0])
						assert.Equal(t, contentful.Controls{
							FieldID:         "field3",
							WidgetNameSpace: toPointer("builtin"),
							WidgetID:        toPointer("numberEditor"),
							Settings: &contentful.Settings{
								BulkEditing: toPointer(true),
								HelpText:    toPointer("blabla"),
							},
						}, editorInterface.Controls[1])
					}),
				),
			},
			{
				Config: testContentTypeLinkConfig("acctest_content_type", os.Getenv("CONTENTFUL_SPACE_ID"), "linked_content_type"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(linkedResourceName, "name", "tf_linked"),
					testAccCheckContentfulContentTypeExists(t, linkedResourceName, func(t *testing.T, contentType *model.ContentType) {
						assert.EqualValues(t, "tf_linked", contentType.Name)
						assert.Equal(t, 2, contentType.Sys.Version)
						assert.EqualValues(t, "tf_linked", contentType.Sys.ID)
						assert.EqualValues(t, "Terraform Acc Test Content Type with links", *contentType.Description)
						assert.EqualValues(t, "asset_field", contentType.DisplayField)
						assert.Len(t, contentType.Fields, 2)
						assert.Equal(t, &model.Field{
							ID:       "asset_field",
							Name:     "Asset Field",
							Type:     "Array",
							LinkType: "",
							Items: &model.FieldTypeArrayItem{
								Type:     "Link",
								LinkType: toPointer("Asset"),
							},
							Required:  true,
							Localized: false,
							Disabled:  false,
							Omitted:   false,
						}, contentType.Fields[0])
						assert.Equal(t, &model.Field{
							ID:        "entry_link_field",
							Name:      "Entry Link Field",
							Type:      "Link",
							LinkType:  "Entry",
							Required:  false,
							Localized: false,
							Disabled:  false,
							Omitted:   false,
							Validations: []model.FieldValidation{
								model.FieldValidationLink{
									LinkContentType: []string{"tf_test1"},
								},
							},
						}, contentType.Fields[1])
					}),
				),
			},
			{
				Config: testContentTypeWithId("acctest_content_type", os.Getenv("CONTENTFUL_SPACE_ID")),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "tf_test1"),
					resource.TestCheckResourceAttr(resourceName, "id", "tf_test2"),
					testAccCheckContentfulContentTypeExists(t, resourceName, func(t *testing.T, contentType *model.ContentType) {
						assert.EqualValues(t, "tf_test1", contentType.Name)
						assert.Equal(t, 2, contentType.Sys.Version)
						assert.EqualValues(t, "tf_test2", contentType.Sys.ID)
						assert.EqualValues(t, "Terraform Acc Test Content Type description change", *contentType.Description)
						assert.EqualValues(t, "field1", contentType.DisplayField)
						assert.Len(t, contentType.Fields, 2)
						assert.Equal(t, &model.Field{
							ID:        "field1",
							Name:      "Field 1 name change",
							Type:      "Text",
							LinkType:  "",
							Required:  true,
							Localized: false,
							Disabled:  false,
							Omitted:   false,
						}, contentType.Fields[0])
						assert.Equal(t, &model.Field{
							ID:        "field3",
							Name:      "Field 3 new field",
							Type:      "Integer",
							LinkType:  "",
							Required:  true,
							Localized: false,
							Disabled:  false,
							Omitted:   false,
						}, contentType.Fields[1])
					}),
				),
			},
		},
	})
}

func TestContentTypeResource_WithDuplicateField(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.TestAccPreCheck(t) },
		CheckDestroy: testAccCheckContentfulContentTypeDestroy,
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"contentful": providerserver.NewProtocol6WithError(provider.New("test", false)),
		},
		Steps: []resource.TestStep{
			{
				Config:      testContentTypeDuplicateFields("acctest_content_type", os.Getenv("CONTENTFUL_SPACE_ID")),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile("Error: Duplicate List Value"),
			},
		},
	})
}

func getContentTypeFromState(s *terraform.State, resourceName string) (*model.ContentType, error) {
	rs, ok := s.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("content type not found")
	}

	if rs.Primary.ID == "" {
		return nil, fmt.Errorf("no content type ID found")
	}

	client := acctest.GetCMA()

	return client.WithSpaceId(os.Getenv("CONTENTFUL_SPACE_ID")).WithEnvironment("master").ContentTypes().Get(context.Background(), rs.Primary.ID)
}

func getEditorInterfaceFromState(id string) (*contentful.EditorInterface, error) {
	client := acctest.GetClient()

	return client.EditorInterfaces.Get(os.Getenv("CONTENTFUL_SPACE_ID"), id)
}

func testAccCheckContentfulContentTypeExists(t *testing.T, resourceName string, assertFunc assertFunc) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		result, err := getContentTypeFromState(s, resourceName)
		if err != nil {
			return err
		}

		assertFunc(t, result)
		return nil
	}
}

func testAccCheckEditorInterfaceExists(t *testing.T, id string, assertFunc assertEditorInterfaceFunc) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		result, err := getEditorInterfaceFromState(id)
		if err != nil {
			return err
		}

		assertFunc(t, result)
		return nil
	}
}

func testAccCheckContentfulContentTypeDestroy(s *terraform.State) (err error) {
	client := acctest.GetCMA()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "contentful_contenttype" {
			continue
		}

		spaceID := rs.Primary.Attributes["space_id"]
		if spaceID == "" {
			return fmt.Errorf("no space_id is set")
		}

		_, err := client.WithSpaceId(spaceID).WithEnvironment("master").ContentTypes().Get(context.Background(), rs.Primary.ID)
		var notFoundError common.NotFoundError
		if errors.As(err, &notFoundError) {
			return nil
		}

		return fmt.Errorf("content type still exists with id: %s", rs.Primary.ID)
	}

	return nil
}

func testContentType(identifier string, spaceId string) string {
	return utils.HCLTemplateFromPath("test_resources/create.tf", map[string]any{
		"identifier":    identifier,
		"spaceId":       spaceId,
		"id_definition": "",
		"desc":          "none",
	})
}

func testContentTypeWithId(identifier string, spaceId string) string {
	return utils.HCLTemplateFromPath("test_resources/create.tf", map[string]any{
		"identifier":    identifier,
		"spaceId":       spaceId,
		"id_definition": "id = \"tf_test2\"",
		"desc":          "Terraform Acc Test Content Type description change",
	})
}

func testContentTypeUpdateWithDifferentOrderOfFields(identifier string, spaceId string) string {
	return utils.HCLTemplateFromPath("test_resources/changed_order.tf", map[string]any{
		"identifier": identifier,
		"spaceId":    spaceId,
	})
}

func testContentTypeUpdate(identifier string, spaceId string) string {
	return utils.HCLTemplateFromPath("test_resources/update.tf", map[string]any{
		"identifier": identifier,
		"spaceId":    spaceId,
	})
}

func testContentTypeDuplicateFields(identifier string, spaceId string) string {
	return utils.HCLTemplateFromPath("test_resources/update_duplicate_field.tf", map[string]any{
		"identifier": identifier,
		"spaceId":    spaceId,
	})
}

func testContentTypeLinkConfig(identifier string, spaceId string, linkIdentifier string) string {
	return utils.HCLTemplateFromPath("test_resources/link_config.tf", map[string]any{
		"identifier":     identifier,
		"linkIdentifier": linkIdentifier,
		"spaceId":        spaceId,
	})

}

func toPointer[T string | bool](value T) *T {
	return &value
}

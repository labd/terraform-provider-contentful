package api_key_test

import (
	"context"
	"errors"
	"fmt"
	"github.com/flaconi/contentful-go/pkgs/common"
	"github.com/flaconi/contentful-go/pkgs/model"
	"github.com/flaconi/terraform-provider-contentful/internal/acctest"
	"github.com/flaconi/terraform-provider-contentful/internal/provider"
	"github.com/flaconi/terraform-provider-contentful/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	hashicor_acctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/assert"
	"os"
	"regexp"
	"testing"
)

type assertFunc func(*testing.T, *model.APIKey)

func TestApiKeyResource_Create(t *testing.T) {
	name := fmt.Sprintf("apikey-name-%s", hashicor_acctest.RandString(3))
	description := fmt.Sprintf("apikey-description-%s", hashicor_acctest.RandString(3))
	resourceName := "contentful_apikey.myapikey"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.TestAccPreCheck(t) },
		CheckDestroy: testAccCheckContentfulApiKeyDestroy,
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"contentful": providerserver.NewProtocol6WithError(provider.New("test", true)),
		},
		Steps: []resource.TestStep{
			{
				Config: testApiKey(os.Getenv("CONTENTFUL_SPACE_ID"), name, description),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "description", description),
					resource.TestMatchResourceAttr(resourceName, "id", regexp.MustCompile(`^[a-zA-Z0-9-_.]{1,64}$`)),
					resource.TestMatchResourceAttr(resourceName, "preview_id", regexp.MustCompile(`^[a-zA-Z0-9-_.]{1,64}$`)),
					testAccCheckContentfulApiKeyExists(t, resourceName, func(t *testing.T, apiKey *model.APIKey) {
						assert.NotEmpty(t, apiKey.AccessToken)
						assert.EqualValues(t, name, apiKey.Name)
						assert.EqualValues(t, description, apiKey.Description)
						assert.Regexp(t, regexp.MustCompile(`^[a-zA-Z0-9-_.]{1,64}$`), apiKey.Sys.ID)
						assert.EqualValues(t, "master", apiKey.Environments[0].Sys.ID)
					}),
				),
			},
			{
				Config: testApiKeyUpdate(os.Getenv("CONTENTFUL_SPACE_ID"), name, description),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("%s-updated", name)),
					resource.TestCheckResourceAttr(resourceName, "description", fmt.Sprintf("%s-updated", description)),
					resource.TestMatchResourceAttr(resourceName, "id", regexp.MustCompile(`^[a-zA-Z0-9-_.]{1,64}$`)),
					resource.TestMatchResourceAttr(resourceName, "preview_id", regexp.MustCompile(`^[a-zA-Z0-9-_.]{1,64}$`)),
					testAccCheckContentfulApiKeyExists(t, resourceName, func(t *testing.T, apiKey *model.APIKey) {
						assert.NotEmpty(t, apiKey.AccessToken)
						assert.EqualValues(t, fmt.Sprintf("%s-updated", name), apiKey.Name)
						assert.EqualValues(t, fmt.Sprintf("%s-updated", description), apiKey.Description)
						assert.Regexp(t, regexp.MustCompile(`^[a-zA-Z0-9-_.]{1,64}$`), apiKey.Sys.ID)
					}),
				),
			},
		},
	})
}

func TestApiKeyResource_CreateWithEnvironmentSet(t *testing.T) {
	name := fmt.Sprintf("apikey-name-%s", hashicor_acctest.RandString(3))
	description := fmt.Sprintf("apikey-description-%s", hashicor_acctest.RandString(3))
	resourceName := "contentful_apikey.myapikey"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.TestAccPreCheck(t) },
		CheckDestroy: testAccCheckContentfulApiKeyDestroy,
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"contentful": providerserver.NewProtocol6WithError(provider.New("test", true)),
		},
		Steps: []resource.TestStep{
			{
				Config: testApiKeyWithEnvironment(os.Getenv("CONTENTFUL_SPACE_ID"), name, description),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "description", description),
					resource.TestMatchResourceAttr(resourceName, "id", regexp.MustCompile(`^[a-zA-Z0-9-_.]{1,64}$`)),
					resource.TestMatchResourceAttr(resourceName, "preview_id", regexp.MustCompile(`^[a-zA-Z0-9-_.]{1,64}$`)),
					testAccCheckContentfulApiKeyExists(t, resourceName, func(t *testing.T, apiKey *model.APIKey) {
						assert.NotEmpty(t, apiKey.AccessToken)
						assert.EqualValues(t, name, apiKey.Name)
						assert.EqualValues(t, description, apiKey.Description)
						assert.Regexp(t, regexp.MustCompile(`^[a-zA-Z0-9-_.]{1,64}$`), apiKey.Sys.ID)
						assert.EqualValues(t, "notexisting", apiKey.Environments[0].Sys.ID)
					}),
				),
			},
		},
	})
}

func testAccCheckContentfulApiKeyExists(t *testing.T, resourceName string, assertFunc assertFunc) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		definition, err := getApiKeyFromState(s, resourceName)
		if err != nil {
			return err
		}

		assertFunc(t, definition)
		return nil
	}
}

func getApiKeyFromState(s *terraform.State, resourceName string) (*model.APIKey, error) {
	rs, ok := s.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("api key not found")
	}

	if rs.Primary.ID == "" {
		return nil, fmt.Errorf("no api key ID found")
	}

	client := acctest.GetCMA()

	return client.WithSpaceId(os.Getenv("CONTENTFUL_SPACE_ID")).ApiKeys().Get(context.Background(), rs.Primary.ID)
}

func testAccCheckContentfulApiKeyDestroy(s *terraform.State) error {
	client := acctest.GetCMA()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "contentful_apikey" {
			continue
		}

		_, err := client.WithSpaceId(os.Getenv("CONTENTFUL_SPACE_ID")).ApiKeys().Get(context.Background(), rs.Primary.ID)
		var notFoundError common.NotFoundError
		if errors.As(err, &notFoundError) {
			return nil
		}

		return fmt.Errorf("api key still exists with id: %s", rs.Primary.ID)
	}

	return nil
}

func testApiKey(spaceId string, name string, description string) string {
	return utils.HCLTemplateFromPath("test_resources/create.tf", map[string]any{
		"spaceId":     spaceId,
		"name":        name,
		"description": description,
	})
}

func testApiKeyWithEnvironment(spaceId string, name string, description string) string {
	return utils.HCLTemplateFromPath("test_resources/create.tf", map[string]any{
		"spaceId":      spaceId,
		"name":         name,
		"description":  description,
		"environments": []string{"notexisting"},
	})
}

func testApiKeyUpdate(spaceId string, name string, description string) string {
	return utils.HCLTemplateFromPath("test_resources/update.tf", map[string]any{
		"spaceId":     spaceId,
		"name":        name,
		"description": description,
	})
}

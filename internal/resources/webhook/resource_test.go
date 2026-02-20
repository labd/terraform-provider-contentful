package webhook_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	hashicoracctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/assert"

	"github.com/labd/terraform-provider-contentful/internal/acctest"
	"github.com/labd/terraform-provider-contentful/internal/provider"
	"github.com/labd/terraform-provider-contentful/internal/sdk"
)

type assertFunc func(*testing.T, *sdk.Webhook)

func TestWebhookResource_Basic(t *testing.T) {
	name := fmt.Sprintf("webhook-name-%s", hashicoracctest.RandString(3))
	url := "https://www.example.com/test"
	resourceName := "contentful_webhook.mywebhook"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.TestAccPreCheck(t) },
		CheckDestroy: testAccCheckContentfulWebhookDestroy,
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"contentful": providerserver.NewProtocol6WithError(provider.New("test", true)()),
		},
		Steps: []resource.TestStep{
			{
				Config: testWebhook(os.Getenv("CONTENTFUL_SPACE_ID"), name, url),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "url", url),
					resource.TestCheckResourceAttr(resourceName, "topics.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "headers.header1", "header1-value"),
					testAccCheckContentfulWebhookExists(t, resourceName, func(t *testing.T, webhook *sdk.Webhook) {
						assert.EqualValues(t, name, webhook.Name)
						assert.EqualValues(t, url, webhook.Url)
						assert.Len(t, webhook.Topics, 2)
						assert.Contains(t, webhook.Topics, "Entry.create")
						assert.Contains(t, webhook.Topics, "ContentType.create")
						assert.Len(t, webhook.Headers, 2)
					}),
					resource.TestCheckResourceAttr(resourceName, "active", "true"),
					resource.TestCheckResourceAttr(resourceName, "filters", "[]"),
				),
			},
			{
				Config: testWebhookUpdate(os.Getenv("CONTENTFUL_SPACE_ID"), name, url),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("%s-updated", name)),
					resource.TestCheckResourceAttr(resourceName, "url", fmt.Sprintf("%s-updated", url)),
					resource.TestCheckResourceAttr(resourceName, "topics.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "headers.header1", "header1-value-updated"),
					testAccCheckContentfulWebhookExists(t, resourceName, func(t *testing.T, webhook *sdk.Webhook) {
						assert.EqualValues(t, fmt.Sprintf("%s-updated", name), webhook.Name)
						assert.EqualValues(t, fmt.Sprintf("%s-updated", url), webhook.Url)
						assert.Len(t, webhook.Topics, 3)
						assert.Contains(t, webhook.Topics, "Entry.create")
						assert.Contains(t, webhook.Topics, "ContentType.create")
						assert.Contains(t, webhook.Topics, "Asset.*")
						assert.Len(t, webhook.Headers, 2)
					}),
					resource.TestCheckResourceAttr(resourceName, "active", "false"),
					resource.TestCheckResourceAttr(resourceName, "filters", "[{\"in\":[{\"doc\":\"sys.environment.sys.id\"},[\"testing\",\"staging\"]]},{\"not\":{\"equals\":[{\"doc\":\"sys.environment.sys.id\"},\"master-2026-02-20\"]}}]"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"http_basic_auth_password"},
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources[resourceName]
					if !ok {
						return "", fmt.Errorf("not found: %s", resourceName)
					}
					return fmt.Sprintf("%s:%s", rs.Primary.ID, rs.Primary.Attributes["space_id"]), nil
				},
			},
		},
	})
}

func testAccCheckContentfulWebhookExists(t *testing.T, resourceName string, assertFunc assertFunc) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		webhook, err := getWebhookFromState(s, resourceName)
		if err != nil {
			return err
		}

		assertFunc(t, webhook)
		return nil
	}
}

func getWebhookFromState(s *terraform.State, resourceName string) (*sdk.Webhook, error) {
	rs, ok := s.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("webhook not found in state: %s", resourceName)
	}

	if rs.Primary.ID == "" {
		return nil, fmt.Errorf("no webhook ID found")
	}

	spaceID := rs.Primary.Attributes["space_id"]
	if spaceID == "" {
		return nil, fmt.Errorf("no space_id is set")
	}

	client := acctest.GetClient()
	resp, err := client.GetWebhookWithResponse(context.Background(), spaceID, rs.Primary.ID)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("webhook not found: %s", rs.Primary.ID)
	}

	return resp.JSON200, nil
}

func testAccCheckContentfulWebhookDestroy(s *terraform.State) error {
	client := acctest.GetClient()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "contentful_webhook" {
			continue
		}

		spaceID := rs.Primary.Attributes["space_id"]
		if spaceID == "" {
			return fmt.Errorf("no space_id is set")
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no webhook ID is set")
		}

		resp, err := client.GetWebhookWithResponse(context.Background(), spaceID, rs.Primary.ID)
		if err != nil {
			return err
		}

		if resp.StatusCode() == 404 {
			return nil
		}

		return fmt.Errorf("webhook still exists with id: %s", rs.Primary.ID)
	}

	return nil
}

// Create test HCL template functions - using inline templates for simplicity
// In a real implementation, you'd probably use the template file pattern from the API key tests

func testWebhook(spaceId string, name string, url string) string {
	return fmt.Sprintf(`
resource "contentful_webhook" "mywebhook" {
  space_id = "%s"
  name = "%s"
  url = "%s"
  topics = [
    "Entry.create",
    "ContentType.create"
  ]
  headers = {
    header1 = "header1-value"
    header2 = "header2-value"
  }
}
`, spaceId, name, url)
}

func testWebhookUpdate(spaceId string, name string, url string) string {
	return fmt.Sprintf(`
resource "contentful_webhook" "mywebhook" {
  space_id = "%s"
  active = false	
  name = "%s-updated"
  url = "%s-updated"
  topics = [
    "Entry.create",
    "ContentType.create",
    "Asset.*"
  ]
  headers = {
    header1 = "header1-value-updated"
    header2 = "header2-value-updated"
  }
  filters = jsonencode([
    {in: [{ "doc" : "sys.environment.sys.id" }, ["testing", "staging" ]]},
    { not : {equals: [{ "doc" : "sys.environment.sys.id" }, "master-2026-02-20"]} },
  ])	
}
`, spaceId, name, url)
}

package contentful

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/labd/terraform-provider-contentful/internal/acctest"
	"github.com/labd/terraform-provider-contentful/internal/sdk"
)

func TestAccContentfulWebhook_Basic(t *testing.T) {
	var webhook sdk.Webhook

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccContentfulWebhookDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccContentfulWebhookConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContentfulWebhookExists("contentful_webhook.mywebhook", &webhook),
					testAccCheckContentfulWebhookAttributes(&webhook, map[string]interface{}{
						"space_id":                 spaceID,
						"name":                     "webhook-name",
						"url":                      "https://www.example.com/test",
						"http_basic_auth_username": "username",
					}),
				),
			},
			{
				Config: testAccContentfulWebhookUpdateConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContentfulWebhookExists("contentful_webhook.mywebhook", &webhook),
					testAccCheckContentfulWebhookAttributes(&webhook, map[string]interface{}{
						"space_id":                 spaceID,
						"name":                     "webhook-name-updated",
						"url":                      "https://www.example.com/test-updated",
						"http_basic_auth_username": "username-updated",
					}),
				),
			},
		},
	})
}

func testAccCheckContentfulWebhookExists(n string, webhook *sdk.Webhook) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not Found: %s", n)
		}

		// get space id from resource data
		spaceID := rs.Primary.Attributes["space_id"]
		if spaceID == "" {
			return fmt.Errorf("no space_id is set")
		}

		// check webhook resource id
		if rs.Primary.ID == "" {
			return fmt.Errorf("no webhook ID is set")
		}

		client := acctest.GetClient()
		resp, err := client.GetWebhookWithResponse(context.Background(), rs.Primary.Attributes["space_id"], rs.Primary.ID)
		if err != nil {
			return err
		}

		if resp.StatusCode() != 200 {
			return fmt.Errorf("webhook not found: %s", rs.Primary.ID)
		}

		*webhook = *resp.JSON200

		return nil
	}
}

func testAccCheckContentfulWebhookAttributes(webhook *sdk.Webhook, attrs map[string]interface{}) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		name := attrs["name"].(string)
		if webhook.Name != name {
			return fmt.Errorf("webhook name does not match: %s, %s", webhook.Name, name)
		}

		url := attrs["url"].(string)
		if webhook.Url != url {
			return fmt.Errorf("webhook url does not match: %s, %s", webhook.Url, url)
		}

		/* topics := attrs["topics"].([]string)

		headers := attrs["headers"].(map[string]string) */

		httpBasicAuthUsername := attrs["http_basic_auth_username"].(string)
		if *webhook.HttpBasicUsername != httpBasicAuthUsername {
			return fmt.Errorf("webhook http_basic_auth_username does not match: %s, %s", *webhook.HttpBasicUsername, httpBasicAuthUsername)
		}

		return nil
	}
}

func testAccContentfulWebhookDestroy(s *terraform.State) error {
	client := acctest.GetClient()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "contentful_webhook" {
			continue
		}

		// get space id from resource data
		spaceID := rs.Primary.Attributes["space_id"]
		if spaceID == "" {
			return fmt.Errorf("no space_id is set")
		}

		// check webhook resource id
		if rs.Primary.ID == "" {
			return fmt.Errorf("no webhook ID is set")
		}

		// sdk client
		resp, err := client.GetWebhookWithResponse(context.Background(), rs.Primary.Attributes["space_id"], rs.Primary.ID)
		if err != nil {
			return err
		}

		if resp.StatusCode() == 404 {
			return nil
		}

		return fmt.Errorf("webhook still exists with id: %s (%d)", rs.Primary.ID, resp.StatusCode())
	}

	return nil
}

var testAccContentfulWebhookConfig = `
resource "contentful_webhook" "mywebhook" {
  space_id = "` + spaceID + `"

  name = "webhook-name"
  url=  "https://www.example.com/test"
  topics = [
	"Entry.create",
	"ContentType.create",
  ]
  headers = {
	header1 = "header1-value"
    header2 = "header2-value"
  }
  http_basic_auth_username = "username"
  http_basic_auth_password = "password"
}
`

var testAccContentfulWebhookUpdateConfig = `
resource "contentful_webhook" "mywebhook" {
  space_id = "` + spaceID + `"


  name = "webhook-name-updated"
  url=  "https://www.example.com/test-updated"
  topics = [
	"Entry.create",
	"ContentType.create",
	"Asset.*",
  ]
  headers = {
	header1 = "header1-value-updated"
    header2 = "header2-value-updated"
  }
  http_basic_auth_username = "username-updated"
  http_basic_auth_password = "password-updated"
}
`

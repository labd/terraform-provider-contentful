package main

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	contentful "github.com/labd/contentful-go"
)

func TestAccContentfulWebhook_Basic(t *testing.T) {
	var webhook contentful.Webhook

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
						"name":                     "webhook-name-updated",
						"url":                      "https://www.example.com/test-updated",
						"http_basic_auth_username": "username-updated",
					}),
				),
			},
		},
	})
}

func testAccCheckContentfulWebhookExists(n string, webhook *contentful.Webhook) resource.TestCheckFunc {
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

		client := testAccProvider.Meta().(*contentful.Client)

		contentfulWebhook, err := client.Webhooks.Get(spaceID, rs.Primary.ID)
		if err != nil {
			return err
		}

		*webhook = *contentfulWebhook

		return nil
	}
}

func testAccCheckContentfulWebhookAttributes(webhook *contentful.Webhook, attrs map[string]interface{}) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		name := attrs["name"].(string)
		if webhook.Name != name {
			return fmt.Errorf("webhook name does not match: %s, %s", webhook.Name, name)
		}

		url := attrs["url"].(string)
		if webhook.URL != url {
			return fmt.Errorf("webhook url does not match: %s, %s", webhook.URL, url)
		}

		/* topics := attrs["topics"].([]string)

		headers := attrs["headers"].(map[string]string) */

		httpBasicAuthUsername := attrs["http_basic_auth_username"].(string)
		if webhook.HTTPBasicUsername != httpBasicAuthUsername {
			return fmt.Errorf("webhook http_basic_auth_username does not match: %s, %s", webhook.HTTPBasicUsername, httpBasicAuthUsername)
		}

		return nil
	}
}

func testAccContentfulWebhookDestroy(s *terraform.State) error {
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
		client := testAccProvider.Meta().(*contentful.Client)

		_, err := client.Webhooks.Get(spaceID, rs.Primary.ID)
		if _, ok := err.(contentful.NotFoundError); ok {
			return nil
		}

		return fmt.Errorf("webhook still exists with id: %s", rs.Primary.ID)
	}

	return nil
}

var testAccContentfulWebhookConfig = `
resource "contentful_space" "myspace" {
  name = "space-name"
}

resource "contentful_webhook" "mywebhook" {
  space_id = "${contentful_space.myspace.id}"
  depends_on = ["contentful_space.myspace"]

  name = "webhook-name"
  url=  "https://www.example.com/test"
  topics = [
	"Entry.create",
	"ContentType.create",
  ]
  headers {
	header1 = "header1-value"
    header2 = "header2-value"
  }
  http_basic_auth_username = "username"
  http_basic_auth_password = "password"
}
`

var testAccContentfulWebhookUpdateConfig = `
resource "contentful_space" "myspace" {
  name = "space-name"
}

resource "contentful_webhook" "mywebhook" {
  depends_on = ["contentful_space.myspace"]
  space_id = "${contentful_space.myspace.id}"

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

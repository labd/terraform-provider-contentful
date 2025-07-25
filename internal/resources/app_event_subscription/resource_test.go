package app_event_subscription_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	hashicor_acctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/labd/terraform-provider-contentful/internal/acctest"
	"github.com/labd/terraform-provider-contentful/internal/provider"
)

func TestAppEventSubscriptionResource_Basic(t *testing.T) {
	name := fmt.Sprintf("locale-name-%s", hashicor_acctest.RandString(3))
	resourceName := "contentful_app_event_subscription.myapp_event_subscription"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.TestAccPreCheck(t) },
		CheckDestroy: testAccCheckContentfulLocaleDestroy,
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"contentful": providerserver.NewProtocol6WithError(provider.New("test", true)()),
		},
		Steps: []resource.TestStep{
			{
				Config: testLocaleConfig(name, "https://example.com/webhook", "Entry.save"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "target_url", "https://example.com/webhook"),
					resource.TestCheckResourceAttr(resourceName, "topics.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "topics.0", "Entry.save"),
				),
			},
			{
				Config: testLocaleConfig(name, "https://example.com/webhook-other", "Entry.create"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "target_url", "https://example.com/webhook-other"),
					resource.TestCheckResourceAttr(resourceName, "topics.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "topics.0", "Entry.create"),
				),
			},
		},
	})
}

func testAccCheckContentfulLocaleDestroy(s *terraform.State) error {
	client := acctest.GetClient()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "contentful_app_event_subscription" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		parts := strings.Split(rs.Primary.ID, ":")
		if len(parts) != 2 {
			return fmt.Errorf("invalid ID format, expected 'organizationID:subscriptionID', got: %s", rs.Primary.ID)
		}

		resp, err := client.GetAppEventSubscriptionWithResponse(context.Background(), parts[0], parts[1])
		if err != nil {
			return err
		}

		if resp.StatusCode() == 404 {
			return nil
		}

		return fmt.Errorf("contentful_app_event_subscription still exists with id: %s", rs.Primary.ID)
	}

	return nil
}

func testLocaleConfig(name, targetUrl, topic string) string {
	return fmt.Sprintf(`
resource "contentful_app_definition" "myapp_definition" {
  name       = "%s"
  use_bundle = false
  src        = "http://localhost:3000"
  locations = [{ location = "app-config" }, { location = "dialog" }, { location = "entry-editor" }]
}

resource "contentful_app_event_subscription" "myapp_event_subscription" {
  app_definition_id = contentful_app_definition.myapp_definition.id
  target_url        = "%s"
  topics = [ "%s" ]
}

`, name, targetUrl, topic)
}

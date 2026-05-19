package oauth_application_test

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

func TestAccOAuthApplicationResource_Basic(t *testing.T) {
	resourceName := "contentful_oauth_application.test"
	name := fmt.Sprintf("tf-acc-test-%d", os.Getpid())

	var app sdk.OAuthApplication

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.TestAccPreCheck(t)
		},
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"contentful": providerserver.NewProtocol6WithError(provider.New("test", false)()),
		},
		CheckDestroy: testAccCheckOAuthApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOAuthApplicationConfig(name, "initial description", "https://example.com/cb"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOAuthApplicationExists(resourceName, &app),
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "description", "initial description"),
					resource.TestCheckResourceAttr(resourceName, "redirect_uri", "https://example.com/cb"),
					resource.TestCheckResourceAttr(resourceName, "confidential", "true"),
					resource.TestCheckResourceAttr(resourceName, "scopes.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "scopes.0", "content_management_manage"),
					resource.TestCheckResourceAttrSet(resourceName, "client_id"),
					resource.TestCheckResourceAttrSet(resourceName, "client_secret"),
				),
			},
			{
				Config: testAccOAuthApplicationConfig(name, "updated description", "https://example.com/cb2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOAuthApplicationExists(resourceName, &app),
					resource.TestCheckResourceAttr(resourceName, "description", "updated description"),
					resource.TestCheckResourceAttr(resourceName, "redirect_uri", "https://example.com/cb2"),
				),
			},
		},
	})
}

func testAccCheckOAuthApplicationExists(n string, app *sdk.OAuthApplication) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		client := acctest.GetClient()
		orgID := os.Getenv("CONTENTFUL_ORGANIZATION_ID")
		resp, err := client.GetOAuthApplicationWithResponse(context.Background(), orgID, rs.Primary.ID)
		if err := utils.CheckClientResponse(resp, err, http.StatusOK); err != nil {
			return fmt.Errorf("error getting OAuth application: %w", err)
		}

		*app = *resp.JSON200
		return nil
	}
}

func testAccCheckOAuthApplicationDestroy(s *terraform.State) error {
	client := acctest.GetClient()
	ctx := context.Background()
	orgID := os.Getenv("CONTENTFUL_ORGANIZATION_ID")

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "contentful_oauth_application" {
			continue
		}
		resp, err := client.GetOAuthApplicationWithResponse(ctx, orgID, rs.Primary.ID)
		if err := utils.CheckClientResponse(resp, err, http.StatusNotFound); err != nil {
			return fmt.Errorf("OAuth application still exists: %w", err)
		}
	}
	return nil
}

func testAccOAuthApplicationConfig(name, description, redirectURI string) string {
	return fmt.Sprintf(`
resource "contentful_oauth_application" "test" {
  name         = %q
  description  = %q
  redirect_uri = %q
  scopes       = ["content_management_manage"]
  confidential = true
}
`, name, description, redirectURI)
}

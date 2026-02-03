package taxonomy_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	hashicor_acctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/labd/terraform-provider-contentful/internal/acctest"
	"github.com/labd/terraform-provider-contentful/internal/provider"
)

func TestTaxonomyConceptResource_Basic(t *testing.T) {
	spaceID := os.Getenv("CONTENTFUL_SPACE_ID")
	conceptSchemeID := fmt.Sprintf("scheme-%s", hashicor_acctest.RandString(6))
	label := fmt.Sprintf("Test Concept %s", hashicor_acctest.RandString(4))
	resourceName := "contentful_taxonomy_concept.test"
	environment := "master"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.TestAccPreCheck(t)
		},
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"contentful": providerserver.NewProtocol6WithError(provider.New("test", false)()),
		},
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccTaxonomyConceptResourceConfig(spaceID, environment, conceptSchemeID, "en", label),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "space_id", spaceID),
					resource.TestCheckResourceAttr(resourceName, "environment", environment),
					resource.TestCheckResourceAttr(resourceName, "concept_scheme_id", conceptSchemeID),
					resource.TestCheckResourceAttr(resourceName, "pref_label.en", label),
				),
			},
			// ImportState testing
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources[resourceName]
					if !ok {
						return "", fmt.Errorf("Not found: %s", resourceName)
					}
					return fmt.Sprintf("%s:%s:%s", rs.Primary.Attributes["space_id"], rs.Primary.Attributes["environment"], rs.Primary.ID), nil
				},
			},
			// Update and Read testing
			{
				Config: testAccTaxonomyConceptResourceConfig(spaceID, environment, conceptSchemeID, "en", "Updated "+label),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "pref_label.en", "Updated "+label),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccTaxonomyConceptResourceConfig(spaceId, environment, conceptSchemeId, locale, label string) string {
	return fmt.Sprintf(`
resource "contentful_taxonomy_concept" "test" {
  space_id         = "%s"
  environment      = "%s"
  concept_scheme_id = "%s"
  pref_label = {
    %s = "%s"
  }
}
`, spaceId, environment, conceptSchemeId, locale, label)
}

package contentful

import (
	"context"
	"github.com/flaconi/contentful-go"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Provider returns the Terraform Provider as a scheme and makes resources reachable
func Provider() func() *schema.Provider {
	return func() *schema.Provider {
		p := &schema.Provider{
			Schema: map[string]*schema.Schema{
				"cma_token": {
					Type:        schema.TypeString,
					Optional:    true,
					Sensitive:   true,
					DefaultFunc: schema.EnvDefaultFunc("CONTENTFUL_MANAGEMENT_TOKEN", nil),
					Description: "The Contentful Management API token",
				},
				"organization_id": {
					Type:        schema.TypeString,
					Optional:    true,
					Sensitive:   true,
					DefaultFunc: schema.EnvDefaultFunc("CONTENTFUL_ORGANIZATION_ID", nil),
					Description: "The organization ID",
				},
			},
			ResourcesMap: map[string]*schema.Resource{
				"contentful_space":       resourceContentfulSpace(),
				"contentful_apikey":      resourceContentfulAPIKey(),
				"contentful_webhook":     resourceContentfulWebhook(),
				"contentful_locale":      resourceContentfulLocale(),
				"contentful_environment": resourceContentfulEnvironment(),
				"contentful_entry":       resourceContentfulEntry(),
				"contentful_asset":       resourceContentfulAsset(),
			},
			ConfigureContextFunc: providerConfigure,
		}

		return p
	}

}

// providerConfigure sets the configuration for the Terraform Provider
func providerConfigure(_ context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	cma := contentful.NewCMA(d.Get("cma_token").(string))
	cma.SetOrganization(d.Get("organization_id").(string))

	if logBoolean != "" {
		cma.Debug = true
	}

	return cma, nil
}

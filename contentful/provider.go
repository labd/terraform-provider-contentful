package contentful

import (
	"context"

	contentful "github.com/flaconi/contentful-go"
	client2 "github.com/flaconi/contentful-go/pkgs/client"
	"github.com/flaconi/contentful-go/pkgs/util"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/labd/terraform-provider-contentful/internal/utils"
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
				"base_url": {
					Type:        schema.TypeString,
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("CONTENTFUL_BASE_URL", "https://api.contentful.com"),
					Description: "The base url to use for the Contentful API. Defaults to https://api.contentful.com",
				},
				"environment": {
					Type:        schema.TypeString,
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("CONTENTFUL_ENVIRONMENT", "master"),
					Description: "The environment to use for the Contentful API. Defaults to master",
				},
			},
			ResourcesMap: map[string]*schema.Resource{
				"contentful_space":       resourceContentfulSpace(),
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
	cma.BaseURL = d.Get("base_url").(string)
	cma.SetEnvironment(d.Get("environment").(string))

	debug := false

	if logBoolean != "" {
		cma.Debug = true
		debug = true
	}

	client, err := contentful.NewCMAV2(client2.ClientConfig{
		Debug:     debug,
		UserAgent: util.ToPointer("terraform-provider-contentful"),
		Token:     d.Get("cma_token").(string),
	})

	if err != nil {
		return nil, diag.FromErr(err)
	}

	data := utils.ProviderData{
		Client:         cma,
		CMAClient:      client,
		OrganizationId: d.Get("organization_id").(string),
	}

	return data, nil
}

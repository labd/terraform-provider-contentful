package contentful

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
	contentful "github.com/labd/contentful-go"
)

// Provider returns the Terraform Provider as a scheme and makes resources reachable
func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"cma_token": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("CONTENTFUL_MANAGEMENT_TOKEN", nil),
				Description: "The Contentful Management API token",
			},
			"organization_id": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("CONTENTFUL_ORGANIZATION_ID", nil),
				Description: "The organization ID",
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"contentful_space":       resourceContentfulSpace(),
			"contentful_contenttype": resourceContentfulContentType(),
			"contentful_apikey":      resourceContentfulAPIKey(),
			"contentful_webhook":     resourceContentfulWebhook(),
			"contentful_locale":      resourceContentfulLocale(),
			"contentful_environment": resourceContentfulEnvironment(),
			"contentful_entry":       resourceContentfulEntry(),
			"contentful_asset":       resourceContentfulAsset(),
			"contentful_upload":      resourceContentfulUpload(),
		},
		ConfigureFunc: providerConfigure,
	}
}

// providerConfigure sets the configuration for the Terraform Provider
func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	cma := contentful.NewCMA(d.Get("cma_token").(string))
	cma.SetOrganization(d.Get("organization_id").(string))

	if logBoolean != "" {
		cma.Debug = true
	}

	return cma, nil
}

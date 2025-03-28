package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/labd/terraform-provider-contentful/internal/resources/api_key"
	"github.com/labd/terraform-provider-contentful/internal/resources/app_definition"
	"github.com/labd/terraform-provider-contentful/internal/resources/app_installation"
	"github.com/labd/terraform-provider-contentful/internal/resources/asset"
	"github.com/labd/terraform-provider-contentful/internal/resources/contenttype"
	"github.com/labd/terraform-provider-contentful/internal/resources/entry"
	"github.com/labd/terraform-provider-contentful/internal/resources/environment"
	"github.com/labd/terraform-provider-contentful/internal/resources/locale"
	"github.com/labd/terraform-provider-contentful/internal/resources/space"
	"github.com/labd/terraform-provider-contentful/internal/resources/webhook"
	"github.com/labd/terraform-provider-contentful/internal/utils"
)

var (
	_ provider.Provider = &contentfulProvider{}
)

func New(version string, debug bool) func() provider.Provider {
	return func() provider.Provider {
		return &contentfulProvider{
			version: version,
			debug:   debug,
		}
	}
}

type contentfulProvider struct {
	version string
	debug   bool
}

// Provider schema struct
type contentfulProviderModel struct {
	CmaToken       types.String `tfsdk:"cma_token"`
	OrganizationId types.String `tfsdk:"organization_id"`
	BaseURL        types.String `tfsdk:"base_url"`
	Environment    types.String `tfsdk:"environment"`
}

func (c contentfulProvider) Metadata(_ context.Context, _ provider.MetadataRequest, response *provider.MetadataResponse) {
	response.TypeName = "contentful"
}

func (c contentfulProvider) Schema(_ context.Context, _ provider.SchemaRequest, response *provider.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"cma_token": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "The Contentful Management API token",
			},
			"organization_id": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "The organization ID",
			},
			"base_url": schema.StringAttribute{
				Optional:    true,
				Description: "The base url to use for the Contentful API. Defaults to https://api.contentful.com",
			},
			"environment": schema.StringAttribute{
				Optional:    true,
				Description: "The environment to use for the Contentful API. Defaults to master",
			},
		},
	}
}

func (c contentfulProvider) Configure(ctx context.Context, request provider.ConfigureRequest, response *provider.ConfigureResponse) {
	var config contentfulProviderModel

	diags := request.Config.Get(ctx, &config)
	response.Diagnostics.Append(diags...)

	if response.Diagnostics.HasError() {
		return
	}

	var cmaToken string

	if config.CmaToken.IsUnknown() || config.CmaToken.IsNull() {
		cmaToken = os.Getenv("CONTENTFUL_MANAGEMENT_TOKEN")
	} else {
		cmaToken = config.CmaToken.ValueString()
	}

	var organizationId string

	if config.OrganizationId.IsUnknown() || config.OrganizationId.IsNull() {
		organizationId = os.Getenv("CONTENTFUL_ORGANIZATION_ID")
	} else {
		organizationId = config.OrganizationId.ValueString()
	}

	var baseURL string
	if config.BaseURL.IsUnknown() || config.BaseURL.IsNull() {
		value, isSet := os.LookupEnv("CONTENTFUL_BASE_URL")
		if isSet {
			baseURL = value
		} else {
			baseURL = "https://api.contentful.com"
		}
	} else {
		baseURL = config.BaseURL.ValueString()
	}

	clientNew, err := utils.CreateClient(baseURL, cmaToken)
	if err != nil {
		panic(err)
	}

	clientUpload, err := utils.CreateClient("https://upload.contentful.com", cmaToken)
	if err != nil {
		panic(err)
	}

	data := utils.ProviderData{
		Client:         clientNew,
		ClientUpload:   clientUpload,
		OrganizationId: organizationId,
	}

	response.ResourceData = data
	response.DataSourceData = data
}

func (c contentfulProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

func (c contentfulProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		api_key.NewApiKeyResource,
		app_definition.NewAppDefinitionResource,
		app_installation.NewAppInstallationResource,
		asset.NewAssetResource,
		contenttype.NewContentTypeResource,
		entry.NewEntryResource,
		environment.NewEnvironmentResource,
		locale.NewLocaleResource,
		space.NewSpaceResource,
		webhook.NewWebhookResource,
	}
}

package space

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"

	"github.com/labd/terraform-provider-contentful/internal/sdk"
	"github.com/labd/terraform-provider-contentful/internal/utils"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource = &spaceDataSource{}
)

func NewSpaceDataSource() datasource.DataSource {
	return &spaceDataSource{}
}

// spaceDataSource is the resource implementation.
type spaceDataSource struct {
	client *sdk.ClientWithResponses
}

func (e *spaceDataSource) Metadata(_ context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_space"
}

func (e *spaceDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "A Contentful Space represents a space in Contentful.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:    true,
				Description: "Space ID",
			},
			"version": schema.Int64Attribute{
				Computed:    true,
				Description: "The current version of the space",
			},
			"name": schema.StringAttribute{
				Computed:    true,
				Description: "Name of the space",
			},
			"default_locale": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Default locale for the space",
			},
		},
	}
}

func (e *spaceDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	data := request.ProviderData.(utils.ProviderData)
	e.client = data.Client
}

func (e *spaceDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	data := &SpaceData{}

	// Read Terraform configuration data into the model
	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)

	resp, err := e.client.GetSpaceWithResponse(ctx, data.ID.ValueString())
	if err != nil {
		response.Diagnostics.AddError(
			"Error reading space",
			"Could not read space: "+err.Error(),
		)
		return
	}

	// Handle 404 Not Found
	if resp.StatusCode() == 404 {
		response.Diagnostics.AddError(
			"Space not found",
			fmt.Sprintf("Space %s was not found", data.ID.ValueString()),
		)
		return
	}

	if resp.StatusCode() != 200 {
		response.Diagnostics.AddError(
			"Error reading space",
			fmt.Sprintf("Received unexpected status code: %d", resp.StatusCode()),
		)
		return
	}

	// Map response to state
	data.Import(resp.JSON200)

	// Set state
	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

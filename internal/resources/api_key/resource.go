package api_key

import (
	"context"
	"fmt"
	"github.com/flaconi/contentful-go/pkgs/model"
	"github.com/flaconi/contentful-go/service/cma"
	"github.com/flaconi/terraform-provider-contentful/internal/custommodifier"
	"github.com/flaconi/terraform-provider-contentful/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"strings"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &apiKeyResource{}
	_ resource.ResourceWithConfigure   = &apiKeyResource{}
	_ resource.ResourceWithImportState = &apiKeyResource{}
)

func NewApiKeyResource() resource.Resource {
	return &apiKeyResource{}
}

// apiKeyResource is the resource implementation.
type apiKeyResource struct {
	client cma.SpaceIdClientBuilder
}

func (e *apiKeyResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_apikey"
}

func (e *apiKeyResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {

	response.Schema = schema.Schema{
		Description: "Todo for explaining apikey",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "api key id",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"preview_id": schema.StringAttribute{
				Computed:    true,
				Description: "preview api key id",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"version": schema.Int64Attribute{
				Computed: true,
			},
			"space_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description: "space id",
			},
			"access_token": schema.StringAttribute{
				Computed:  true,
				Sensitive: true,
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"description": schema.StringAttribute{
				Optional: true,
			},
			"preview_token": schema.StringAttribute{
				Computed:  true,
				Sensitive: true,
			},
			"environments": schema.ListAttribute{
				Optional:    true,
				Computed:    true,
				Description: "List of needed environments if not added then master is used",
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.List{
					custommodifier.ListDefault([]attr.Value{types.StringValue("master")}),
				},
			},
		},
	}
}

func (e *apiKeyResource) Configure(_ context.Context, request resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	data := request.ProviderData.(utils.ProviderData)
	e.client = data.CMAClient
}

func (e *apiKeyResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan *ApiKey
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	draft := plan.Draft()

	if err := e.client.WithSpaceId(plan.SpaceId.ValueString()).ApiKeys().Upsert(ctx, draft); err != nil {
		response.Diagnostics.AddError("Error creating api_key", err.Error())
		return
	}

	plan.Import(draft)

	previewApiKeyContentful, err := e.getPreviewApiKey(ctx, plan)
	if err != nil {
		response.Diagnostics.AddError(
			"Error reading preview api key",
			"Could not retrieve preview api key, unexpected error: "+err.Error(),
		)
		return
	}

	plan.PreviewToken = types.StringValue(previewApiKeyContentful.AccessToken)

	// Set state to fully populated data
	response.Diagnostics.Append(response.State.Set(ctx, plan)...)
	if response.Diagnostics.HasError() {
		return
	}
}

func (e *apiKeyResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	// Get current state
	var state *ApiKey
	diags := request.State.Get(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	e.doRead(ctx, state, &response.State, &response.Diagnostics)
}

func (e *apiKeyResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan *ApiKey
	diags := request.Plan.Get(ctx, &plan)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	// Get current state
	var state *ApiKey
	diags = request.State.Get(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	draft := plan.Draft()

	if err := e.client.WithSpaceId(state.SpaceId.ValueString()).ApiKeys().Upsert(ctx, draft); err != nil {
		response.Diagnostics.AddError(
			"Error updating api key",
			"Could not update api key, unexpected error: "+err.Error(),
		)
		return
	}

	plan.Import(draft)

	previewApiKeyContentful, err := e.getPreviewApiKey(ctx, plan)
	if err != nil {
		response.Diagnostics.AddError(
			"Error reading preview api key",
			"Could not retrieve preview api key, unexpected error: "+err.Error(),
		)
		return
	}

	plan.PreviewToken = types.StringValue(previewApiKeyContentful.AccessToken)

	// Set state to fully populated data
	response.Diagnostics.Append(response.State.Set(ctx, plan)...)
	if response.Diagnostics.HasError() {
		return
	}
}

func (e *apiKeyResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	// Get current state
	var state *ApiKey
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)

	if err := e.client.WithSpaceId(state.SpaceId.ValueString()).ApiKeys().Delete(ctx, state.Draft()); err != nil {
		response.Diagnostics.AddError(
			"Error deleting api_key",
			"Could not delete api_key, unexpected error: "+err.Error(),
		)
		return
	}
}

func (e *apiKeyResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	idParts := strings.SplitN(request.ID, ":", 2)

	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		response.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: apiKeyId:spaceId. Got: %q", request.ID),
		)
		return
	}

	futureState := &ApiKey{
		ID:      types.StringValue(idParts[0]),
		SpaceId: types.StringValue(idParts[1]),
	}

	e.doRead(ctx, futureState, &response.State, &response.Diagnostics)
}

func (e *apiKeyResource) doRead(ctx context.Context, apiKey *ApiKey, state *tfsdk.State, d *diag.Diagnostics) {

	apiKeyContentful, err := e.getApiKey(ctx, apiKey)
	if err != nil {
		d.AddError(
			"Error reading api key",
			"Could not retrieve api key, unexpected error: "+err.Error(),
		)
		return
	}

	apiKey.Import(apiKeyContentful)

	previewApiKeyContentful, err := e.getPreviewApiKey(ctx, apiKey)
	if err != nil {
		d.AddError(
			"Error reading preview api key",
			"Could not retrieve preview api key, unexpected error: "+err.Error(),
		)
		return
	}

	apiKey.PreviewToken = types.StringValue(previewApiKeyContentful.AccessToken)

	// Set refreshed state
	d.Append(state.Set(ctx, &apiKey)...)
	if d.HasError() {
		return
	}
}

func (e *apiKeyResource) getApiKey(ctx context.Context, apiKey *ApiKey) (*model.APIKey, error) {
	return e.client.WithSpaceId(apiKey.SpaceId.ValueString()).ApiKeys().Get(ctx, apiKey.ID.ValueString())
}

func (e *apiKeyResource) getPreviewApiKey(ctx context.Context, apiKey *ApiKey) (*model.PreviewAPIKey, error) {
	return e.client.WithSpaceId(apiKey.SpaceId.ValueString()).PreviewApiKeys().Get(ctx, apiKey.PreviewID.ValueString())
}

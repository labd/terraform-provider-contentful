package space

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/labd/terraform-provider-contentful/internal/sdk"
	"github.com/labd/terraform-provider-contentful/internal/utils"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &spaceResource{}
	_ resource.ResourceWithConfigure   = &spaceResource{}
	_ resource.ResourceWithImportState = &spaceResource{}
)

func NewSpaceResource() resource.Resource {
	return &spaceResource{}
}

// spaceResource is the resource implementation.
type spaceResource struct {
	client *sdk.ClientWithResponses
}

func (e *spaceResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_space"
}

func (e *spaceResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "A Contentful Space represents a space in Contentful.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Space ID",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"version": schema.Int64Attribute{
				Computed:    true,
				Description: "The current version of the space",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the space",
			},
			"default_locale": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Default locale for the space",
				Default:     stringdefault.StaticString("en"),
			},
		},
	}
}

func (e *spaceResource) Configure(_ context.Context, request resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	data := request.ProviderData.(utils.ProviderData)
	e.client = data.Client
}

func (e *spaceResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	// Get plan values
	var plan Space
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Create the space
	draft := plan.DraftForCreate()

	resp, err := e.client.CreateSpaceWithResponse(ctx, nil, draft)
	if err != nil {
		response.Diagnostics.AddError(
			"Error creating space",
			"Could not create space: "+err.Error(),
		)
		return
	}

	if resp.StatusCode() != 201 {
		response.Diagnostics.AddError(
			"Error creating space",
			fmt.Sprintf("Received unexpected status code: %d", resp.StatusCode()),
		)
		return
	}

	// Map response to state
	plan.Import(resp.JSON201)

	// Set state
	response.Diagnostics.Append(response.State.Set(ctx, plan)...)
}

func (e *spaceResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	// Get current state
	var state Space
	diags := request.State.Get(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	e.doRead(ctx, &state, &response.State, &response.Diagnostics)
}

func (e *spaceResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	// Get plan values
	var plan Space
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Get current state
	var state Space
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Create update parameters with version
	params := &sdk.UpdateSpaceParams{
		XContentfulVersion: state.Version.ValueInt64(),
	}

	// Update the space
	draft := plan.DraftForUpdate()
	resp, err := e.client.UpdateSpaceWithResponse(
		ctx,
		state.ID.ValueString(),
		params,
		draft,
	)

	if err != nil {
		response.Diagnostics.AddError(
			"Error updating space",
			"Could not update space: "+err.Error(),
		)
		return
	}

	if resp.StatusCode() != 200 {
		response.Diagnostics.AddError(
			"Error updating space",
			fmt.Sprintf("Received unexpected status code: %d", resp.StatusCode()),
		)
		return
	}

	// Map response to state
	plan.Import(resp.JSON200)

	// Keep the default locale value since it's not returned in the response
	plan.DefaultLocale = state.DefaultLocale

	// Set state
	response.Diagnostics.Append(response.State.Set(ctx, plan)...)
}

func (e *spaceResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	// Get current state
	var state Space
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Create delete parameters with version
	params := &sdk.DeleteSpaceParams{
		XContentfulVersion: state.Version.ValueInt64(),
	}

	// Delete the space
	resp, err := e.client.DeleteSpaceWithResponse(
		ctx,
		state.ID.ValueString(),
		params,
	)

	if err != nil {
		response.Diagnostics.AddError(
			"Error deleting space",
			"Could not delete space: "+err.Error(),
		)
		return
	}

	if resp.StatusCode() != 204 && resp.StatusCode() != 404 {
		response.Diagnostics.AddError(
			"Error deleting space",
			fmt.Sprintf("Received unexpected status code: %d", resp.StatusCode()),
		)
		return
	}
}

func (e *spaceResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)

	// Set a default value for default_locale since it's not returned in the space GET response
	futureState := &Space{
		ID:            types.StringValue(request.ID),
		DefaultLocale: types.StringValue("en"),
	}

	e.doRead(ctx, futureState, &response.State, &response.Diagnostics)
}

func (e *spaceResource) doRead(ctx context.Context, space *Space, state *tfsdk.State, d *diag.Diagnostics) {
	resp, err := e.client.GetSpaceWithResponse(ctx, space.ID.ValueString())
	if err != nil {
		d.AddError(
			"Error reading space",
			"Could not read space: "+err.Error(),
		)
		return
	}

	// Handle 404 Not Found
	if resp.StatusCode() == 404 {
		d.AddWarning(
			"Space not found",
			fmt.Sprintf("Space %s was not found, removing from state",
				space.ID.ValueString()),
		)
		return
	}

	if resp.StatusCode() != 200 {
		d.AddError(
			"Error reading space",
			fmt.Sprintf("Received unexpected status code: %d", resp.StatusCode()),
		)
		return
	}

	// Keep default_locale from input since it's not returned in the response
	defaultLocale := space.DefaultLocale

	// Map response to state
	space.Import(resp.JSON200)

	// Restore default locale
	space.DefaultLocale = defaultLocale

	// Set state
	d.Append(state.Set(ctx, space)...)
}

package environment_alias

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"

	"github.com/labd/terraform-provider-contentful/internal/sdk"
	"github.com/labd/terraform-provider-contentful/internal/utils"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &environmentAliasResource{}
	_ resource.ResourceWithConfigure   = &environmentAliasResource{}
	_ resource.ResourceWithImportState = &environmentAliasResource{}
)

func NewEnvironmentAliasResource() resource.Resource {
	return &environmentAliasResource{}
}

// environmentAliasResource is the resource implementation.
type environmentAliasResource struct {
	client *sdk.ClientWithResponses
}

func (e *environmentAliasResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_environment_alias"
}

func (e *environmentAliasResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "An Environment Alias allows you to reference an environment with a static identifier.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the environment alias (e.g., 'master', 'production')",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"version": schema.Int64Attribute{
				Computed:    true,
				Description: "The current version of the environment alias",
			},
			"space_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the space",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"environment_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the environment this alias points to",
			},
		},
	}
}

func (e *environmentAliasResource) Configure(_ context.Context, request resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	data := request.ProviderData.(utils.ProviderData)
	e.client = data.Client
}

func (e *environmentAliasResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	// Get plan values
	var plan EnvironmentAlias
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Environment aliases use PUT for both create and update (upsert)
	draft := plan.DraftForUpdate()

	resp, err := e.client.UpsertEnvironmentAliasWithResponse(
		ctx,
		plan.SpaceID.ValueString(),
		plan.ID.ValueString(),
		nil, // No version header for create
		draft,
	)
	if err != nil {
		response.Diagnostics.AddError(
			"Error creating environment alias",
			"Could not create environment alias: "+err.Error(),
		)
		return
	}

	if resp.StatusCode() != 200 && resp.StatusCode() != 201 {
		response.Diagnostics.AddError(
			"Error creating environment alias",
			fmt.Sprintf("Received unexpected status code: %d", resp.StatusCode()),
		)
		return
	}

	// Map response to state
	if resp.JSON200 != nil {
		plan.Import(resp.JSON200)
	} else if resp.JSON201 != nil {
		plan.Import(resp.JSON201)
	}

	// Set state
	response.Diagnostics.Append(response.State.Set(ctx, plan)...)
}

func (e *environmentAliasResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	// Get current state
	var state EnvironmentAlias
	diags := request.State.Get(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	e.doRead(ctx, &state, &response.State, &response.Diagnostics)
}

func (e *environmentAliasResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	// Get plan values
	var plan EnvironmentAlias
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Get current state
	var state EnvironmentAlias
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Create update parameters with version
	params := &sdk.UpsertEnvironmentAliasParams{
		XContentfulVersion: state.Version.ValueInt64(),
	}

	// Update the environment alias
	draft := plan.DraftForUpdate()
	resp, err := e.client.UpsertEnvironmentAliasWithResponse(
		ctx,
		plan.SpaceID.ValueString(),
		plan.ID.ValueString(),
		params,
		draft,
	)

	if err != nil {
		response.Diagnostics.AddError(
			"Error updating environment alias",
			"Could not update environment alias: "+err.Error(),
		)
		return
	}

	if resp.StatusCode() != 200 {
		response.Diagnostics.AddError(
			"Error updating environment alias",
			fmt.Sprintf("Received unexpected status code: %d", resp.StatusCode()),
		)
		return
	}

	// Map response to state
	plan.Import(resp.JSON200)

	// Set state
	response.Diagnostics.Append(response.State.Set(ctx, plan)...)
}

func (e *environmentAliasResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	// Get current state
	var state EnvironmentAlias
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Create delete parameters with version
	params := &sdk.DeleteEnvironmentAliasParams{
		XContentfulVersion: state.Version.ValueInt64(),
	}

	// Delete the environment alias
	resp, err := e.client.DeleteEnvironmentAliasWithResponse(
		ctx,
		state.SpaceID.ValueString(),
		state.ID.ValueString(),
		params,
	)

	if err != nil {
		response.Diagnostics.AddError(
			"Error deleting environment alias",
			"Could not delete environment alias: "+err.Error(),
		)
		return
	}

	if resp.StatusCode() != 204 && resp.StatusCode() != 404 {
		response.Diagnostics.AddError(
			"Error deleting environment alias",
			fmt.Sprintf("Received unexpected status code: %d", resp.StatusCode()),
		)
		return
	}
}

func (e *environmentAliasResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	// Import format: space_id/alias_id
	// For example: abc123/master
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (e *environmentAliasResource) doRead(ctx context.Context, alias *EnvironmentAlias, state *tfsdk.State, d *diag.Diagnostics) {
	resp, err := e.client.GetEnvironmentAliasWithResponse(
		ctx,
		alias.SpaceID.ValueString(),
		alias.ID.ValueString(),
	)
	if err != nil {
		d.AddError(
			"Error reading environment alias",
			"Could not read environment alias: "+err.Error(),
		)
		return
	}

	// Handle 404 Not Found
	if resp.StatusCode() == 404 {
		d.AddWarning(
			"Environment alias not found",
			fmt.Sprintf("Environment alias %s in space %s was not found, removing from state",
				alias.ID.ValueString(), alias.SpaceID.ValueString()),
		)
		return
	}

	if resp.StatusCode() != 200 {
		d.AddError(
			"Error reading environment alias",
			fmt.Sprintf("Received unexpected status code: %d", resp.StatusCode()),
		)
		return
	}

	// Map response to state
	alias.Import(resp.JSON200)

	// Set state
	d.Append(state.Set(ctx, alias)...)
}

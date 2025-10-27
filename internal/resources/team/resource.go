package team

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/labd/terraform-provider-contentful/internal/sdk"
	"github.com/labd/terraform-provider-contentful/internal/utils"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &teamResource{}
	_ resource.ResourceWithConfigure   = &teamResource{}
	_ resource.ResourceWithImportState = &teamResource{}
)

func NewTeamResource() resource.Resource {
	return &teamResource{}
}

// teamResource is the resource implementation.
type teamResource struct {
	client         *sdk.ClientWithResponses
	organizationId string
}

func (t *teamResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_team"
}

func (t *teamResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "A Contentful Team represents a team in a Contentful organization.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Team ID",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"version": schema.Int64Attribute{
				Computed:    true,
				Description: "The current version of the team",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the team",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "Description of the team",
			},
		},
	}
}

func (t *teamResource) Configure(_ context.Context, request resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	data := request.ProviderData.(utils.ProviderData)
	t.client = data.Client
	t.organizationId = data.OrganizationId
}

func (t *teamResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	// Get plan values
	var plan Team
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Check if organization ID is configured
	if t.organizationId == "" {
		response.Diagnostics.AddError(
			"Organization ID not configured",
			"The organization_id must be set in the provider configuration to create teams",
		)
		return
	}

	// Create the team
	draft := plan.DraftForCreate()

	resp, err := t.client.CreateTeamWithResponse(ctx, t.organizationId, draft)
	if err != nil {
		response.Diagnostics.AddError(
			"Error creating team",
			"Could not create team: "+err.Error(),
		)
		return
	}

	if resp.StatusCode() != 201 {
		response.Diagnostics.AddError(
			"Error creating team",
			fmt.Sprintf("Received unexpected status code: %d", resp.StatusCode()),
		)
		return
	}

	// Map response to state
	plan.Import(resp.JSON201)

	// Set state
	response.Diagnostics.Append(response.State.Set(ctx, plan)...)
}

func (t *teamResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	// Get current state
	var state Team
	diags := request.State.Get(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	// Check if organization ID is configured
	if t.organizationId == "" {
		response.Diagnostics.AddError(
			"Organization ID not configured",
			"The organization_id must be set in the provider configuration to read teams",
		)
		return
	}

	resp, err := t.client.GetTeamWithResponse(ctx, t.organizationId, state.ID.ValueString())
	if err != nil {
		response.Diagnostics.AddError(
			"Error reading team",
			"Could not read team: "+err.Error(),
		)
		return
	}

	// Handle 404 Not Found
	if resp.StatusCode() == 404 {
		response.Diagnostics.AddWarning(
			"Team not found",
			fmt.Sprintf("Team %s was not found, removing from state",
				state.ID.ValueString()),
		)
		response.State.RemoveResource(ctx)
		return
	}

	if resp.StatusCode() != 200 {
		response.Diagnostics.AddError(
			"Error reading team",
			fmt.Sprintf("Received unexpected status code: %d", resp.StatusCode()),
		)
		return
	}

	// Map response to state
	state.Import(resp.JSON200)

	// Set state
	response.Diagnostics.Append(response.State.Set(ctx, state)...)
}

func (t *teamResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	// Get plan values
	var plan Team
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Get current state
	var state Team
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Check if organization ID is configured
	if t.organizationId == "" {
		response.Diagnostics.AddError(
			"Organization ID not configured",
			"The organization_id must be set in the provider configuration to update teams",
		)
		return
	}

	// Create update parameters with version
	params := &sdk.UpdateTeamParams{
		XContentfulVersion: state.Version.ValueInt64(),
	}

	// Update the team
	draft := plan.DraftForUpdate()
	resp, err := t.client.UpdateTeamWithResponse(
		ctx,
		t.organizationId,
		state.ID.ValueString(),
		params,
		draft,
	)

	if err != nil {
		response.Diagnostics.AddError(
			"Error updating team",
			"Could not update team: "+err.Error(),
		)
		return
	}

	if resp.StatusCode() != 200 {
		response.Diagnostics.AddError(
			"Error updating team",
			fmt.Sprintf("Received unexpected status code: %d", resp.StatusCode()),
		)
		return
	}

	// Map response to state
	plan.Import(resp.JSON200)

	// Set state
	response.Diagnostics.Append(response.State.Set(ctx, plan)...)
}

func (t *teamResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	// Get current state
	var state Team
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Check if organization ID is configured
	if t.organizationId == "" {
		response.Diagnostics.AddError(
			"Organization ID not configured",
			"The organization_id must be set in the provider configuration to delete teams",
		)
		return
	}

	// Create delete parameters with version
	params := &sdk.DeleteTeamParams{
		XContentfulVersion: state.Version.ValueInt64(),
	}

	// Delete the team
	resp, err := t.client.DeleteTeamWithResponse(
		ctx,
		t.organizationId,
		state.ID.ValueString(),
		params,
	)

	if err != nil {
		response.Diagnostics.AddError(
			"Error deleting team",
			"Could not delete team: "+err.Error(),
		)
		return
	}

	if resp.StatusCode() != 204 && resp.StatusCode() != 404 {
		response.Diagnostics.AddError(
			"Error deleting team",
			fmt.Sprintf("Received unexpected status code: %d", resp.StatusCode()),
		)
		return
	}
}

func (t *teamResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)

	// Set the team ID in state and then read the team to populate other attributes
	futureState := &Team{
		ID: types.StringValue(request.ID),
	}

	// Check if organization ID is configured
	if t.organizationId == "" {
		response.Diagnostics.AddError(
			"Organization ID not configured",
			"The organization_id must be set in the provider configuration to import teams",
		)
		return
	}

	resp, err := t.client.GetTeamWithResponse(ctx, t.organizationId, request.ID)
	if err != nil {
		response.Diagnostics.AddError(
			"Error reading team",
			"Could not read team: "+err.Error(),
		)
		return
	}

	if resp.StatusCode() != 200 {
		response.Diagnostics.AddError(
			"Error reading team",
			fmt.Sprintf("Received unexpected status code: %d", resp.StatusCode()),
		)
		return
	}

	// Map response to state
	futureState.Import(resp.JSON200)

	// Set state
	response.Diagnostics.Append(response.State.Set(ctx, futureState)...)
}

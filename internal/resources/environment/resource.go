package environment

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/labd/terraform-provider-contentful/internal/sdk"
	"github.com/labd/terraform-provider-contentful/internal/utils"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &environmentResource{}
	_ resource.ResourceWithConfigure   = &environmentResource{}
	_ resource.ResourceWithImportState = &environmentResource{}
)

func NewEnvironmentResource() resource.Resource {
	return &environmentResource{}
}

// environmentResource is the resource implementation.
type environmentResource struct {
	client *sdk.ClientWithResponses
}

func (e *environmentResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_environment"
}

func (e *environmentResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "A Contentful Environment represents a space environment.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Environment ID",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"version": schema.Int64Attribute{
				Computed:    true,
				Description: "The current version of the environment",
			},
			"space_id": schema.StringAttribute{
				Required:    true,
				Description: "Space ID",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the environment",
			},
		},
	}
}

func (e *environmentResource) Configure(_ context.Context, request resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	data := request.ProviderData.(utils.ProviderData)
	e.client = data.Client
}

func (e *environmentResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	// Get plan values
	var plan Environment
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Create the environment
	draft := plan.DraftForCreate()

	resp, err := e.client.CreateEnvironmentWithResponse(ctx, plan.SpaceId.ValueString(), draft)
	if err != nil {
		response.Diagnostics.AddError(
			"Error creating environment",
			"Could not create environment: "+err.Error(),
		)
		return
	}

	if resp.StatusCode() != 201 {
		response.Diagnostics.AddError(
			"Error creating environment",
			fmt.Sprintf("Received unexpected status code: %d", resp.StatusCode()),
		)
		return
	}

	// Map response to state
	plan.Import(resp.JSON201)

	// Set state
	response.Diagnostics.Append(response.State.Set(ctx, plan)...)
}

func (e *environmentResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	// Get current state
	var state Environment
	diags := request.State.Get(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	e.doRead(ctx, &state, &response.State, &response.Diagnostics)
}

func (e *environmentResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	// Get plan values
	var plan Environment
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Get current state
	var state Environment
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Create update parameters with version
	params := &sdk.UpdateEnvironmentParams{
		XContentfulVersion: state.Version.ValueInt64(),
	}

	// Update the environment
	draft := plan.DraftForUpdate()
	resp, err := e.client.UpdateEnvironmentWithResponse(
		ctx,
		plan.SpaceId.ValueString(),
		state.ID.ValueString(),
		params,
		draft,
	)

	if err != nil {
		response.Diagnostics.AddError(
			"Error updating environment",
			"Could not update environment: "+err.Error(),
		)
		return
	}

	if resp.StatusCode() != 200 {
		response.Diagnostics.AddError(
			"Error updating environment",
			fmt.Sprintf("Received unexpected status code: %d", resp.StatusCode()),
		)
		return
	}

	// Map response to state
	plan.Import(resp.JSON200)

	// Set state
	response.Diagnostics.Append(response.State.Set(ctx, plan)...)
}

func (e *environmentResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	// Get current state
	var state Environment
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Create delete parameters with version
	params := &sdk.DeleteEnvironmentParams{
		XContentfulVersion: state.Version.ValueInt64(),
	}

	// Delete the environment
	resp, err := e.client.DeleteEnvironmentWithResponse(
		ctx,
		state.SpaceId.ValueString(),
		state.ID.ValueString(),
		params,
	)

	if err != nil {
		response.Diagnostics.AddError(
			"Error deleting environment",
			"Could not delete environment: "+err.Error(),
		)
		return
	}

	if resp.StatusCode() != 204 && resp.StatusCode() != 404 {
		response.Diagnostics.AddError(
			"Error deleting environment",
			fmt.Sprintf("Received unexpected status code: %d", resp.StatusCode()),
		)
		return
	}
}

func (e *environmentResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	// Import format: space_id:environment_id
	// For example: abc123:staging
	idParts := strings.SplitN(request.ID, ":", 2)

	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		response.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: space_id:environment_id. Got: %q", request.ID),
		)
		return
	}

	futureState := &Environment{
		SpaceId: types.StringValue(idParts[0]),
		ID:      types.StringValue(idParts[1]),
	}

	e.doRead(ctx, futureState, &response.State, &response.Diagnostics)
}

func (e *environmentResource) doRead(ctx context.Context, environment *Environment, state *tfsdk.State, d *diag.Diagnostics) {
	resp, err := e.client.GetEnvironmentWithResponse(ctx, environment.SpaceId.ValueString(), environment.ID.ValueString())
	if err != nil {
		d.AddError(
			"Error reading environment",
			"Could not read environment: "+err.Error(),
		)
		return
	}

	// Handle 404 Not Found
	if resp.StatusCode() == 404 {
		d.AddWarning(
			"Environment not found",
			fmt.Sprintf("Environment %s in space %s was not found, removing from state",
				environment.ID.ValueString(), environment.SpaceId.ValueString()),
		)
		return
	}

	if resp.StatusCode() != 200 {
		d.AddError(
			"Error reading environment",
			fmt.Sprintf("Received unexpected status code: %d", resp.StatusCode()),
		)
		return
	}

	// Map response to state
	environment.Import(resp.JSON200)

	// Set state
	d.Append(state.Set(ctx, environment)...)
}

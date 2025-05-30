package role

import (
	"context"
	"fmt"
	"net/http"
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
	_ resource.Resource                = &roleResource{}
	_ resource.ResourceWithConfigure   = &roleResource{}
	_ resource.ResourceWithImportState = &roleResource{}
)

func NewRoleResource() resource.Resource {
	return &roleResource{}
}

// roleResource is the resource implementation.
type roleResource struct {
	client *sdk.ClientWithResponses
}

func (e *roleResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_role"
}

func (e *roleResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "A Contentful Role represents user role.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Role ID",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"role_id": schema.StringAttribute{
				Required:    true,
				Description: "Role Identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"version": schema.Int64Attribute{
				Computed:    true,
				Description: "The current version of the role",
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
				Description: "The name of the role",
			},
			"description": schema.StringAttribute{
				Description: "The description of the role",
				Optional:    true,
			},
		},
		Blocks: map[string]schema.Block{
			"permission": schema.ListNestedBlock{
				Description: "The list of permissions defined",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Required:    true,
							Description: "Permission ID",
						},
						"value": schema.StringAttribute{
							Optional:    true,
							Description: "If all are allowed this should be `all`.",
						},
						"values": schema.ListAttribute{
							ElementType: types.StringType,
							Optional:    true,
							Description: "List of permission values, e.g. [\"create\", \"read\"].",
						},
					},
				},
			},
			"policy": schema.ListNestedBlock{
				Description: "The list of policies defined.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"effect": schema.StringAttribute{
							Required:    true,
							Description: "The effect of the policy (e.g., allow or deny).",
						},
						"actions": schema.SingleNestedAttribute{
							Required:    true,
							Description: "Policy action. Use `value` for a single action, or `values` for multiple actions.",
							Attributes: map[string]schema.Attribute{
								"value": schema.StringAttribute{
									Optional:    true,
									Description: "Single action value (e.g., 'all').",
								},
								"values": schema.ListAttribute{
									ElementType: types.StringType,
									Optional:    true,
									Description: "List of action values (e.g., ['read', 'write']).",
								},
							},
						},
						"constraint": schema.StringAttribute{
							Optional:    true,
							Description: "JSON-encoded constraint for the policy.",
						},
					},
				},
			},
		},
	}
}

func (e *roleResource) Configure(_ context.Context, request resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	data := request.ProviderData.(utils.ProviderData)
	e.client = data.Client
}

func (e *roleResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	// Get plan values
	var plan Role
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Create the webhook
	draft := plan.DraftForCreate()

	resp, err := e.client.CreateRoleWithResponse(ctx, plan.SpaceID.ValueString(), draft)
	if err := utils.CheckClientResponse(resp, err, http.StatusCreated); err != nil {
		response.Diagnostics.AddError(
			"Error creating role",
			"Could not create role: "+err.Error(),
		)
		return
	}

	// Map response to state
	state := &Role{}
	state.Import(resp.JSON201)

	// Set state
	response.Diagnostics.Append(response.State.Set(ctx, state)...)
}

func (e *roleResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	// Get current state
	var state Role
	diags := request.State.Get(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, err := e.client.GetRoleWithResponse(ctx, state.SpaceID.ValueString(), state.RoleID.ValueString())
	if err := utils.CheckClientResponse(resp, err, http.StatusOK); err != nil {
		if resp.StatusCode() == 404 {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.AddError("Error reading role", err.Error())
	}

	state.Import(resp.JSON200)

	response.Diagnostics.Append(request.State.Set(ctx, &state)...)
}

func (e *roleResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	// Get plan values
	var plan Role
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Get current state
	var state Role
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Create update parameters with version
	params := &sdk.UpdateRoleParams{
		XContentfulVersion: state.Version.ValueInt64(),
	}

	// Update the role
	draft := plan.DraftForUpdate()
	resp, err := e.client.UpdateRoleWithResponse(
		ctx,
		plan.SpaceID.ValueString(),
		plan.RoleID.ValueString(),
		params,
		draft,
	)

	if err := utils.CheckClientResponse(resp, err, http.StatusOK); err != nil {
		response.Diagnostics.AddError(
			"Error updating role",
			"Could not update role: "+err.Error(),
		)
		return
	}

	state.Import(resp.JSON200)

	// Set state
	response.Diagnostics.Append(response.State.Set(ctx, state)...)
}

func (e *roleResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	// Get current state
	var state Role
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Get latest version first to avoid conflicts
	resp, err := e.client.GetRoleWithResponse(
		ctx,
		state.SpaceID.ValueString(),
		state.ID.ValueString(),
	)
	if err := utils.CheckClientResponse(resp, err, http.StatusOK); err != nil {
		if resp.StatusCode() == 404 {
			// Role already deleted, nothing to do
			return
		}

		response.Diagnostics.AddError(
			"Error deleting role",
			"Could not get latest role version: "+err.Error(),
		)
		return
	}

	state.Import(resp.JSON200)

	// Create delete parameters with latest version
	params := &sdk.DeleteRoleParams{
		XContentfulVersion: int64(state.Version.ValueInt64()),
	}

	// Delete the role
	deleteResp, err := e.client.DeleteRoleWithResponse(
		ctx,
		state.SpaceID.ValueString(),
		state.ID.ValueString(),
		params,
	)

	if err != nil {
		response.Diagnostics.AddError(
			"Error deleting role",
			"Could not delete role: "+err.Error(),
		)
		return
	}

	if deleteResp.StatusCode() != 204 && deleteResp.StatusCode() != 404 {
		response.Diagnostics.AddError(
			"Error deleting role",
			fmt.Sprintf("Received unexpected status code: %d", deleteResp.StatusCode()),
		)
		return
	}
}

func (e *roleResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	// Extract the role ID and space ID from the import ID
	idParts := strings.Split(request.ID, ":")
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		response.Diagnostics.AddError(
			"Error importing role",
			fmt.Sprintf("Expected import format: role_id:space_id, got: %s", request.ID),
		)
		return
	}

	roleID := idParts[0]
	spaceID := idParts[1]

	resp, err := e.client.GetRoleWithResponse(ctx, spaceID, roleID)

	if err := utils.CheckClientResponse(resp, err, http.StatusOK); err != nil {
		response.Diagnostics.AddError(
			"Error importing asset",
			fmt.Sprintf("Could not import role with ID %s: %s", roleID, err.Error()),
		)
		return
	}

	role := Role{}
	role.Import(resp.JSON200)

	response.Diagnostics.Append(response.State.Set(ctx, role)...)
}

func (e *roleResource) doRead(ctx context.Context, role *Role, state *tfsdk.State, d *diag.Diagnostics) {
	resp, err := e.client.GetRoleWithResponse(
		ctx,
		role.SpaceID.ValueString(),
		role.ID.ValueString(),
	)

	if err != nil {
		d.AddError(
			"Error reading role",
			"Could not read role: "+err.Error(),
		)
		return
	}

	// Handle 404 Not Found
	if resp.StatusCode() == 404 {
		d.AddWarning(
			"Role not found",
			fmt.Sprintf("role %s was not found, removing from state",
				role.ID.ValueString()),
		)
		return
	}

	if resp.StatusCode() != 200 {
		d.AddError(
			"Error reading role",
			fmt.Sprintf("Received unexpected status code: %d", resp.StatusCode()),
		)
		return
	}

	// Map response to state
	role.Import(resp.JSON200)

	// Set state
	d.Append(state.Set(ctx, role)...)
}

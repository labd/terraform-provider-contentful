package role

import (
	"context"
	"fmt"
	"net/http"
	"strings"

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
				Computed:    true,
				Optional:    true,
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
	var plan Role

	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	draft := plan.DraftForCreate()

	resp, err := e.client.CreateRoleWithResponse(ctx, plan.SpaceID.ValueString(), draft)
	if err := utils.CheckClientResponse(resp, err, http.StatusCreated); err != nil {
		response.Diagnostics.AddError(
			"Error creating role",
			"Could not create role: "+err.Error(),
		)
		return
	}

	state := &Role{}
	err = state.Import(resp.JSON201)
	if err != nil {
		response.Diagnostics.AddError(
			"Error creating role",
			"Could not parse response: "+err.Error(),
		)
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, state)...)
}

func (e *roleResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state Role

	diags := request.State.Get(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, err := e.client.GetRoleWithResponse(ctx, state.SpaceID.ValueString(), state.ID.ValueString())
	if err := utils.CheckClientResponse(resp, err, http.StatusOK); err != nil {
		if resp.StatusCode() == 404 {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.AddError("Error reading role", err.Error())
	}

	err = state.Import(resp.JSON200)
	if err != nil {
		response.Diagnostics.AddError(
			"Error reading role",
			"Could not parse response: "+err.Error(),
		)
		return
	}

	response.Diagnostics.Append(request.State.Set(ctx, &state)...)
}

func (e *roleResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
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
		plan.ID.ValueString(),
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

	err = state.Import(resp.JSON200)
	if err != nil {
		response.Diagnostics.AddError(
			"Error updating role",
			"Could not parse response: "+err.Error(),
		)
		return
	}

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

	err = state.Import(resp.JSON200)
	if err != nil {
		response.Diagnostics.AddError(
			"Error deleting role",
			"Could not parse response: "+err.Error(),
		)
		return
	}

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
	// Extract the role ID and space ID from the import
	idParts := strings.Split(request.ID, ":")
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		response.Diagnostics.AddError(
			"Error importing role",
			fmt.Sprintf("Expected import format: role_id:space_id, got: %s", request.ID),
		)
		return
	}

	roleId := idParts[0]
	spaceID := idParts[1]

	resp, err := e.client.GetRoleWithResponse(ctx, spaceID, roleId)

	if err := utils.CheckClientResponse(resp, err, http.StatusOK); err != nil {
		response.Diagnostics.AddError(
			"Error importing asset",
			fmt.Sprintf("Could not import role with ID %s:%s", roleId, err.Error()),
		)
		return
	}

	role := Role{}
	err = role.Import(resp.JSON200)
	if err != nil {
		response.Diagnostics.AddError(
			"Error importing role",
			"Could not parse response: "+err.Error(),
		)
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, role)...)
}

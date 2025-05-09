package role

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
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
			"description": schema.StringAttribute{
				Optional: true,
			},
			"published": schema.BoolAttribute{
				Required:    true,
				Description: "Whether the role is published",
			},
			"archived": schema.BoolAttribute{
				Required:    true,
				Description: "Whether the role is archived",
			},
			"permissions": schema.MapNestedAttribute{
				Required:    true,
				Description: "The list of permissions defined",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"all": schema.StringAttribute{
							Optional: true, // For "all"
						},
						"actions": schema.ListAttribute{
							ElementType: types.StringType, // For list of actions
							Optional:    true,
						},
					},
				},
			},
			"policies": schema.ListNestedAttribute{
				Required: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"effect": schema.StringAttribute{
							Required: true,
						},
						"actions": schema.ListAttribute{
							ElementType: types.StringType,
							Required:    true,
						},
						"constraint": schema.MapNestedAttribute{
							Optional: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"and": schema.ListAttribute{
										ElementType: types.ListType{
											ElemType: types.StringType,
										},
										Optional: true,
									},
								},
							},
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

	// Create the role
	draft := plan.Draft()

	var role *sdk.Role
	if plan.RoleID.IsUnknown() || plan.RoleID.IsNull() {
		params := &sdk.CreateRoleParams{
			XContentfulContentType: plan.ContentTypeID.ValueString(),
		}
		resp, err := e.client.CreateRoleWithResponse(ctx, plan.SpaceID.ValueString(), plan.Environment.ValueString(), params, draft)
		if err := utils.CheckClientResponse(resp, err, http.StatusCreated); err != nil {
			response.Diagnostics.AddError(
				"Error creating role",
				"Could not create role: "+err.Error(),
			)
			return
		}

		role = resp.JSON201
	} else {
		params := &sdk.UpdateRoleParams{
			XContentfulContentType: plan.ContentTypeID.ValueString(),
		}
		resp, err := e.client.UpdateRoleWithResponse(ctx, plan.SpaceID.ValueString(), plan.Environment.ValueString(), plan.RoleID.ValueString(), params, draft)
		if err := utils.CheckClientResponse(resp, err, http.StatusCreated); err != nil {
			response.Diagnostics.AddError(
				"Error creating role",
				"Could not create role: "+err.Error(),
			)
			return
		}

		role = resp.JSON201
	}

	// Map response to state
	state := Role{}
	state.Import(role)

	// Set role state (published/archived)
	if err := e.setRoleState(ctx, &state, &plan); err != nil {
		response.Diagnostics.AddError(
			"Error setting role state",
			err.Error(),
		)
		return
	}

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

	e.doRead(ctx, &state, &response.State, &response.Diagnostics)
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
	draft := plan.Draft()
	resp, err := e.client.UpdateRoleWithResponse(
		ctx,
		plan.SpaceID.ValueString(),
		plan.Environment.ValueString(),
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

	state.Import(resp.JSON200)

	// Set role state (published/archived)
	if err := e.setRoleState(ctx, &state, &plan); err != nil {
		response.Diagnostics.AddError(
			"Error setting role state",
			err.Error(),
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
		state.Environment.ValueString(),
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

	if state.Published.ValueBool() {
		resp, err := e.client.UnpublishRoleWithResponse(
			ctx,
			state.SpaceID.ValueString(),
			state.Environment.ValueString(),
			state.ID.ValueString(),
			&sdk.UnpublishRoleParams{
				XContentfulVersion: state.Version.ValueInt64(),
			},
		)
		if err := utils.CheckClientResponse(resp, err, http.StatusOK); err != nil {
			response.Diagnostics.AddError(
				"Error deleting role",
				"Could not unpublish role before deletion: "+err.Error(),
			)
			return
		}
		state.Import(resp.JSON200)
	}

	// Create delete parameters with latest version
	params := &sdk.DeleteRoleParams{
		XContentfulVersion: int64(state.Version.ValueInt64()),
	}

	// Delete the role
	deleteResp, err := e.client.DeleteRoleWithResponse(
		ctx,
		state.SpaceID.ValueString(),
		state.Environment.ValueString(),
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
	// Extract the role ID, space ID, and environment ID from the import ID
	idParts := strings.Split(request.ID, ":")
	if len(idParts) != 3 || idParts[0] == "" || idParts[1] == "" || idParts[2] == "" {
		response.Diagnostics.AddError(
			"Error importing role",
			fmt.Sprintf("Expected import format: role_id:space_id:environment, got: %s", request.ID),
		)
		return
	}

	roleID := idParts[0]
	spaceID := idParts[1]
	environment := idParts[2]

	role := Role{
		ID:          types.StringValue(roleID),
		RoleID:      types.StringValue(roleID),
		SpaceID:     types.StringValue(spaceID),
		Environment: types.StringValue(environment),
	}

	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("id"), roleID)...)
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("role_id"), roleID)...)
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("space_id"), spaceID)...)
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("environment"), environment)...)

	e.doRead(ctx, &role, &response.State, &response.Diagnostics)
}

func (e *roleResource) doRead(ctx context.Context, role *Role, state *tfsdk.State, d *diag.Diagnostics) {
	resp, err := e.client.GetRoleWithResponse(
		ctx,
		role.SpaceID.ValueString(),
		role.Environment.ValueString(),
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

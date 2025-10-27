package space

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
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
			"deletion_protection": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Allow deletion of the space",
				Default:     booldefault.StaticBool(false),
			},
			"admin_role_id": schema.StringAttribute{
				Computed:    true,
				Description: "ID of the automatically created admin role for this space",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
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

	// Fetch the admin role ID
	adminRoleID, err := e.getAdminRoleID(ctx, plan.ID.ValueString())
	if err != nil {
		response.Diagnostics.AddWarning(
			"Could not fetch admin role ID",
			"Space was created successfully, but could not fetch admin role ID: "+err.Error(),
		)
		// Set admin_role_id to unknown since we couldn't fetch it
		plan.AdminRoleID = types.StringUnknown()
	} else {
		plan.AdminRoleID = types.StringValue(adminRoleID)
	}

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

	// Deletion protection
	if state.AllowDeletion.ValueBool() {
		response.Diagnostics.AddError(
			"Space deletion not allowed",
			"Space deletion is not allowed. Set allow_deletion to true to be able to delete the space.",
		)
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

	// Fetch the admin role ID
	adminRoleID, err := e.getAdminRoleID(ctx, space.ID.ValueString())
	if err != nil {
		d.AddWarning(
			"Could not fetch admin role ID",
			"Space was read successfully, but could not fetch admin role ID: "+err.Error(),
		)
		// Keep existing admin_role_id if we can't fetch it
		if space.AdminRoleID.IsNull() || space.AdminRoleID.IsUnknown() {
			space.AdminRoleID = types.StringUnknown()
		}
	} else {
		space.AdminRoleID = types.StringValue(adminRoleID)
	}

	// Set state
	d.Append(state.Set(ctx, space)...)
}

// getAdminRoleID fetches all roles for the space and returns the ID of the admin role
func (e *spaceResource) getAdminRoleID(ctx context.Context, spaceID string) (string, error) {
	resp, err := e.client.GetAllRolesWithResponse(ctx, spaceID, nil)
	if err != nil {
		return "", fmt.Errorf("failed to fetch roles: %w", err)
	}

	if resp.StatusCode() != 200 {
		return "", fmt.Errorf("received unexpected status code: %d", resp.StatusCode())
	}

	if resp.JSON200 == nil || resp.JSON200.Items == nil {
		return "", fmt.Errorf("no roles found in response")
	}

	// Look for the admin role
	// The admin role typically has permissions for all operations and resources
	for _, role := range *resp.JSON200.Items {
		if e.isAdminRole(&role) {
			return role.Sys.Id, nil
		}
	}

	return "", fmt.Errorf("admin role not found")
}

// isAdminRole determines if a role is the admin role by examining its permissions
func (e *spaceResource) isAdminRole(role *sdk.Role) bool {
	// Admin roles typically have permissions for all operations
	// Check if this role has "all" permissions for key resources
	if role.Permissions == nil {
		return false
	}

	// Check for common admin permissions patterns
	hasAllPermissions := false
	permissionCount := 0

	for _, key := range role.Permissions.Keys() {
		if value, exists := role.Permissions.Get(key); exists {
			permissionCount++
			switch v := value.(type) {
			case string:
				if v == "all" {
					hasAllPermissions = true
				}
			case []interface{}:
				// Check if it contains "all" or has comprehensive permissions
				for _, item := range v {
					if str, ok := item.(string); ok && str == "all" {
						hasAllPermissions = true
						break
					}
				}
			}
		}
	}

	// Admin roles typically have many permissions, including "all" permissions
	// This is a heuristic - the exact logic may need adjustment based on actual admin role structure
	return hasAllPermissions && permissionCount > 3
}

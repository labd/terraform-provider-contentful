package role_assignment

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"

	"github.com/labd/terraform-provider-contentful/internal/sdk"
	"github.com/labd/terraform-provider-contentful/internal/utils"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &roleAssignmentResource{}
	_ resource.ResourceWithConfigure   = &roleAssignmentResource{}
	_ resource.ResourceWithImportState = &roleAssignmentResource{}
)

func NewRoleAssignmentResource() resource.Resource {
	return &roleAssignmentResource{}
}

// roleAssignmentResource is the resource implementation.
type roleAssignmentResource struct {
	client *sdk.ClientWithResponses
}

func (r *roleAssignmentResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_role_assignment"
}

func (r *roleAssignmentResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "A Contentful Role Assignment assigns roles to teams in a space.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Role Assignment ID",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"version": schema.Int64Attribute{
				Computed:    true,
				Description: "The current version of the role assignment",
			},
			"space_id": schema.StringAttribute{
				Required:    true,
				Description: "Space ID where the role assignment is made",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"team_id": schema.StringAttribute{
				Required:    true,
				Description: "Team ID to assign the role to (from contentful_team resource)",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"role_id": schema.StringAttribute{
				Optional:    true,
				Description: "Role ID to assign (from contentful_role resource). Mutually exclusive with is_admin.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"is_admin": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether to assign the admin role. Mutually exclusive with role_id.",
				Default:     booldefault.StaticBool(false),
			},
		},
	}
}

func (r *roleAssignmentResource) Configure(_ context.Context, request resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	data := request.ProviderData.(utils.ProviderData)
	r.client = data.Client
}

func (r *roleAssignmentResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan RoleAssignment

	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Validate mutually exclusive fields
	if err := r.validateMutuallyExclusive(&plan); err != nil {
		response.Diagnostics.AddError(
			"Invalid Configuration",
			err.Error(),
		)
		return
	}

	// TODO: Implement actual API call when endpoints become available
	response.Diagnostics.AddError(
		"Feature Not Available",
		"Role assignment functionality is not yet available. This resource is a placeholder for when Contentful adds role assignment APIs to their Management API.",
	)
}

func (r *roleAssignmentResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state RoleAssignment

	diags := request.State.Get(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	// TODO: Implement actual API call when endpoints become available
	response.Diagnostics.AddError(
		"Feature Not Available",
		"Role assignment functionality is not yet available. This resource is a placeholder for when Contentful adds role assignment APIs to their Management API.",
	)
}

func (r *roleAssignmentResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan RoleAssignment

	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Validate mutually exclusive fields
	if err := r.validateMutuallyExclusive(&plan); err != nil {
		response.Diagnostics.AddError(
			"Invalid Configuration",
			err.Error(),
		)
		return
	}

	// TODO: Implement actual API call when endpoints become available
	response.Diagnostics.AddError(
		"Feature Not Available",
		"Role assignment functionality is not yet available. This resource is a placeholder for when Contentful adds role assignment APIs to their Management API.",
	)
}

func (r *roleAssignmentResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state RoleAssignment
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	// TODO: Implement actual API call when endpoints become available
	// For now, just log that this would delete the role assignment
	fmt.Printf("Would delete role assignment: space=%s, team=%s, role=%s, admin=%t\n",
		state.SpaceID.ValueString(),
		state.TeamID.ValueString(),
		state.RoleID.ValueString(),
		state.IsAdmin.ValueBool(),
	)
}

func (r *roleAssignmentResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	// Expected import format: space_id:team_id:role_id or space_id:team_id:admin
	idParts := strings.Split(request.ID, ":")
	if len(idParts) != 3 || idParts[0] == "" || idParts[1] == "" || idParts[2] == "" {
		response.Diagnostics.AddError(
			"Error importing role assignment",
			fmt.Sprintf("Expected import format: space_id:team_id:role_id or space_id:team_id:admin, got: %s", request.ID),
		)
		return
	}

	// TODO: Implement actual import when API endpoints become available
	response.Diagnostics.AddError(
		"Feature Not Available",
		"Role assignment functionality is not yet available. This resource is a placeholder for when Contentful adds role assignment APIs to their Management API.",
	)
}

// validateMutuallyExclusive ensures that role_id and is_admin are mutually exclusive
func (r *roleAssignmentResource) validateMutuallyExclusive(assignment *RoleAssignment) error {
	hasRoleID := !assignment.RoleID.IsNull() && !assignment.RoleID.IsUnknown() && assignment.RoleID.ValueString() != ""
	isAdmin := assignment.IsAdmin.ValueBool()

	if hasRoleID && isAdmin {
		return fmt.Errorf("role_id and is_admin are mutually exclusive - specify either a specific role_id or set is_admin to true")
	}

	if !hasRoleID && !isAdmin {
		return fmt.Errorf("either role_id must be specified or is_admin must be set to true")
	}

	return nil
}
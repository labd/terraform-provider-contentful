package role_assignment

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// RoleAssignment is the main resource schema data
type RoleAssignment struct {
	ID      types.String `tfsdk:"id"`
	Version types.Int64  `tfsdk:"version"`
	SpaceID types.String `tfsdk:"space_id"`
	TeamID  types.String `tfsdk:"team_id"`
	RoleID  types.String `tfsdk:"role_id"`
	IsAdmin types.Bool   `tfsdk:"is_admin"`
}

// Import populates the RoleAssignment struct from an SDK object
// Note: This is a placeholder - actual API endpoints are needed
func (r *RoleAssignment) Import(assignment interface{}) {
	// TODO: Implement when API endpoints become available
	// This would populate from actual role assignment API response
}

// DraftForCreate creates a request object for creating a new role assignment
// Note: This is a placeholder - actual API endpoints are needed
func (r *RoleAssignment) DraftForCreate() interface{} {
	// TODO: Implement when API endpoints become available
	// This would create the appropriate request body for role assignment creation
	return map[string]interface{}{
		"spaceId": r.SpaceID.ValueString(),
		"teamId":  r.TeamID.ValueString(),
		"roleId":  r.RoleID.ValueString(),
		"isAdmin": r.IsAdmin.ValueBool(),
	}
}

// DraftForUpdate creates a request object for updating an existing role assignment
// Note: This is a placeholder - actual API endpoints are needed
func (r *RoleAssignment) DraftForUpdate() interface{} {
	// TODO: Implement when API endpoints become available
	// This would create the appropriate request body for role assignment updates
	return r.DraftForCreate()
}

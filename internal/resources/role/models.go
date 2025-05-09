package role

import (
	"encoding/json"
	// "fmt"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/labd/terraform-provider-contentful/internal/sdk"
	// "github.com/labd/terraform-provider-contentful/internal/utils"
)

// Role is the main resource schema data
type Role struct {
	ID      types.String `tfsdk:"id"`
	Version types.Int64  `tfsdk:"version"`

	Name        types.String          `tfsdk:"name"`
	Description types.String          `tfsdk:"description"`
	Permissions map[string]Permission `tfsdk:"permissions"`
	Policies    []Policy              `tfsdk:"policies"`
}

type Permission struct {
	All     types.String `tfsdk:"all"`
	Actions types.List   `tfsdk:"actions"`
}

type Policy struct {
	Effect     types.String `tfsdk:"effect:"`
	Actions    types.List   `tfsdk:"actions"`
	Constraint Constraint   `tfsdk:"constraint"`
}

type Constraint struct {
	And []types.List `tfsdk:"and"`
}

// Import populates the Role struct from an sdk.Role object
func (r *Role) Import(role *sdk.Role) {
	r.ID = types.StringValue(*role.Sys.Id)
	r.Version = types.Int64Value(int64(*role.Sys.Version))
	r.Name = types.StringValue(role.Name)
	r.Description = types.StringValue(role.Description)
}

func (r *Role) DraftForCreate() sdk.RoleCreate {
	return sdk.RoleCreate{
		Name:        r.Name.ValueString(),
		Description: r.Description.ValueString(),
		Permissions: convertPermissions(r.Permissions),
	}
}

func (r *Role) DraftForUpdate() sdk.RoleUpdate {
	return sdk.RoleUpdate{
		Name:        r.Name.ValueString(),
		Description: r.Description.ValueString(),
		Permissions: convertPermissions(r.Permissions),
	}
}

// NOTE: Validate this function
func convertPermissions(input map[string]Permission) sdk.RolePermissions {
	rawMap := make(map[string]json.RawMessage)

	for key, val := range input {
		// Handle the "all" shortcut
		if !val.All.IsNull() && val.All.ValueString() != "" {
			b, err := json.Marshal(val.All.ValueString())
			if err != nil {
				continue // or log
			}
			rawMap[key] = b
			continue
		}

		// Handle the explicit actions list
		if !val.Actions.IsNull() && val.Actions.Elements() != nil {
			var actions []string
			for _, elem := range val.Actions.Elements() {
				strVal, ok := elem.(types.String)
				if !ok || strVal.IsNull() {
					continue
				}
				actions = append(actions, strVal.ValueString())
			}

			if len(actions) > 0 {
				b, err := json.Marshal(actions)
				if err != nil {
					continue // or log
				}
				rawMap[key] = b
			}
		}
	}

	// Marshal and unmarshal into the SDK's RolePermissions type
	b, err := json.Marshal(rawMap)
	if err != nil {
		return nil
	}

	var perms sdk.RolePermissions
	if err := json.Unmarshal(b, &perms); err != nil {
		return nil
	}

	return perms
}

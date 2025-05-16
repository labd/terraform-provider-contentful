package role

import (
	// "encoding/json"

	// "fmt"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/iancoleman/orderedmap"
	"github.com/labd/terraform-provider-contentful/internal/sdk"
)

// Role is the main resource schema data
type Role struct {
	ID      types.String `tfsdk:"id"`
	Version types.Int64  `tfsdk:"version"`
	SpaceID types.String `tfsdk:"space_id"`

	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Permission  []Permission `tfsdk:"permission"`
	Policy      []Policy     `tfsdk:"policy"`
}

type Permission struct {
	ID     types.String   `tfsdk:"id"`
	Value  types.String   `tfsdk:"value"`
	Values []types.String `tfsdk:"values"`
}

type Policy struct {
	Effect     types.String `tfsdk:"effect"`
	Actions    Action       `tfsdk:"actions"`
	Constraint types.String `tfsdk:"constraint"`
}

type Action struct {
	Value  types.String   `tfsdk:"value"`
	Values []types.String `tfsdk:"values"`
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
		Permissions: convertPermissions(r.Permission),
		Policies:    convertPolicies(r.Policy),
	}
}

func (r *Role) DraftForUpdate() sdk.RoleUpdate {
	return sdk.RoleUpdate{
		Name:        r.Name.ValueString(),
		Description: r.Description.ValueString(),
		Permissions: convertPermissions(r.Permission),
		Policies:    convertPolicies(r.Policy),
	}
}

func convertPermissions(p []Permission) *orderedmap.OrderedMap {
	permissions := orderedmap.New()

	for _, permission := range p {
		id := permission.ID.ValueString()
		if v := permission.Value.ValueString(); v != "" {
			permissions.Set(id, permission.Value.String())
		} else if len(permission.Values) > 0 {
			strVals := make([]string, len(permission.Values))
			for i, val := range permission.Values {
				strVals[i] = val.ValueString() // extract the raw string
			}
			permissions.Set(id, strVals)
		}
	}

	return permissions
}

func convertPolicies(policies []Policy) *[]any {
	var out []any
	for _, policy := range policies {
		policyMap := map[string]interface{}{
			"effect": policy.Effect.ValueString(),
		}

		// Handle action as a nested map
		if v := policy.Actions.Value.ValueString(); v != "" {
			policyMap["actions"] = v
		}
		if len(policy.Actions.Values) > 0 {
			strVals := make([]string, len(policy.Actions.Values))
			for i, val := range policy.Actions.Values {
				strVals[i] = val.ValueString()
			}
			policyMap["actions"] = strVals
		}

		if c := policy.Constraint.ValueString(); c != "" {
			policyMap["constraint"] = c
		}
		out = append(out, policyMap)
	}
	return &out
}

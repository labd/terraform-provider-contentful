package role

import (
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/iancoleman/orderedmap"
	"github.com/labd/terraform-provider-contentful/internal/sdk"
	"github.com/labd/terraform-provider-contentful/internal/utils"
)

// Role is the main resource schema data
type Role struct {
	ID      types.String `tfsdk:"id"`
	RoleId  types.String `tfsdk:"role_id"`
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
func (r *Role) Import(role *sdk.Role) error {
	r.ID = types.StringValue(role.Sys.Id)
	r.RoleId = types.StringValue(role.Sys.Id)
	r.SpaceID = types.StringValue(role.Sys.Space.Sys.Id)
	r.Version = types.Int64Value(int64(role.Sys.Version))
	r.Name = types.StringValue(role.Name)
	r.Description = types.StringValue(role.Description)

	err := r.BuildPermissionsFromAPIResponse(role)
	if err != nil {
		return err
	}
	r.BuildPoliciesFromAPIResponse(role)

	return nil
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
			permissions.Set(id, v)
		} else if len(permission.Values) > 0 {
			strVals := make([]string, len(permission.Values))
			for i, val := range permission.Values {
				strVals[i] = val.ValueString()
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
			policyMap["constraint"] = ParseContentValue(c)
		}

		out = append(out, policyMap)
	}

	return &out
}

func (r *Role) BuildPermissionsFromAPIResponse(role *sdk.Role) error {
	var permissions []Permission

	// If no fields are present in the response, return early
	if len(role.Permissions.Keys()) == 0 {
		return nil
	}

	for _, key := range role.Permissions.Keys() {
		rawActions, _ := role.Permissions.Get(key)
		permission := Permission{}

		var actions []string

		switch val := rawActions.(type) {
		case string:
			actions = []string{val}
		case []string:
			actions = val
		case []interface{}:
			for _, item := range val {
				if str, ok := item.(string); ok {
					actions = append(actions, str)
				} else {
					return fmt.Errorf("unexpected type in permission actions slice: %T", item)
				}
			}
		default:
			return fmt.Errorf("unexpected type for permission actions: %T", val)
		}

		permission.ID = types.StringValue(key)

		// If the permission is ["all"], set Value, not Values
		if len(actions) == 1 && actions[0] == "all" {
			permission.Value = types.StringValue("all")
		} else {
			var values []types.String
			if len(actions) == 0 {
				continue
			}

			for _, action := range actions {
				values = append(values, types.StringValue(action))
			}
			permission.Values = values
		}

		permissions = append(permissions, permission)
	}

	r.Permission = permissions

	return nil
}

func (r *Role) BuildPoliciesFromAPIResponse(role *sdk.Role) {
	var policies []Policy

	if role.Policies == nil || len(*role.Policies) == 0 {
		return
	}

	for _, raw := range *role.Policies {
		policyMap, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}

		policy := Policy{}

		// Handle "effect"
		if effect, ok := policyMap["effect"].(string); ok {
			policy.Effect = types.StringValue(effect)
		}

		// Handle "actions" as string or []interface{}
		if rawActions, ok := policyMap["actions"]; ok {
			switch actions := rawActions.(type) {
			case string:
				policy.Actions.Value = types.StringValue(actions)
			case []interface{}:
				var actionValues []types.String
				for _, item := range actions {
					if str, ok := item.(string); ok {
						actionValues = append(actionValues, types.StringValue(str))
					}
				}
				policy.Actions.Values = actionValues
			}
		}

		// Handle "constraint" as optional JSON string
		if constraintRaw, ok := policyMap["constraint"]; ok {
			if marshaled, err := json.Marshal(constraintRaw); err == nil {
				policy.Constraint = types.StringValue(string(marshaled))
			}
		}

		policies = append(policies, policy)
	}

	r.Policy = policies
}

// ParseContentValue tries to parse a string as JSON, otherwise returns the original value
func ParseContentValue(value string) interface{} {
	var content any

	err := json.Unmarshal([]byte(value), &content)
	if err != nil {
		return value
	}

	return utils.SortOrderedMapRecursively(content)
}

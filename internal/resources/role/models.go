package role

import (
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/types"
	// "github.com/iancoleman/orderedmap"

	"github.com/labd/terraform-provider-contentful/internal/sdk"
	"github.com/labd/terraform-provider-contentful/internal/utils"
)

// Role is the main resource schema data
type Role struct {
	ID      types.String `tfsdk:"id"`
	Version types.Int64  `tfsdk:"version"`

	Name        types.String          `tfsdk:"name"`
	Description types.String          `tfsdk:"description"`
	Permissions map[string]Permission `tfsdk:"permission"`
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

	r.BuildFieldsFromAPIResponse(role)
}

func (r *Role) DraftForCreate() sdk.RoleCreate {
	return sdk.RoleCreate{
		Name:        r.Name.ValueString(),
		Description: r.Description.ValueString(),
		Permissions: r.Permissions,
	}
}

func (r *Role) DraftForUpdate() sdk.RoleUpdate {
	return sdk.RoleUpdate{
		Name:        r.Name.ValueString(),
		Description: r.Description.ValueString(),
		Permissions: r.Permissions,
	}
}

// func (r *Role) DraftForCreate() sdk.RoleCreate {
// 	fieldProperties := orderedmap.New()
//
// 	// for _, field := range r.Field {
// 	// 	fieldID := field.ID.ValueString()
// 	// 	locale := field.Locale.ValueString()
// 	// 	content := ParseContentValue(field.Content.ValueString())
// 	//
// 	// 	prop, ok := fieldProperties.Get(fieldID)
// 	// 	if !ok {
// 	// 		prop = map[string]any{}
// 	// 		fieldProperties.Set(fieldID, prop)
// 	// 	}
// 	//
// 	// 	prop.(map[string]any)[locale] = content
// 	// }
//
// 	return sdk.RoleCreate{
// 		Fields: fieldProperties,
// 	}
// }

// parseContentValue tries to parse a string as JSON, otherwise returns the original value
func ParseContentValue(value string) interface{} {
	var content any
	err := json.Unmarshal([]byte(value), &content)
	if err != nil {
		content = value
	}

	return utils.SortOrderedMapRecursively(content)
}

// BuildFieldsFromAPIResponse builds the Field array from API response
func (r *Role) BuildFieldsFromAPIResponse(role *sdk.Role) {
	r.Field = []Field{}

	// If no fields are present in the response, return early
	if len(entry.Fields.Keys()) == 0 {
		return
	}

	// Convert fields from the API response to the Field structure
	fields := entry.Fields
	for _, fieldID := range fields.Keys() {
		fieldValue, _ := fields.Get(fieldID)

		subFields := fieldValue.(orderedmap.OrderedMap)

		for _, locale := range subFields.Keys() {
			content, _ := subFields.Get(locale)

			// Convert the content back to string representation for storage
			contentStr := ""
			switch v := content.(type) {
			case string:
				contentStr = v
			default:
				v = utils.SortOrderedMapRecursively(v)
				// Try to marshal complex types back to JSON string
				if jsonBytes, err := json.Marshal(v); err == nil {
					contentStr = string(jsonBytes)
				} else {
					contentStr = fmt.Sprintf("%v", v)
				}
			}

			e.Field = append(e.Field, Field{
				ID:      types.StringValue(fieldID),
				Locale:  types.StringValue(locale),
				Content: types.StringValue(contentStr),
			})
		}
	}
}

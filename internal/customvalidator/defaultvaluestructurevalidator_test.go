package customvalidator

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func TestDefaultValueStructureValidator_Creation(t *testing.T) {
	// Basic test to ensure the validator can be created
	v := DefaultValueStructure()
	if v == nil {
		t.Error("DefaultValueStructure() should not return nil")
	}
}

// defaultValueAttrTypes returns the attribute types for the default_value object,
// matching the schema definition in contenttype/resource.go.
func defaultValueAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"string": types.MapType{ElemType: types.StringType},
		"bool":   types.MapType{ElemType: types.BoolType},
		"array":  types.MapType{ElemType: types.ListType{ElemType: types.StringType}},
	}
}

func runValidator(t *testing.T, configValue basetypes.ObjectValue) diag.Diagnostics {
	t.Helper()
	ctx := context.Background()
	v := DefaultValueStructureValidator{}
	request := validator.ObjectRequest{
		Path:        path.Root("default_value"),
		ConfigValue: configValue,
	}
	response := &validator.ObjectResponse{}
	v.ValidateObject(ctx, request, response)
	return response.Diagnostics
}

func TestValidateObject_NullObject(t *testing.T) {
	configValue := types.ObjectNull(defaultValueAttrTypes())
	diags := runValidator(t, configValue)
	if diags.HasError() {
		t.Errorf("expected no errors for null object, got: %v", diags.Errors())
	}
}

func TestValidateObject_UnknownObject(t *testing.T) {
	configValue := types.ObjectUnknown(defaultValueAttrTypes())
	diags := runValidator(t, configValue)
	if diags.HasError() {
		t.Errorf("expected no errors for unknown object, got: %v", diags.Errors())
	}
}

func TestValidateObject_ValidStringMap(t *testing.T) {
	configValue := types.ObjectValueMust(defaultValueAttrTypes(), map[string]attr.Value{
		"string": types.MapValueMust(types.StringType, map[string]attr.Value{
			"en-US": types.StringValue("green"),
		}),
		"bool":  types.MapNull(types.BoolType),
		"array": types.MapNull(types.ListType{ElemType: types.StringType}),
	})
	diags := runValidator(t, configValue)
	if diags.HasError() {
		t.Errorf("expected no errors for valid string map, got: %v", diags.Errors())
	}
}

func TestValidateObject_ValidBoolMap(t *testing.T) {
	configValue := types.ObjectValueMust(defaultValueAttrTypes(), map[string]attr.Value{
		"string": types.MapNull(types.StringType),
		"bool": types.MapValueMust(types.BoolType, map[string]attr.Value{
			"en-US": types.BoolValue(true),
		}),
		"array": types.MapNull(types.ListType{ElemType: types.StringType}),
	})
	diags := runValidator(t, configValue)
	if diags.HasError() {
		t.Errorf("expected no errors for valid bool map, got: %v", diags.Errors())
	}
}

func TestValidateObject_EmptyStringMap_Error(t *testing.T) {
	configValue := types.ObjectValueMust(defaultValueAttrTypes(), map[string]attr.Value{
		"string": types.MapValueMust(types.StringType, map[string]attr.Value{}),
		"bool":   types.MapNull(types.BoolType),
		"array":  types.MapNull(types.ListType{ElemType: types.StringType}),
	})
	diags := runValidator(t, configValue)
	if !diags.HasError() {
		t.Error("expected error for empty string map, got none")
	}
	found := false
	for _, d := range diags.Errors() {
		if d.Summary() == "Empty default_value" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 'Empty default_value' error, got: %v", diags.Errors())
	}
}

func TestValidateObject_AllNullMaps_Error(t *testing.T) {
	configValue := types.ObjectValueMust(defaultValueAttrTypes(), map[string]attr.Value{
		"string": types.MapNull(types.StringType),
		"bool":   types.MapNull(types.BoolType),
		"array":  types.MapNull(types.ListType{ElemType: types.StringType}),
	})
	diags := runValidator(t, configValue)
	if !diags.HasError() {
		t.Error("expected error when all maps are null, got none")
	}
}

func TestValidateObject_UnknownStringMap_NoError(t *testing.T) {
	// Simulates a variable used as a map key: default_value = { string = { (var.locale) = "value" } }
	// The framework marks the entire map as unknown when it cannot resolve the key at validation time.
	configValue := types.ObjectValueMust(defaultValueAttrTypes(), map[string]attr.Value{
		"string": types.MapUnknown(types.StringType),
		"bool":   types.MapNull(types.BoolType),
		"array":  types.MapNull(types.ListType{ElemType: types.StringType}),
	})
	diags := runValidator(t, configValue)
	if diags.HasError() {
		t.Errorf("expected no errors for unknown string map (variable as key), got: %v", diags.Errors())
	}
}

func TestValidateObject_UnknownBoolMap_NoError(t *testing.T) {
	configValue := types.ObjectValueMust(defaultValueAttrTypes(), map[string]attr.Value{
		"string": types.MapNull(types.StringType),
		"bool":   types.MapUnknown(types.BoolType),
		"array":  types.MapNull(types.ListType{ElemType: types.StringType}),
	})
	diags := runValidator(t, configValue)
	if diags.HasError() {
		t.Errorf("expected no errors for unknown bool map, got: %v", diags.Errors())
	}
}

func TestValidateObject_UnknownArrayMap_NoError(t *testing.T) {
	configValue := types.ObjectValueMust(defaultValueAttrTypes(), map[string]attr.Value{
		"string": types.MapNull(types.StringType),
		"bool":   types.MapNull(types.BoolType),
		"array":  types.MapUnknown(types.ListType{ElemType: types.StringType}),
	})
	diags := runValidator(t, configValue)
	if diags.HasError() {
		t.Errorf("expected no errors for unknown array map, got: %v", diags.Errors())
	}
}

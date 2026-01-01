package contenttype

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestDefaultValue_Draft_WithNullMaps(t *testing.T) {
	// Test case where DefaultValue with null maps returns nil from Draft()
	defaultValue := &DefaultValue{
		Bool:   types.MapNull(types.BoolType),
		String: types.MapNull(types.StringType),
	}

	result := defaultValue.Draft()

	// This should return nil
	assert.Nil(t, result)
}

func TestDefaultValue_Draft_WithEmptyMaps(t *testing.T) {
	// Test case with empty but non-null maps
	defaultValue := &DefaultValue{
		Bool:   types.MapValueMust(types.BoolType, map[string]attr.Value{}),
		String: types.MapValueMust(types.StringType, map[string]attr.Value{}),
	}

	result := defaultValue.Draft()

	// This should also return nil since both maps are empty
	assert.Nil(t, result)
}

func TestDefaultValue_Draft_WithStringValues(t *testing.T) {
	// Test case with actual string values
	defaultValue := &DefaultValue{
		Bool: types.MapNull(types.BoolType),
		String: types.MapValueMust(types.StringType, map[string]attr.Value{
			"en-US": types.StringValue("green"),
			"de-DE": types.StringValue("grün"),
		}),
	}

	result := defaultValue.Draft()

	// This should return a valid map
	assert.NotNil(t, result)
	assert.Equal(t, "green", (*result)["en-US"])
	assert.Equal(t, "grün", (*result)["de-DE"])
}

func TestDefaultValue_Draft_WithBoolValues(t *testing.T) {
	// Test case with boolean values
	defaultValue := &DefaultValue{
		Bool: types.MapValueMust(types.BoolType, map[string]attr.Value{
			"en-US": types.BoolValue(true),
			"de-DE": types.BoolValue(false),
		}),
		String: types.MapNull(types.StringType),
	}

	result := defaultValue.Draft()

	// This should return a valid map
	assert.NotNil(t, result)
	assert.Equal(t, true, (*result)["en-US"])
	assert.Equal(t, false, (*result)["de-DE"])
}

func TestDefaultValue_HasContent_WithNullMaps(t *testing.T) {
	defaultValue := &DefaultValue{
		Bool:   types.MapNull(types.BoolType),
		String: types.MapNull(types.StringType),
	}

	result := defaultValue.HasContent()
	assert.False(t, result)
}

func TestDefaultValue_HasContent_WithEmptyMaps(t *testing.T) {
	defaultValue := &DefaultValue{
		Bool:   types.MapValueMust(types.BoolType, map[string]attr.Value{}),
		String: types.MapValueMust(types.StringType, map[string]attr.Value{}),
	}

	result := defaultValue.HasContent()
	assert.False(t, result)
}

func TestDefaultValue_HasContent_WithStringValues(t *testing.T) {
	defaultValue := &DefaultValue{
		Bool: types.MapNull(types.BoolType),
		String: types.MapValueMust(types.StringType, map[string]attr.Value{
			"en-US": types.StringValue("test"),
		}),
	}

	result := defaultValue.HasContent()
	assert.True(t, result)
}

func TestDefaultValue_HasContent_WithBoolValues(t *testing.T) {
	defaultValue := &DefaultValue{
		Bool: types.MapValueMust(types.BoolType, map[string]attr.Value{
			"en-US": types.BoolValue(true),
		}),
		String: types.MapNull(types.StringType),
	}

	result := defaultValue.HasContent()
	assert.True(t, result)
}

func TestDefaultValue_HasContent_NilPointer(t *testing.T) {
	var defaultValue *DefaultValue = nil

	result := defaultValue.HasContent()
	assert.False(t, result)
}

func TestField_ToNative_WithStringDefaultValue(t *testing.T) {
	// Test ToNative with string default value (your original use case)
	field := &Field{
		Id:   types.StringValue("themeColor"),
		Name: types.StringValue("Theme Color"),
		Type: types.StringValue("Symbol"),
		DefaultValue: &DefaultValue{
			Bool: types.MapNull(types.BoolType),
			String: types.MapValueMust(types.StringType, map[string]attr.Value{
				"en-US": types.StringValue("green"),
			}),
		},
		Required:    types.BoolValue(false),
		Localized:   types.BoolValue(false),
		Disabled:    types.BoolValue(false),
		Omitted:     types.BoolValue(false),
		Validations: []Validation{},
	}

	result, err := field.ToNative()

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotNil(t, result.DefaultValue)
	assert.Equal(t, "green", (*result.DefaultValue)["en-US"])
}

func TestField_ToNative_WithBoolDefaultValue(t *testing.T) {
	// Test ToNative with boolean default value
	field := &Field{
		Id:   types.StringValue("enableFeature"),
		Name: types.StringValue("Enable Feature"),
		Type: types.StringValue("Boolean"),
		DefaultValue: &DefaultValue{
			Bool: types.MapValueMust(types.BoolType, map[string]attr.Value{
				"en-US": types.BoolValue(true),
			}),
			String: types.MapNull(types.StringType),
		},
		Required:    types.BoolValue(false),
		Localized:   types.BoolValue(false),
		Disabled:    types.BoolValue(false),
		Omitted:     types.BoolValue(false),
		Validations: []Validation{},
	}

	result, err := field.ToNative()

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotNil(t, result.DefaultValue)
	assert.Equal(t, true, (*result.DefaultValue)["en-US"])
}

func TestField_ToNative_WithNullDefaultValue(t *testing.T) {
	// Test ToNative with no default value
	field := &Field{
		Id:           types.StringValue("noDefault"),
		Name:         types.StringValue("No Default"),
		Type:         types.StringValue("Symbol"),
		DefaultValue: nil,
		Required:     types.BoolValue(false),
		Localized:    types.BoolValue(false),
		Disabled:     types.BoolValue(false),
		Omitted:      types.BoolValue(false),
		Validations:  []Validation{},
	}

	result, err := field.ToNative()

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Nil(t, result.DefaultValue)
}

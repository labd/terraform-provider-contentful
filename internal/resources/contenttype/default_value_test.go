package contenttype

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestDefaultValue_Draft_WithNullMaps(t *testing.T) {
	// Test case to reproduce the issue where DefaultValue with null maps
	// returns nil from Draft(), causing plan discrepancies
	defaultValue := &DefaultValue{
		Bool:   types.MapNull(types.BoolType),
		String: types.MapNull(types.StringType),
	}

	result := defaultValue.Draft()
	
	// This should return nil, which is the current behavior causing the issue
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

func TestDefaultValue_Draft_WithValues(t *testing.T) {
	// Test case with actual values
	defaultValue := &DefaultValue{
		Bool:   types.MapNull(types.BoolType),
		String: types.MapValueMust(types.StringType, map[string]attr.Value{
			"en": types.StringValue("test"),
		}),
	}

	result := defaultValue.Draft()
	
	// This should return a valid map
	assert.NotNil(t, result)
	assert.Equal(t, "test", (*result)["en"])
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

func TestDefaultValue_HasContent_WithValues(t *testing.T) {
	defaultValue := &DefaultValue{
		Bool:   types.MapNull(types.BoolType),
		String: types.MapValueMust(types.StringType, map[string]attr.Value{
			"en": types.StringValue("test"),
		}),
	}

	result := defaultValue.HasContent()
	assert.True(t, result)
}

func TestDefaultValue_HasContent_NilPointer(t *testing.T) {
	var defaultValue *DefaultValue = nil

	result := defaultValue.HasContent()
	assert.False(t, result)
}

func TestField_ToNative_ArrayFieldWithDefaultValue(t *testing.T) {
	// Test case to reproduce the issue with Array fields and default values
	field := &Field{
		Id:   types.StringValue("test_array"),
		Name: types.StringValue("Test Array"),
		Type: types.StringValue("Array"),
		Items: &Items{
			Type: types.StringValue("Symbol"),
		},
		DefaultValue: &DefaultValue{
			Bool:   types.MapNull(types.BoolType),
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
	
	// The key issue: DefaultValue should not be set for Array fields when it's empty
	// Currently this would be nil due to Draft() returning nil, causing plan discrepancy
	// After fix, it should be nil consistently
	assert.Nil(t, result.DefaultValue)
}
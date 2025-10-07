package contenttype

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/labd/terraform-provider-contentful/internal/sdk"
	"github.com/labd/terraform-provider-contentful/internal/utils"
	"github.com/stretchr/testify/assert"
)

func TestValidationDraftReturnsErrorForUnsupportedValidation(t *testing.T) {
	validation := Validation{}

	_, err := validation.Draft()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "at least one validation property must be set")
}

func TestValidationDraftReturnsCorrectUniqueValidation(t *testing.T) {
	validation := Validation{
		Unique:  types.BoolValue(true),
		Message: types.StringValue("Unique validation message"),
	}

	result, err := validation.Draft()

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, true, *result.Unique)
	assert.Equal(t, "Unique validation message", *result.Message)
}

func TestValidationDraftReturnsCorrectEnabledNodeTypesValidation(t *testing.T) {
	validation := Validation{
		EnabledNodeTypes: []types.String{},
		Message:          types.StringValue("Unique validation message"),
	}

	result, err := validation.Draft()

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, []string{}, *result.EnabledNodeTypes)
	assert.Equal(t, "Unique validation message", *result.Message)
}

func TestDefaultValue_HasContent(t *testing.T) {
	tests := []struct {
		name     string
		input    *DefaultValue
		expected bool
	}{
		{
			name: "StringContent",
			input: &DefaultValue{
				String: types.MapValueMust(types.StringType, map[string]attr.Value{
					"foo": types.StringValue("bar"),
				}),
				Bool:  types.MapNull(types.BoolType),
				Array: nil,
			},
			expected: true,
		},
		{
			name: "BoolContent",
			input: &DefaultValue{
				String: types.MapNull(types.StringType),
				Bool: types.MapValueMust(types.BoolType, map[string]attr.Value{
					"flag": types.BoolValue(true),
				}),
				Array: nil,
			},
			expected: true,
		},
		{
			name: "ArrayContent",
			input: &DefaultValue{
				String: types.MapNull(types.StringType),
				Bool:   types.MapNull(types.BoolType),
				Array: map[string]types.List{
					"arr": types.ListValueMust(types.StringType, []attr.Value{types.StringValue("val")}),
				},
			},
			expected: true,
		},
		{
			name: "Empty",
			input: &DefaultValue{
				String: types.MapNull(types.StringType),
				Bool:   types.MapNull(types.BoolType),
				Array:  nil,
			},
			expected: false,
		},
		{
			name:     "NilReceiver",
			input:    nil,
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.input.HasContent()
			if result != tc.expected {
				t.Errorf("expected %v, got %v", tc.expected, result)
			}
		})
	}
}

func TestDefaultValue_Draft(t *testing.T) {
	tests := []struct {
		name     string
		input    *DefaultValue
		expected map[string]any
	}{
		{
			name: "StringContent",
			input: &DefaultValue{
				String: types.MapValueMust(types.StringType, map[string]attr.Value{
					"foo": types.StringValue("bar"),
				}),
				Bool:  types.MapNull(types.BoolType),
				Array: nil,
			},
			expected: map[string]any{"foo": "bar"},
		},
		{
			name: "BoolContent",
			input: &DefaultValue{
				String: types.MapNull(types.StringType),
				Bool: types.MapValueMust(types.BoolType, map[string]attr.Value{
					"flag": types.BoolValue(true),
				}),
				Array: nil,
			},
			expected: map[string]any{"flag": true},
		},
		{
			name: "ArrayContent",
			input: &DefaultValue{
				String: types.MapNull(types.StringType),
				Bool:   types.MapNull(types.BoolType),
				Array: map[string]types.List{
					"arr": types.ListValueMust(types.StringType, []attr.Value{types.StringValue("val")}),
				},
			},
			expected: map[string]any{"arr": []string{"val"}},
		},
		{
			name: "MultipleContent",
			input: &DefaultValue{
				String: types.MapValueMust(types.StringType, map[string]attr.Value{
					"foo": types.StringValue("bar"),
				}),
				Bool: types.MapValueMust(types.BoolType, map[string]attr.Value{
					"flag": types.BoolValue(false),
				}),
				Array: map[string]types.List{
					"arr": types.ListValueMust(types.StringType, []attr.Value{types.StringValue("val")}),
				},
			},
			expected: map[string]any{
				"foo":  "bar",
				"flag": false,
				"arr":  []string{"val"},
			},
		},
		{
			name: "Empty",
			input: &DefaultValue{
				String: types.MapNull(types.StringType),
				Bool:   types.MapNull(types.BoolType),
				Array:  nil,
			},
			expected: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.input.Draft()
			if tc.expected == nil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				assert.Equal(t, tc.expected, *result)
			}
		})
	}
}

func TestGetTypeOfMap(t *testing.T) {
	tests := []struct {
		name     string
		input    *map[string]any
		expected *string
		wantErr  bool
	}{
		{
			name:     "NilMap",
			input:    nil,
			expected: nil,
			wantErr:  false,
		},
		{
			name:     "StringType",
			input:    &map[string]any{"foo": "bar"},
			expected: ptr("string"),
			wantErr:  false,
		},
		{
			name:     "BoolType",
			input:    &map[string]any{"flag": true},
			expected: ptr("bool"),
			wantErr:  false,
		},
		{
			name:     "Float64Type",
			input:    &map[string]any{"num": float64(42)},
			expected: ptr("float64"),
			wantErr:  false,
		},
		{
			name:     "SliceType",
			input:    &map[string]any{"arr": []interface{}{"a", "b"}},
			expected: ptr("[]interface{}"),
			wantErr:  false,
		},
		{
			name:     "UnsupportedType",
			input:    &map[string]any{"foo": 123}, // int, not supported
			expected: nil,
			wantErr:  true,
		},
		{
			name:     "EmptyMap",
			input:    &map[string]any{},
			expected: nil,
			wantErr:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			typ, err := getTypeOfMap(tc.input)
			if tc.wantErr {
				assert.Error(t, err)
				assert.Nil(t, typ)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, typ)
			}
		})
	}
}

func ptr[T any](s T) *T {
	return &s
}

func TestField_Import(t *testing.T) {
	tests := []struct {
		name         string
		sdkField     sdk.Field
		defaultValue *map[string]any
		wantErr      bool
		check        func(t *testing.T, f *Field)
	}{
		{
			name: "StringDefaultValue",
			sdkField: sdk.Field{
				Id:        "id1",
				Name:      "name1",
				Type:      sdk.FieldType("Text"),
				Required:  true,
				Omitted:   utils.Pointer(false),
				Localized: true,
				Disabled:  utils.Pointer(false),
				DefaultValue: &map[string]any{
					"foo": "bar",
				},
			},
			wantErr: false,
			check: func(t *testing.T, f *Field) {
				assert.Equal(t, "id1", f.Id.ValueString())
				assert.Equal(t, "bar", f.DefaultValue.String.Elements()["foo"].(types.String).ValueString())
			},
		},
		{
			name: "BoolDefaultValue",
			sdkField: sdk.Field{
				Id:        "id2",
				Name:      "name2",
				Type:      sdk.FieldType("Boolean"),
				Required:  false,
				Omitted:   utils.Pointer(false),
				Localized: false,
				Disabled:  utils.Pointer(false),
				DefaultValue: &map[string]any{
					"flag": true,
				},
			},
			wantErr: false,
			check: func(t *testing.T, f *Field) {
				assert.Equal(t, true, f.DefaultValue.Bool.Elements()["flag"].(types.Bool).ValueBool())
			},
		},
		{
			name: "UnsupportedDefaultValueType",
			sdkField: sdk.Field{
				Id:        "id4",
				Name:      "name4",
				Type:      sdk.FieldType("Text"),
				Required:  false,
				Omitted:   utils.Pointer(false),
				Localized: false,
				Disabled:  utils.Pointer(false),
				DefaultValue: &map[string]any{
					"foo": 123, // int is unsupported
				},
			},
			wantErr: true,
			check:   func(t *testing.T, f *Field) {},
		},
		{
			name: "NilDefaultValue",
			sdkField: sdk.Field{
				Id:           "id5",
				Name:         "name5",
				Type:         sdk.FieldType("Text"),
				Required:     false,
				Omitted:      utils.Pointer(false),
				Localized:    false,
				Disabled:     utils.Pointer(false),
				DefaultValue: nil,
			},
			wantErr: false,
			check: func(t *testing.T, f *Field) {
				assert.Nil(t, f.DefaultValue)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var f Field
			err := f.Import(tc.sdkField)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				tc.check(t, &f)
			}
		})
	}
}

package contenttype

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestValidationDraftReturnsErrorForUnsupportedValidation(t *testing.T) {
	validation := Validation{}

	_, err := validation.Draft()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported validation used")
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

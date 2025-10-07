package customvalidator

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// DefaultValueStructureValidator checks that default_value has the correct nested structure
type DefaultValueStructureValidator struct{}

// Description returns a description of the validator.
func (v DefaultValueStructureValidator) Description(_ context.Context) string {
	return "Validates that default_value has the correct structure with 'string' or 'bool' keys"
}

// MarkdownDescription returns a markdown description of the validator.
func (v DefaultValueStructureValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

// ValidateObject implements validator.Object.
func (v DefaultValueStructureValidator) ValidateObject(ctx context.Context, request validator.ObjectRequest, response *validator.ObjectResponse) {
	if request.ConfigValue.IsNull() || request.ConfigValue.IsUnknown() {
		return
	}

	attributes := request.ConfigValue.Attributes()

	// Check if both string and bool are null/empty
	stringAttr, hasString := attributes["string"]
	boolAttr, hasBool := attributes["bool"]
	arrayAttr, hasArray := attributes["array"]

	// If we don't have the expected string/bool attributes, check for wrong syntax
	if !hasString && !hasBool && !hasArray {
		// User likely used flat map syntax like { "en-US" = "green" }
		if len(attributes) > 0 {
			// Get first attribute as example
			var firstKey string
			for key := range attributes {
				firstKey = key
				break
			}
			response.Diagnostics.AddAttributeError(
				request.Path,
				"Invalid default_value structure",
				fmt.Sprintf("Found unexpected attribute '%s' in default_value. "+
					"The correct syntax is: default_value = { string = { \"%s\" = \"value\" } } "+
					"or default_value = { bool = { \"%s\" = true } }", firstKey, firstKey, firstKey),
			)
			return
		}
		response.Diagnostics.AddAttributeError(
			request.Path,
			"Invalid default_value structure",
			"default_value must contain either 'string' or 'bool' attribute. "+
				"Example: default_value = { string = { \"en-US\" = \"green\" } }",
		)
		return
	}

	// Check if any content exists
	stringHasContent := false
	boolHasContent := false
	arrayHasContent := false

	if hasString && !stringAttr.IsNull() && !stringAttr.IsUnknown() {
		if stringMap, ok := stringAttr.(types.Map); ok && len(stringMap.Elements()) > 0 {
			stringHasContent = true
		}
	}

	if hasBool && !boolAttr.IsNull() && !boolAttr.IsUnknown() {
		if boolMap, ok := boolAttr.(types.Map); ok && len(boolMap.Elements()) > 0 {
			boolHasContent = true
		}
	}

	if hasArray && !arrayAttr.IsNull() && !arrayAttr.IsUnknown() {
		if arrayMap, ok := arrayAttr.(types.Map); ok && len(arrayMap.Elements()) > 0 {
			arrayHasContent = true
		}
	}

	// If we have string/bool attributes but they're both empty, that's an error
	if !stringHasContent && !boolHasContent && !arrayHasContent && (hasString || hasBool || hasArray) {
		response.Diagnostics.AddAttributeError(
			request.Path,
			"Empty default_value",
			"default_value must contain actual values. "+
				"Example: default_value = { string = { \"en-US\" = \"green\" } }",
		)
	}
}

// DefaultValueStructure returns a validator that ensures default_value has correct structure
func DefaultValueStructure() validator.Object {
	return DefaultValueStructureValidator{}
}

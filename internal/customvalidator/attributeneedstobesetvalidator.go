package customvalidator

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var _ validator.String = &attributeNeedsToBeSetValidator{}

type attributeNeedsToBeSetValidator struct {
	pathToBeSet path.Expression
	value       string
}

func (s attributeNeedsToBeSetValidator) Description(_ context.Context) string {
	finalStep, _ := s.pathToBeSet.Steps().LastStep()
	return fmt.Sprintf("Attribute \"%s\" needs to be set when value of the current field is \"%s\"", finalStep.String(), s.value)
}

func (s attributeNeedsToBeSetValidator) MarkdownDescription(ctx context.Context) string {
	return s.Description(ctx)
}

func (s attributeNeedsToBeSetValidator) ValidateString(ctx context.Context, request validator.StringRequest, response *validator.StringResponse) {
	// If the current attribute configuration is null or unknown, there
	// cannot be any value comparisons, so exit early without error.
	if request.ConfigValue.IsNull() || request.ConfigValue.IsUnknown() {
		return
	}

	// field is not set to the expected value, so we skip
	if request.ConfigValue.ValueString() != s.value {
		return
	}

	// Combine the given path expressions with the current attribute path
	// expression. This call automatically handles relative and absolute
	// expressions.
	expression := request.PathExpression.Merge(s.pathToBeSet)

	// Find paths matching the expression in the configuration data.
	matchedPaths, diags := request.Config.PathMatches(ctx, expression)

	response.Diagnostics.Append(diags...)

	// Collect all errors
	if diags.HasError() {
		return
	}

	for _, matchedPath := range matchedPaths {
		var matchedPathValue attr.Value

		diags = request.Config.GetAttribute(ctx, matchedPath, &matchedPathValue)

		response.Diagnostics.Append(diags...)

		// Collect all errors
		if diags.HasError() {
			continue
		}

		// If the matched path value is null or unknown, we cannot compare
		// values, so continue to other matched paths.
		if matchedPathValue.IsNull() || matchedPathValue.IsUnknown() {
			finalStep, _ := expression.Steps().LastStep()
			response.Diagnostics.AddAttributeError(
				request.Path,
				"Attribute Value Is Not Set",
				fmt.Sprintf("The attribute \"%s\" needs to be set as the value of \"%s\" is \"%s\".", finalStep.String(), request.Path.String(), s.value),
			)
		}
	}
}

func AttributeNeedsToBeSetValidator(pathToBeSet path.Expression, value string) validator.String {
	return attributeNeedsToBeSetValidator{
		pathToBeSet: pathToBeSet,
		value:       value,
	}
}

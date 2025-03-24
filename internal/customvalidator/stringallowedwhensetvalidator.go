package customvalidator

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ validator.String = &stringAllowedWhenSetValidator{}

type stringAllowedWhenSetValidator struct {
	expression path.Expression
	widgetType string
}

func (s stringAllowedWhenSetValidator) Description(_ context.Context) string {
	return fmt.Sprintf("Can only be configured for widget.id of type %s", s.widgetType)
}

func (s stringAllowedWhenSetValidator) MarkdownDescription(ctx context.Context) string {
	return s.Description(ctx)
}

func (s stringAllowedWhenSetValidator) ValidateString(ctx context.Context, request validator.StringRequest, response *validator.StringResponse) {
	// If the current attribute configuration is null or unknown, there
	// cannot be any value comparisons, so exit early without error.
	if request.ConfigValue.IsNull() || request.ConfigValue.IsUnknown() {
		return
	}

	// Combine the given path expressions with the current attribute path
	// expression. This call automatically handles relative and absolute
	// expressions.
	expression := request.PathExpression.Merge(s.expression)

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
			continue
		}

		// Now that we know the matched path value is not null or unknown,
		// it is safe to attempt converting it to the intended attr.Value
		// implementation, in this case a types.Int64 value.
		var matchedPathConfig types.String

		diags = tfsdk.ValueAs(ctx, matchedPathValue, &matchedPathConfig)

		response.Diagnostics.Append(diags...)

		// If the matched path value was not able to convert from
		// attr.Value to the intended types.Int64 implementation, it most
		// likely means that the path expression was not pointing at a
		// types.Int64Type attribute. Collect the error and continue to
		// other matched paths.
		if diags.HasError() {
			continue
		}

		if matchedPathConfig.ValueString() != s.widgetType {
			response.Diagnostics.AddAttributeError(
				request.Path,
				"Invalid Attribute Value",
				fmt.Sprintf("This value can only be set if the widget_id is \"%s\" but it is \"%s\". Path: %s", s.widgetType, matchedPathConfig.ValueString(), request.Path.String()),
			)
		}
	}
}

func StringAllowedWhenSetValidator(expression path.Expression, widgetType string) validator.String {
	return stringAllowedWhenSetValidator{
		expression: expression,
		widgetType: widgetType,
	}
}

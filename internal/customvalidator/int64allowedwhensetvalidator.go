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

var _ validator.Int64 = &int64AllowedWhenSetValidator{}

type int64AllowedWhenSetValidator struct {
	expression path.Expression
	widgetType string
}

func (i int64AllowedWhenSetValidator) Description(_ context.Context) string {
	return fmt.Sprintf("Can only be configured for widget.id of type %s", i.widgetType)
}

func (i int64AllowedWhenSetValidator) MarkdownDescription(ctx context.Context) string {
	return i.Description(ctx)
}

func (i int64AllowedWhenSetValidator) ValidateInt64(ctx context.Context, request validator.Int64Request, response *validator.Int64Response) {
	// If the current attribute configuration is null or unknown, there
	// cannot be any value comparisons, so exit early without error.
	if request.ConfigValue.IsNull() || request.ConfigValue.IsUnknown() {
		return
	}

	// Combine the given path expressions with the current attribute path
	// expression. This call automatically handles relative and absolute
	// expressions.
	expression := request.PathExpression.Merge(i.expression)

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

		// If the matched path value was not able to be converted from
		// attr.Value to the intended types.Int64 implementation, it most
		// likely means that the path expression was not pointing at a
		// types.Int64Type attribute. Collect the error and continue to
		// other matched paths.
		if diags.HasError() {
			continue
		}

		if matchedPathConfig.ValueString() != i.widgetType {
			response.Diagnostics.AddAttributeError(
				request.Path,
				"Invalid Attribute Value",
				fmt.Sprintf("This value can only be set if the widget_id is \"%s\" but it is \"%s\". Path: %s", i.widgetType, matchedPathConfig.ValueString(), request.Path.String()),
			)
		}
	}
}

func Int64AllowedWhenSetValidator(expression path.Expression, widgetType string) validator.Int64 {
	return int64AllowedWhenSetValidator{
		expression: expression,
		widgetType: widgetType,
	}
}

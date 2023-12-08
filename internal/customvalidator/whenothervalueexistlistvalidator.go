package customvalidator

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var _ validator.List = &whenOtherValueExistListValidator{}

type whenOtherValueExistListValidator struct {
	expression    path.Expression
	listValidator validator.List
}

func (l whenOtherValueExistListValidator) Description(_ context.Context) string {
	return "Validates a list when a given property is set"
}

func (l whenOtherValueExistListValidator) MarkdownDescription(ctx context.Context) string {
	return l.Description(ctx)
}

func (l whenOtherValueExistListValidator) ValidateList(ctx context.Context, request validator.ListRequest, response *validator.ListResponse) {
	// If the current attribute configuration is null or unknown, there
	// cannot be any value comparisons, so exit early without error.
	if request.ConfigValue.IsNull() || request.ConfigValue.IsUnknown() {
		return
	}

	// Combine the given path expressions with the current attribute path
	// expression. This call automatically handles relative and absolute
	// expressions.
	expression := request.PathExpression.Merge(l.expression)

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

		// when the matched path has a value we will execute the given list validation

		l.listValidator.ValidateList(ctx, request, response)
	}
}

func WhenOtherValueExistListValidator(expression path.Expression, listValidator validator.List) validator.List {
	return whenOtherValueExistListValidator{
		expression:    expression,
		listValidator: listValidator,
	}
}

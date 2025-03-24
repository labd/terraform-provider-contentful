package custommodifier

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// From https://developer.hashicorp.com/terraform/plugin/framework/resources/plan-modification#creating-attribute-plan-modifiers

var _ planmodifier.List = listDefaultModifier{}
var _ planmodifier.Describer = listDefaultModifier{}

type listDefaultModifier struct {
	Default []attr.Value
}

func (l listDefaultModifier) Description(_ context.Context) string {
	return fmt.Sprintf("If value is not configured, defaults to %s", l.Default)
}

func (l listDefaultModifier) MarkdownDescription(ctx context.Context) string {
	return l.Description(ctx)
}

func (l listDefaultModifier) PlanModifyList(ctx context.Context, _ planmodifier.ListRequest, resp *planmodifier.ListResponse) {
	// If the value is unknown or known, do not set default value.
	if resp.PlanValue.IsNull() || resp.PlanValue.IsUnknown() {
		planValue, diags := types.ListValue(resp.PlanValue.ElementType(ctx), l.Default)

		resp.Diagnostics.Append(diags...)

		if resp.Diagnostics.HasError() {
			return
		}

		resp.PlanValue = planValue
	}
}

func ListDefault(defaultValue []attr.Value) planmodifier.List {
	return listDefaultModifier{
		Default: defaultValue,
	}
}

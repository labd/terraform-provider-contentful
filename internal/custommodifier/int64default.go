package custommodifier

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// From https://developer.hashicorp.com/terraform/plugin/framework/resources/plan-modification#creating-attribute-plan-modifiers

var _ planmodifier.Int64 = int64DefaultModifier{}
var _ planmodifier.Describer = int64DefaultModifier{}

type int64DefaultModifier struct {
	Default int64
}

// Description returns a human-readable description of the plan modifier.
func (m int64DefaultModifier) Description(_ context.Context) string {
	return fmt.Sprintf("If value is not configured, defaults to %d", m.Default)
}

// MarkdownDescription returns a markdown description of the plan modifier.
func (m int64DefaultModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m int64DefaultModifier) PlanModifyInt64(_ context.Context, _ planmodifier.Int64Request, resp *planmodifier.Int64Response) {
	// If the value is unknown or known, do not set default value.
	if resp.PlanValue.IsNull() || resp.PlanValue.IsUnknown() {
		resp.PlanValue = types.Int64Value(m.Default)
	}
}

func Int64Default(defaultValue int64) planmodifier.Int64 {
	return int64DefaultModifier{
		Default: defaultValue,
	}
}

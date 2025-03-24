package custommodifier

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// From https://developer.hashicorp.com/terraform/plugin/framework/resources/plan-modification#creating-attribute-plan-modifiers

var _ planmodifier.Bool = boolDefaultModifier{}
var _ planmodifier.Describer = boolDefaultModifier{}

type boolDefaultModifier struct {
	Default bool
}

// Description returns a human-readable description of the plan modifier.
func (m boolDefaultModifier) Description(_ context.Context) string {
	return fmt.Sprintf("If value is not configured, defaults to %t", m.Default)
}

// MarkdownDescription returns a markdown description of the plan modifier.
func (m boolDefaultModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m boolDefaultModifier) PlanModifyBool(_ context.Context, _ planmodifier.BoolRequest, resp *planmodifier.BoolResponse) {
	// If the value is unknown or known, do not set default value.
	if resp.PlanValue.IsNull() || resp.PlanValue.IsUnknown() {
		resp.PlanValue = types.BoolValue(m.Default)
	}
}

func BoolDefault(defaultValue bool) planmodifier.Bool {
	return boolDefaultModifier{
		Default: defaultValue,
	}
}

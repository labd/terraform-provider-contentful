package custommodifier

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// From https://developer.hashicorp.com/terraform/plugin/framework/resources/plan-modification#creating-attribute-plan-modifiers

var _ planmodifier.String = stringDefaultModifier{}
var _ planmodifier.Describer = stringDefaultModifier{}

type stringDefaultModifier struct {
	Default string
}

// Description returns a human-readable description of the plan modifier.
func (m stringDefaultModifier) Description(_ context.Context) string {
	return fmt.Sprintf("If value is not configured, defaults to %s", m.Default)
}

// MarkdownDescription returns a markdown description of the plan modifier.
func (m stringDefaultModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m stringDefaultModifier) PlanModifyString(_ context.Context, _ planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	// If the value is unknown or known, do not set default value.
	if resp.PlanValue.IsNull() || resp.PlanValue.IsUnknown() {
		resp.PlanValue = types.StringValue(m.Default)
	}
}

func StringDefault(defaultValue string) planmodifier.String {
	return stringDefaultModifier{
		Default: defaultValue,
	}
}

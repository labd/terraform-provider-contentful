package custommodifier

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

// From https://developer.hashicorp.com/terraform/plugin/framework/resources/plan-modification#creating-attribute-plan-modifiers

var _ planmodifier.String = stringChangeProhibited{}
var _ planmodifier.String = stringChangeProhibited{}

type stringChangeProhibited struct {
	Message string
	Summary string
}

func (s stringChangeProhibited) Description(_ context.Context) string {
	return fmt.Sprint("If value is already configured, prevent any change")
}

func (s stringChangeProhibited) MarkdownDescription(ctx context.Context) string {
	return s.Description(ctx)
}

func (s stringChangeProhibited) PlanModifyString(_ context.Context, request planmodifier.StringRequest, response *planmodifier.StringResponse) {
	if request.StateValue.IsNull() {
		return
	}

	if !request.PlanValue.IsUnknown() && !request.PlanValue.Equal(request.StateValue) {
		// Return an example warning diagnostic to practitioners.
		response.Diagnostics.AddError(
			s.Summary,
			s.Message,
		)
	}
}

func StringChangeProhibited(summary string, detail string) planmodifier.String {
	return stringChangeProhibited{
		Message: detail,
		Summary: summary,
	}
}

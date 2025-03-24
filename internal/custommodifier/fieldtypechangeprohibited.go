package custommodifier

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// From https://developer.hashicorp.com/terraform/plugin/framework/resources/plan-modification#creating-attribute-plan-modifiers

var _ planmodifier.Object = fieldTypeChangeProhibited{}
var _ planmodifier.Object = fieldTypeChangeProhibited{}

type fieldTypeChangeProhibited struct{}

func (s fieldTypeChangeProhibited) Description(_ context.Context) string {
	return fmt.Sprint("If value is already configured, prevent any change")
}

func (s fieldTypeChangeProhibited) MarkdownDescription(ctx context.Context) string {
	return s.Description(ctx)
}

func (s fieldTypeChangeProhibited) PlanModifyObject(ctx context.Context, request planmodifier.ObjectRequest, response *planmodifier.ObjectResponse) {
	if request.StateValue.IsNull() {
		return
	}

	var fieldsList basetypes.ListValue

	diags := request.State.GetAttribute(ctx, request.Path.ParentPath(), &fieldsList)

	response.Diagnostics.Append(diags...)

	for _, value := range fieldsList.Elements() {

		stateFields := value.(basetypes.ObjectValue).Attributes()
		planFields := request.PlanValue.Attributes()

		if planFields["id"].Equal(stateFields["id"]) {
			if !planFields["type"].Equal(stateFields["type"]) {
				response.Diagnostics.AddError(
					fmt.Sprintf("Content Type Field Type Change for Field %s", planFields["id"].String()), "Changing a field type in contentful is not possible. Pls follow this faq: "+
						"https://www.contentful.com/faq/best-practices/#how-to-change-field-type",
				)
			}
		}
	}
}

func FieldTypeChangeProhibited() planmodifier.Object {
	return fieldTypeChangeProhibited{}
}

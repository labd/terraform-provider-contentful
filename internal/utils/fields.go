package utils

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func FromOptionalString(value string) basetypes.StringValue {
	if value == "" {
		return types.StringNull()
	}
	return types.StringValue(value)
}

func CompareStringPointer(tfPointer types.String, stringPointer *string) bool {
	if stringPointer != nil && tfPointer.ValueStringPointer() == nil {
		return false
	}

	if tfPointer.ValueStringPointer() != nil {

		if stringPointer == nil {
			return false
		}

		if tfPointer.ValueString() != *stringPointer {
			return false
		}
	}

	return true
}

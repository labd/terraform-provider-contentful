package locale

import (
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/labd/terraform-provider-contentful/internal/sdk"
	"github.com/labd/terraform-provider-contentful/internal/utils"
)

// Locale is the main resource schema data
type Locale struct {
	ID           types.String `tfsdk:"id"`
	Version      types.Int64  `tfsdk:"version"`
	SpaceID      types.String `tfsdk:"space_id"`
	Environment  types.String `tfsdk:"environment"`
	Name         types.String `tfsdk:"name"`
	Code         types.String `tfsdk:"code"`
	FallbackCode types.String `tfsdk:"fallback_code"`
	Optional     types.Bool   `tfsdk:"optional"`
	CDA          types.Bool   `tfsdk:"cda"`
	CMA          types.Bool   `tfsdk:"cma"`
}

// Import populates the Locale struct from an SDK locale object
func (l *Locale) Import(locale *sdk.Locale) {
	l.ID = types.StringValue(*locale.Sys.Id)
	l.Version = types.Int64Value(int64(*locale.Sys.Version))
	l.Name = types.StringValue(locale.Name)
	l.Code = types.StringValue(locale.Code)

	// Handle nullable fields
	if locale.FallbackCode != nil {
		l.FallbackCode = types.StringValue(*locale.FallbackCode)
	} else {
		l.FallbackCode = types.StringNull()
	}

	l.Optional = types.BoolValue(locale.Optional)
	l.CDA = types.BoolValue(locale.ContentDeliveryApi)
	l.CMA = types.BoolValue(locale.ContentManagementApi)
}

// DraftForCreate creates a LocaleCreate object for creating a new locale
func (l *Locale) DraftForCreate() sdk.LocaleCreate {
	localeCreate := sdk.LocaleCreate{
		Name:                 l.Name.ValueString(),
		Code:                 l.Code.ValueString(),
		Optional:             utils.Pointer(l.Optional.ValueBool()),
		ContentDeliveryApi:   utils.Pointer(l.CDA.ValueBool()),
		ContentManagementApi: utils.Pointer(l.CMA.ValueBool()),
	}

	if !l.FallbackCode.IsNull() && !l.FallbackCode.IsUnknown() {
		fallbackCode := l.FallbackCode.ValueString()
		localeCreate.FallbackCode = &fallbackCode
	}

	return localeCreate
}

// DraftForUpdate creates a LocaleUpdate object for updating an existing locale
func (l *Locale) DraftForUpdate() sdk.LocaleUpdate {
	localeUpdate := sdk.LocaleUpdate{
		Name:                 l.Name.ValueString(),
		Code:                 l.Code.ValueString(),
		Optional:             utils.Pointer(l.Optional.ValueBool()),
		ContentDeliveryApi:   utils.Pointer(l.CDA.ValueBool()),
		ContentManagementApi: utils.Pointer(l.CMA.ValueBool()),
	}

	if !l.FallbackCode.IsNull() && !l.FallbackCode.IsUnknown() {
		fallbackCode := l.FallbackCode.ValueString()
		localeUpdate.FallbackCode = &fallbackCode
	}

	return localeUpdate
}

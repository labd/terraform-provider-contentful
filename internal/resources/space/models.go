package space

import (
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/labd/terraform-provider-contentful/internal/sdk"
	"github.com/labd/terraform-provider-contentful/internal/utils"
)

// Space is the main resource schema data
type Space struct {
	ID            types.String `tfsdk:"id"`
	Version       types.Int64  `tfsdk:"version"`
	Name          types.String `tfsdk:"name"`
	DefaultLocale types.String `tfsdk:"default_locale"`
	AllowDeletion types.Bool   `tfsdk:"deletion_protection"`
}

// Space is the main datasource schema data
type SpaceData struct {
	ID            types.String `tfsdk:"id"`
	Version       types.Int64  `tfsdk:"version"`
	Name          types.String `tfsdk:"name"`
	DefaultLocale types.String `tfsdk:"default_locale"`
}

// Import populates the Space struct from an SDK space object
func (s *SpaceData) Import(space *sdk.Space) {
	s.ID = types.StringValue(space.Sys.Id)
	s.Version = types.Int64Value(int64(space.Sys.Version))
	s.Name = types.StringValue(space.Name)

	// Default locale is not directly exposed in the Space response
	// It needs to be fetched separately or passed separately
	if s.DefaultLocale.IsNull() || s.DefaultLocale.IsUnknown() {
		s.DefaultLocale = types.StringValue("en") // Default value
	}
}

// Import populates the Space struct from an SDK space object
func (s *Space) Import(space *sdk.Space) {
	s.ID = types.StringValue(space.Sys.Id)
	s.Version = types.Int64Value(int64(space.Sys.Version))
	s.Name = types.StringValue(space.Name)

	// Default locale is not directly exposed in the Space response
	// It needs to be fetched separately or passed separately
	if s.DefaultLocale.IsNull() || s.DefaultLocale.IsUnknown() {
		s.DefaultLocale = types.StringValue("en") // Default value
	}
}

// DraftForCreate creates a SpaceCreate object for creating a new space
func (s *Space) DraftForCreate() sdk.SpaceCreate {
	return sdk.SpaceCreate{
		Name:          s.Name.ValueString(),
		DefaultLocale: utils.Pointer(s.DefaultLocale.ValueString()),
	}
}

// DraftForUpdate creates a SpaceUpdate object for updating an existing space
func (s *Space) DraftForUpdate() sdk.SpaceUpdate {
	return sdk.SpaceUpdate{
		Name: s.Name.ValueString(),
		// Note: Default locale can't be updated directly via the Space update endpoint
		// It requires updating via the Locales endpoint
	}
}

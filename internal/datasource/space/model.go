package space

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/labd/terraform-provider-contentful/internal/sdk"
)

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
	s.Version = types.Int64Value(space.Sys.Version)
	s.Name = types.StringValue(space.Name)
	s.DefaultLocale = types.StringPointerValue(space.DefaultLocale)
}

package tag

import (
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/labd/terraform-provider-contentful/internal/sdk"
)

// Tag is the main resource schema data.
type Tag struct {
	ID          types.String `tfsdk:"id"`
	Version     types.Int64  `tfsdk:"version"`
	SpaceID     types.String `tfsdk:"space_id"`
	Environment types.String `tfsdk:"environment"`
	Name        types.String `tfsdk:"name"`
	Visibility  types.String `tfsdk:"visibility"`
}

// Import populates the Tag struct from an SDK tag object.
func (t *Tag) Import(tag *sdk.Tag) {
	t.ID = types.StringValue(tag.Sys.Id)
	t.Version = types.Int64Value(tag.Sys.Version)
	t.SpaceID = types.StringValue(tag.Sys.Space.Sys.Id)
	t.Environment = types.StringValue(tag.Sys.Environment.Sys.Id)
	t.Name = types.StringValue(tag.Name)
	t.Visibility = types.StringValue(string(tag.Sys.Visibility))
}

// VisibilityForCreate returns the configured visibility or Contentful's
// documented private default when visibility is omitted.
func (t *Tag) VisibilityForCreate() sdk.TagVisibility {
	if t.Visibility.IsNull() || t.Visibility.IsUnknown() {
		return sdk.TagVisibilityPrivate
	}

	return sdk.TagVisibility(t.Visibility.ValueString())
}

// Draft creates a TagDraft object for creating or updating a tag.
func (t *Tag) Draft() sdk.TagDraft {
	return sdk.TagDraft{
		Name: t.Name.ValueString(),
	}
}

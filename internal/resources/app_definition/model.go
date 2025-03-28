package app_definition

import (
	"github.com/elliotchance/pie/v2"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/labd/terraform-provider-contentful/internal/sdk"
	"github.com/labd/terraform-provider-contentful/internal/utils"
)

// AppDefinition is the main resource schema data
type AppDefinition struct {
	ID        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	Src       types.String `tfsdk:"src"`
	BundleId  types.String `tfsdk:"bundle_id"`
	UseBundle types.Bool   `tfsdk:"use_bundle"`
	Locations []Location   `tfsdk:"locations"`
}

func (a AppDefinition) Draft() *sdk.AppDefinitionDraft {
	app := &sdk.AppDefinitionDraft{
		Parameters: map[string]interface{}{},
	}

	if !a.Src.IsNull() && !a.Src.IsUnknown() {
		app.Src = a.Src.ValueStringPointer()
	}

	app.Name = a.Name.ValueString()

	app.Locations = pie.Map(a.Locations, func(t Location) sdk.AppLocation {
		return t.Draft()
	})

	if a.UseBundle.ValueBool() && !a.BundleId.IsNull() && !a.BundleId.IsUnknown() {
		app.Bundle = &sdk.AppDefinitionBundle{Sys: &sdk.SystemPropertiesLink{
			Id:       a.BundleId.ValueString(),
			Type:     "Link",
			LinkType: "AppBundle",
		}}
	}

	return app
}

func (a *AppDefinition) Equal(n *sdk.AppDefinition) bool {

	if a.Name.ValueString() != n.Name {
		return false
	}

	if !utils.CompareStringPointer(a.Src, n.Src) {
		return false
	}

	if len(a.Locations) != len(n.Locations) {
		return false
	}

	for _, location := range a.Locations {
		idx := pie.FindFirstUsing(n.Locations, func(f sdk.AppLocation) bool {
			return f.Location == location.Location.ValueString()
		})

		if idx == -1 {
			return false
		}
	}

	return true
}

func (a *AppDefinition) Import(n *sdk.AppDefinition) {
	a.ID = types.StringValue(n.Sys.Id)
	a.UseBundle = types.BoolValue(false)
	a.BundleId = types.StringNull()

	a.Name = types.StringValue(n.Name)

	a.Src = types.StringPointerValue(n.Src)

	fields := []Location{}
	for _, location := range n.Locations {
		field := &Location{}
		field.Import(location)
		fields = append(fields, *field)
	}

	if n.Bundle != nil {
		a.BundleId = types.StringValue(n.Bundle.Sys.Id)
		a.UseBundle = types.BoolValue(true)
	}

	a.Locations = fields
}

type Location struct {
	Location       types.String    `tfsdk:"location"`
	FieldTypes     []FieldType     `tfsdk:"field_types"`
	NavigationItem *NavigationItem `tfsdk:"navigation_item"`
}

func (l *Location) Import(n sdk.AppLocation) {
	l.Location = types.StringValue(n.Location)

	if n.NavigationItem != nil {
		l.NavigationItem = &NavigationItem{}
		l.NavigationItem.Import(n.NavigationItem)
	}

	if n.FieldTypes != nil {
		values := n.FieldTypes
		l.FieldTypes = pie.Map(values, func(t sdk.AppFieldType) FieldType {
			field := &FieldType{}
			field.Import(t)
			return *field
		})
	} else {
		l.FieldTypes = nil
	}

}

func (l *Location) Draft() sdk.AppLocation {
	location := sdk.AppLocation{
		Location: l.Location.ValueString(),
	}

	if l.NavigationItem != nil {
		location.NavigationItem = &sdk.AppNavigationItem{
			Name: l.NavigationItem.Name.ValueString(),
			Path: l.NavigationItem.Path.String(),
		}
	}

	location.FieldTypes = pie.Map(l.FieldTypes, func(t FieldType) sdk.AppFieldType {
		return t.Draft()
	})

	return location
}

type FieldType struct {
	Type     types.String `tfsdk:"type"`
	LinkType types.String `tfsdk:"link_type"`
	Items    *Items       `tfsdk:"items"`
}

func (f *FieldType) Draft() sdk.AppFieldType {
	result := sdk.AppFieldType{
		Type: sdk.AppFieldTypeType(f.Type.ValueString()),
	}

	if !f.LinkType.IsNull() && !f.LinkType.IsUnknown() {
		result.LinkType = utils.Pointer(sdk.AppFieldTypeLinkType(f.LinkType.ValueString()))
	}

	if f.Items != nil {
		result.Items = f.Items.Draft()
	}

	return result
}

func (f *FieldType) Import(n sdk.AppFieldType) {
	f.Type = types.StringValue(string(n.Type))

	if n.LinkType != nil {
		f.LinkType = types.StringValue(string(*n.LinkType))
	}

	if n.Items != nil {
		value, err := n.Items.ValueByDiscriminator()
		if err != nil {
			return
		}

		switch v := value.(type) {
		case *sdk.FieldItemSymbol:
			{
				f.Items = &Items{
					Type: types.StringValue(string(v.Type)),
				}
			}
		case *sdk.FieldItemLink:
			{
				f.Items = &Items{
					Type:     types.StringValue(string(v.Type)),
					LinkType: types.StringValue(string(v.LinkType)),
				}
			}
		}
	}
}

type Items struct {
	Type     types.String `tfsdk:"type"`
	LinkType types.String `tfsdk:"link_type"`
}

func (i *Items) Draft() *sdk.FieldItem {
	result := &sdk.FieldItem{}
	if !i.LinkType.IsNull() && !i.LinkType.IsUnknown() {
		result.FromFieldItemLink(sdk.FieldItemLink{
			Type:     sdk.FieldItemLinkType(i.Type.ValueString()),
			LinkType: sdk.FieldItemLinkLinkType(i.LinkType.ValueString()),
		})
	} else {
		result.FromFieldItemSymbol(sdk.FieldItemSymbol{
			Type: sdk.FieldItemSymbolType(i.Type.ValueString()),
		})
	}

	return result
}

type NavigationItem struct {
	Name types.String `tfsdk:"name"`
	Path types.String `tfsdk:"path"`
}

func (l *NavigationItem) Import(n *sdk.AppNavigationItem) {
	l.Name = types.StringValue(n.Name)
	l.Path = types.StringValue(n.Path)
}

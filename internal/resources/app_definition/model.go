package app_definition

import (
	"github.com/elliotchance/pie/v2"
	"github.com/flaconi/contentful-go/pkgs/model"
	"github.com/flaconi/terraform-provider-contentful/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/types"
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

func (a AppDefinition) Draft() *model.AppDefinition {

	app := &model.AppDefinition{}

	if !a.ID.IsUnknown() || !a.ID.IsNull() {
		app.Sys = &model.CreatedSys{BaseSys: model.BaseSys{ID: a.ID.ValueString()}}
	}

	if !a.Src.IsNull() && !a.Src.IsUnknown() {
		app.SRC = a.Src.ValueStringPointer()
	}

	app.Name = a.Name.ValueString()

	app.Locations = pie.Map(a.Locations, func(t Location) model.Locations {
		return t.Draft()
	})

	if a.UseBundle.ValueBool() && !a.BundleId.IsNull() && !a.BundleId.IsUnknown() {
		app.Bundle = &model.Bundle{Sys: &model.BaseSys{
			ID:       a.BundleId.ValueString(),
			Type:     "Link",
			LinkType: "AppBundle",
		}}
	}

	return app
}

func (a *AppDefinition) Equal(n *model.AppDefinition) bool {

	if a.Name.ValueString() != n.Name {
		return false
	}

	if !utils.CompareStringPointer(a.Src, n.SRC) {
		return false
	}

	if len(a.Locations) != len(n.Locations) {
		return false
	}

	for _, location := range a.Locations {
		idx := pie.FindFirstUsing(n.Locations, func(f model.Locations) bool {
			return f.Location == location.Location.ValueString()
		})

		if idx == -1 {
			return false
		}
	}

	return true
}

func (a *AppDefinition) Import(n *model.AppDefinition) {
	a.ID = types.StringValue(n.Sys.ID)
	a.UseBundle = types.BoolValue(false)
	a.BundleId = types.StringNull()

	a.Name = types.StringValue(n.Name)

	a.Src = types.StringPointerValue(n.SRC)

	fields := []Location{}

	for _, location := range n.Locations {
		field := &Location{}
		field.Import(location)
		fields = append(fields, *field)
	}

	if n.Bundle != nil {
		a.BundleId = types.StringValue(n.Bundle.Sys.ID)
		a.UseBundle = types.BoolValue(true)
	}

	a.Locations = fields
}

type Location struct {
	Location       types.String    `tfsdk:"location"`
	FieldTypes     []FieldType     `tfsdk:"field_types"`
	NavigationItem *NavigationItem `tfsdk:"navigation_item"`
}

func (l *Location) Import(n model.Locations) {
	l.Location = types.StringValue(n.Location)

	if n.NavigationItem != nil {
		l.NavigationItem = &NavigationItem{}
		l.NavigationItem.Import(n.NavigationItem)
	}

	l.FieldTypes = pie.Map(n.FieldTypes, func(t model.FieldType) FieldType {
		field := &FieldType{}
		field.Import(t)
		return *field
	})
}

func (l *Location) Draft() model.Locations {
	location := model.Locations{
		Location: l.Location.ValueString(),
	}

	if l.NavigationItem != nil {
		location.NavigationItem = &model.NavigationItem{
			Name: l.NavigationItem.Name.ValueString(),
			Path: l.NavigationItem.Path.String(),
		}
	}

	location.FieldTypes = pie.Map(l.FieldTypes, func(t FieldType) model.FieldType {
		return t.Draft()
	})

	return location
}

type FieldType struct {
	Type     types.String `tfsdk:"type"`
	LinkType types.String `tfsdk:"link_type"`
	Items    *Items       `tfsdk:"items"`
}

func (f *FieldType) Draft() model.FieldType {
	fieldType := model.FieldType{
		Type:     f.Type.ValueString(),
		LinkType: f.LinkType.ValueStringPointer(),
	}

	if f.Items != nil {
		fieldType.Items = f.Items.Draft()
	}

	return fieldType
}

func (f *FieldType) Import(n model.FieldType) {
	f.Type = types.StringValue(n.Type)
	f.LinkType = types.StringPointerValue(n.LinkType)

	if n.Items != nil {
		f.Items = &Items{
			Type:     types.StringValue(n.Items.Type),
			LinkType: types.StringPointerValue(n.Items.LinkType),
		}
	}
}

type Items struct {
	Type     types.String `tfsdk:"type"`
	LinkType types.String `tfsdk:"link_type"`
}

func (i *Items) Draft() *model.Items {
	return &model.Items{
		Type:     i.Type.ValueString(),
		LinkType: i.LinkType.ValueStringPointer(),
	}
}

type NavigationItem struct {
	Name types.String `tfsdk:"name"`
	Path types.String `tfsdk:"path"`
}

func (l *NavigationItem) Import(n *model.NavigationItem) {
	l.Name = types.StringValue(n.Name)
	l.Path = types.StringValue(n.Path)
}

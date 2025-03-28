package contenttype

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"

	"github.com/elliotchance/pie/v2"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/labd/terraform-provider-contentful/internal/sdk"
	"github.com/labd/terraform-provider-contentful/internal/utils"
)

// ContentType is the main resource schema data
type ContentType struct {
	ID                  types.String `tfsdk:"id"`
	SpaceId             types.String `tfsdk:"space_id"`
	Environment         types.String `tfsdk:"environment"`
	Name                types.String `tfsdk:"name"`
	DisplayField        types.String `tfsdk:"display_field"`
	Description         types.String `tfsdk:"description"`
	Version             types.Int64  `tfsdk:"version"`
	VersionControls     types.Int64  `tfsdk:"version_controls"`
	Fields              []Field      `tfsdk:"fields"`
	ManageFieldControls types.Bool   `tfsdk:"manage_field_controls"`
	Sidebar             []Sidebar    `tfsdk:"sidebar"`
}

type Sidebar struct {
	WidgetId        types.String         `tfsdk:"widget_id"`
	WidgetNamespace types.String         `tfsdk:"widget_namespace"`
	Settings        jsontypes.Normalized `tfsdk:"settings"`
	Disabled        types.Bool           `tfsdk:"disabled"`
}

type Field struct {
	Id           types.String  `tfsdk:"id"`
	Name         types.String  `tfsdk:"name"`
	Type         types.String  `tfsdk:"type"`
	LinkType     types.String  `tfsdk:"link_type"`
	Required     types.Bool    `tfsdk:"required"`
	Localized    types.Bool    `tfsdk:"localized"`
	Disabled     types.Bool    `tfsdk:"disabled"`
	Omitted      types.Bool    `tfsdk:"omitted"`
	Validations  []Validation  `tfsdk:"validations"`
	Items        *Items        `tfsdk:"items"`
	Control      *Control      `tfsdk:"control"`
	DefaultValue *DefaultValue `tfsdk:"default_value"`
}

type DefaultValue struct {
	Bool   types.Map `tfsdk:"bool"`
	String types.Map `tfsdk:"string"`
}

func (d *DefaultValue) Draft() *map[string]any {
	var defaultValues = map[string]any{}

	if !d.String.IsNull() && !d.String.IsUnknown() {

		for k, v := range d.String.Elements() {
			defaultValues[k] = v.(types.String).ValueString()
		}
	}

	if !d.Bool.IsNull() && !d.Bool.IsUnknown() {

		for k, v := range d.Bool.Elements() {
			defaultValues[k] = v.(types.Bool).ValueBool()
		}
	}

	if len(defaultValues) == 0 {
		return nil
	}

	return &defaultValues
}

type Control struct {
	WidgetId        types.String `tfsdk:"widget_id"`
	WidgetNamespace types.String `tfsdk:"widget_namespace"`
	Settings        *Settings    `tfsdk:"settings"`
}

type Validation struct {
	Unique            types.Bool     `tfsdk:"unique"`
	Size              *Size          `tfsdk:"size"`
	Range             *Size          `tfsdk:"range"`
	AssetFileSize     *Size          `tfsdk:"asset_file_size"`
	Regexp            *Regexp        `tfsdk:"regexp"`
	LinkContentType   []types.String `tfsdk:"link_content_type"`
	LinkMimetypeGroup []types.String `tfsdk:"link_mimetype_group"`
	In                []types.String `tfsdk:"in"`
	EnabledMarks      []types.String `tfsdk:"enabled_marks"`
	EnabledNodeTypes  []types.String `tfsdk:"enabled_node_types"`
	Message           types.String   `tfsdk:"message"`
}

func (v Validation) Draft() (*sdk.FieldValidation, error) {

	base := &sdk.FieldValidation{
		Message: v.Message.ValueStringPointer(),
	}

	if !v.Unique.IsUnknown() && !v.Unique.IsNull() {
		base.Unique = v.Unique.ValueBoolPointer()
		return base, nil
	}

	if v.Size != nil {
		base.Size = &sdk.RangeMinMax{
			Min: v.Size.Min.ValueFloat64Pointer(),
			Max: v.Size.Max.ValueFloat64Pointer(),
		}
		return base, nil
	}

	if v.Range != nil {
		base.Range = &sdk.RangeMinMax{
			Min: v.Range.Min.ValueFloat64Pointer(),
			Max: v.Range.Max.ValueFloat64Pointer(),
		}
		return base, nil
	}

	if v.AssetFileSize != nil {
		base.AssetFileSize = &sdk.RangeMinMax{
			Min: v.AssetFileSize.Min.ValueFloat64Pointer(),
			Max: v.AssetFileSize.Max.ValueFloat64Pointer(),
		}
		return base, nil
	}

	if v.Regexp != nil {
		base.Regexp = &sdk.RegexValidationValue{
			Pattern: v.Regexp.Pattern.ValueString(),
		}
		return base, nil
	}

	if len(v.LinkContentType) > 0 {
		value := pie.Map(v.LinkContentType, func(t types.String) string {
			return t.ValueString()
		})
		base.LinkContentType = &value
		return base, nil
	}

	if len(v.LinkMimetypeGroup) > 0 {
		value := pie.Map(v.LinkMimetypeGroup, func(t types.String) string {
			return t.ValueString()
		})
		base.LinkMimetypeGroup = &value
		return base, nil
	}

	if len(v.In) > 0 {
		value := pie.Map(v.In, func(t types.String) string {
			return t.ValueString()
		})
		base.In = &value
		return base, nil
	}

	if len(v.EnabledMarks) > 0 {
		value := pie.Map(v.EnabledMarks, func(t types.String) string {
			return t.ValueString()
		})
		base.EnabledMarks = &value
		return base, nil
	}

	if len(v.EnabledNodeTypes) > 0 {
		value := pie.Map(v.EnabledNodeTypes, func(t types.String) string {
			return t.ValueString()
		})
		base.EnabledNodeTypes = &value
		return base, nil
	}

	return nil, fmt.Errorf("unsupported validation used, %s. Please implement", reflect.TypeOf(v).String())
}

type Size struct {
	Min types.Float64 `tfsdk:"min"`
	Max types.Float64 `tfsdk:"max"`
}

type Regexp struct {
	Pattern types.String `tfsdk:"pattern"`
}

func (f *Field) Equal(n sdk.Field) bool {

	if string(n.Type) != f.Type.ValueString() {
		return false
	}

	if n.Id != f.Id.ValueString() {
		return false
	}

	if n.Name != f.Name.ValueString() {
		return false
	}

	if n.LinkType != nil && string(*n.LinkType) != f.LinkType.ValueString() {
		return false
	}

	if n.LinkType == nil && !f.LinkType.IsNull() {
		return false
	}

	if n.Required != f.Required.ValueBool() {
		return false
	}

	if n.Omitted != f.Omitted.ValueBoolPointer() {
		return false
	}

	if n.Disabled != f.Disabled.ValueBoolPointer() {
		return false
	}

	if n.Localized != f.Localized.ValueBool() {
		return false
	}

	if f.Items == nil && n.Items != nil {
		return false
	}

	if f.Items != nil && !f.Items.Equal(n.Items) {
		return false
	}

	if !compareValidations(f.Validations, n.Validations) {
		return false
	}

	if f.DefaultValue != nil && !reflect.DeepEqual(f.DefaultValue.Draft(), n.DefaultValue) {
		return false
	}

	return true
}

func createValidations(validations []Validation) ([]sdk.FieldValidation, error) {
	var contentfulValidations []sdk.FieldValidation
	for _, validation := range validations {
		value, err := validation.Draft()
		if err != nil {
			return nil, err
		}
		contentfulValidations = append(contentfulValidations, *value)
	}
	return contentfulValidations, nil
}

func (f *Field) ToNative() (*sdk.Field, error) {

	validations, err := createValidations(f.Validations)
	if err != nil {
		return nil, err
	}

	contentfulField := &sdk.Field{
		Id:          f.Id.ValueString(),
		Name:        f.Name.ValueString(),
		Type:        sdk.FieldType(f.Type.ValueString()),
		Localized:   f.Localized.ValueBool(),
		Required:    f.Required.ValueBool(),
		Disabled:    f.Disabled.ValueBoolPointer(),
		Omitted:     f.Omitted.ValueBoolPointer(),
		Validations: &validations,
	}

	if !f.LinkType.IsNull() && !f.LinkType.IsUnknown() {
		contentfulField.LinkType = utils.Pointer(sdk.FieldLinkType(f.LinkType.ValueString()))
	}

	if contentfulField.Type == sdk.FieldTypeArray {
		items, errItem := f.Items.ToNative()

		if errItem != nil {
			return nil, errItem
		}

		contentfulField.Items = items
	}

	if f.DefaultValue != nil {
		contentfulField.DefaultValue = f.DefaultValue.Draft()
	}

	return contentfulField, nil
}

func getTypeOfMap(mapValues *map[string]any) (*string, error) {
	if mapValues == nil {
		return nil, nil
	}

	for _, v := range *mapValues {
		switch c := v.(type) {
		case string:
			t := "string"
			return &t, nil
		case bool:
			t := "bool"
			return &t, nil
		case float64:
			t := "float64"
			return &t, nil
		default:
			return nil, fmt.Errorf("The default type %T is not supported by the provider", c)
		}
	}

	return nil, nil
}

func (f *Field) Import(n sdk.Field, c []sdk.EditorInterfaceControl) error {
	f.Id = types.StringValue(n.Id)
	f.Name = types.StringValue(n.Name)
	f.Type = types.StringValue(string(n.Type))
	f.Required = types.BoolValue(n.Required)
	f.Omitted = types.BoolPointerValue(n.Omitted)
	f.Localized = types.BoolValue(n.Localized)
	f.Disabled = types.BoolPointerValue(n.Disabled)

	if n.LinkType == nil {
		f.LinkType = types.StringNull()
	} else {
		f.LinkType = types.StringValue(string(*n.LinkType))
	}

	defaultValueType, err := getTypeOfMap(n.DefaultValue)
	if err != nil {
		return err
	}

	if defaultValueType != nil {

		f.DefaultValue = &DefaultValue{
			Bool:   types.MapNull(types.BoolType),
			String: types.MapNull(types.StringType),
		}

		if n.DefaultValue == nil {
			return fmt.Errorf("default value is nil")
		}

		switch *defaultValueType {
		case "string":
			stringMap := map[string]attr.Value{}

			for k, v := range *n.DefaultValue {
				stringMap[k] = types.StringValue(v.(string))
			}

			f.DefaultValue.String = types.MapValueMust(types.StringType, stringMap)
		case "bool":
			boolMap := map[string]attr.Value{}

			for k, v := range *n.DefaultValue {
				boolMap[k] = types.BoolValue(v.(bool))
			}

			f.DefaultValue.Bool = types.MapValueMust(types.BoolType, boolMap)
		}

	}

	validations, err := getValidations(n.Validations)

	if err != nil {
		return err
	}

	f.Validations = validations

	if n.Type == sdk.FieldTypeArray {

		itemType, err := n.Items.Discriminator()
		if err != nil {
			return err
		}

		if itemType == "FieldItemSymbol" {
			symbolItem, err := n.Items.AsFieldItemSymbol()
			if err != nil {
				return err
			}

			itemValidations, err := getValidations(symbolItem.Validations)
			if err != nil {
				return err
			}

			f.Items = &Items{
				Type:        types.StringValue(itemType),
				Validations: itemValidations,
			}
		} else {
			linkItem, err := n.Items.AsFieldItemLink()
			if err != nil {
				return err
			}

			itemValidations, err := getValidations(linkItem.Validations)
			if err != nil {
				return err
			}

			f.Items = &Items{
				Type:        types.StringValue(itemType),
				LinkType:    types.StringValue(string(linkItem.LinkType)),
				Validations: itemValidations,
			}
		}
	}

	idx := pie.FindFirstUsing(c, func(control sdk.EditorInterfaceControl) bool {
		return n.Id == control.FieldId
	})

	if idx != -1 && c[idx].WidgetId != nil {

		var settings *Settings

		if c[idx].Settings != nil {
			settings = &Settings{}

			settings.Import(c[idx].Settings)
		}

		var namespace *string = nil
		if c[idx].WidgetNamespace != nil {
			namespace = utils.Pointer(string(*c[idx].WidgetNamespace))
		}

		f.Control = &Control{
			WidgetId:        types.StringPointerValue(c[idx].WidgetId),
			WidgetNamespace: types.StringPointerValue(namespace),
			Settings:        settings,
		}
	}

	return nil
}

type Settings struct {
	HelpText        types.String `tfsdk:"help_text"`
	TrueLabel       types.String `tfsdk:"true_label"`
	FalseLabel      types.String `tfsdk:"false_label"`
	Stars           types.Int64  `tfsdk:"stars"`
	Format          types.String `tfsdk:"format"`
	TimeFormat      types.String `tfsdk:"ampm"`
	BulkEditing     types.Bool   `tfsdk:"bulk_editing"`
	TrackingFieldId types.String `tfsdk:"tracking_field_id"`
}

func (s *Settings) Import(settings *sdk.EditorInterfaceSettings) {
	if settings.Stars != nil {
		numStars, err := strconv.ParseInt(*settings.Stars, 10, 64)
		if err != nil {
			numStars = 0
		}
		s.Stars = types.Int64Value(numStars)
	}

	s.HelpText = types.StringPointerValue(settings.HelpText)
	s.TrueLabel = types.StringPointerValue(settings.TrueLabel)
	s.FalseLabel = types.StringPointerValue(settings.FalseLabel)
	s.Format = types.StringPointerValue(settings.Format)
	s.TimeFormat = types.StringPointerValue(settings.Ampm)
	s.BulkEditing = types.BoolPointerValue(settings.BulkEditing)
	s.TrackingFieldId = types.StringPointerValue(settings.TrackingFieldId)
}

func (s *Settings) Draft() *sdk.EditorInterfaceSettings {
	settings := &sdk.EditorInterfaceSettings{}

	if !s.Stars.IsNull() && !s.Stars.IsUnknown() {
		settings.Stars = utils.Pointer(strconv.FormatInt(s.Stars.ValueInt64(), 10))
	}

	settings.HelpText = s.HelpText.ValueStringPointer()
	settings.TrueLabel = s.TrueLabel.ValueStringPointer()
	settings.FalseLabel = s.FalseLabel.ValueStringPointer()
	settings.Format = s.Format.ValueStringPointer()
	settings.Ampm = s.TimeFormat.ValueStringPointer()
	settings.BulkEditing = s.BulkEditing.ValueBoolPointer()
	settings.TrackingFieldId = s.TrackingFieldId.ValueStringPointer()
	return settings
}

type Items struct {
	Type        types.String `tfsdk:"type"`
	LinkType    types.String `tfsdk:"link_type"`
	Validations []Validation `tfsdk:"validations"`
}

func (i *Items) ToNative() (*sdk.FieldItem, error) {

	validations, err := createValidations(i.Validations)
	if err != nil {
		return nil, err
	}

	fieldType := i.Type.ValueString()

	item := sdk.FieldItem{}

	if fieldType == "Symbol" {
		err := item.FromFieldItemSymbol(sdk.FieldItemSymbol{
			Validations: &validations,
		})
		if err != nil {
			return nil, err
		}
		return &item, nil
	}

	if fieldType == "Link" || fieldType == "ResourceLink" {
		err := item.FromFieldItemLink(sdk.FieldItemLink{
			Validations: &validations,
			LinkType:    sdk.FieldItemLinkLinkType(i.LinkType.ValueString()),
		})

		if err != nil {
			return nil, err
		}

		return &item, nil
	}

	return nil, fmt.Errorf("unsupported item type used, %s. Please implement", fieldType)
}

func (i *Items) Equal(n *sdk.FieldItem) bool {

	if n == nil {
		return false
	}

	itemType, err := n.Discriminator()
	if err != nil {
		panic(err)
	}

	if i.Type.ValueString() != itemType {
		return false
	}

	if itemType != "Symbol" {
		linkItem, err := n.AsFieldItemLink()
		if err != nil {
			panic(err)
		}

		if !utils.CompareStringPointer(i.LinkType, utils.Pointer(string(linkItem.LinkType))) {
			return false
		}

		if !compareValidations(i.Validations, linkItem.Validations) {
			return false
		}

	} else {
		symbolItem, err := n.AsFieldItemSymbol()
		if err != nil {
			panic(err)
		}

		if !compareValidations(i.Validations, symbolItem.Validations) {
			return false
		}
	}

	return true
}

func (c *ContentType) Create() (*sdk.ContentTypeCreate, error) {
	var fields []sdk.Field

	for _, field := range c.Fields {

		nativeField, err := field.ToNative()
		if err != nil {
			return nil, err
		}

		fields = append(fields, *nativeField)
	}

	contentfulType := &sdk.ContentTypeCreate{
		Name:         c.Name.ValueString(),
		DisplayField: c.DisplayField.ValueString(),
		Fields:       fields,
	}

	if !c.Description.IsNull() && !c.Description.IsUnknown() {
		contentfulType.Description = c.Description.ValueStringPointer()
	}

	return contentfulType, nil
}

func (c *ContentType) Update() (*sdk.ContentTypeUpdate, error) {
	var fields []sdk.Field

	for _, field := range c.Fields {

		nativeField, err := field.ToNative()
		if err != nil {
			return nil, err
		}

		fields = append(fields, *nativeField)
	}

	contentfulType := &sdk.ContentTypeUpdate{
		Name:         c.Name.ValueString(),
		DisplayField: c.DisplayField.ValueString(),
		Fields:       fields,
	}

	if !c.Description.IsNull() && !c.Description.IsUnknown() {
		contentfulType.Description = c.Description.ValueStringPointer()
	}

	return contentfulType, nil
}

func (c *ContentType) Import(n *sdk.ContentType, e *sdk.EditorInterface) error {
	c.ID = types.StringValue(n.Sys.Id)
	c.Version = types.Int64Value(n.Sys.Version)

	c.Description = types.StringPointerValue(n.Description)

	c.Name = types.StringValue(n.Name)
	c.DisplayField = types.StringValue(n.DisplayField)

	var fields []Field

	var controls []sdk.EditorInterfaceControl
	var sidebar []sdk.EditorInterfaceSidebarItem
	c.VersionControls = types.Int64Value(0)
	if e != nil {
		controls = e.Controls

		if e.Sidebar != nil {
			sidebar = *e.Sidebar
		}
		c.VersionControls = types.Int64Value(int64(*e.Sys.Version))
	}

	for _, nf := range n.Fields {
		field := &Field{}
		err := field.Import(nf, controls)
		if err != nil {
			return fmt.Errorf("field import failed: %w", err)
		}
		fields = append(fields, *field)
	}

	c.Sidebar = pie.Map(sidebar, func(t sdk.EditorInterfaceSidebarItem) Sidebar {

		settings := jsontypes.NewNormalizedValue("{}")

		if t.Settings != nil {
			data, _ := json.Marshal(t.Settings)
			settings = jsontypes.NewNormalizedValue(string(data))
		}
		return Sidebar{
			WidgetId:        types.StringValue(*t.WidgetId),
			WidgetNamespace: types.StringValue(string(*t.WidgetNamespace)),
			Settings:        settings,
			Disabled:        types.BoolPointerValue(t.Disabled),
		}
	})

	c.Fields = fields

	return nil

}

func (c *ContentType) Equal(n *sdk.ContentType) bool {

	if !utils.CompareStringPointer(c.Description, n.Description) {
		return false
	}

	if c.Name.ValueString() != n.Name {
		return false
	}

	if c.DisplayField.ValueString() != n.DisplayField {
		return false
	}

	if len(c.Fields) != len(n.Fields) {
		return false
	}

	for idxOrg, field := range c.Fields {
		idx := pie.FindFirstUsing(n.Fields, func(f sdk.Field) bool {
			return f.Id == field.Id.ValueString()
		})

		if idx == -1 {
			return false
		}

		if !field.Equal(n.Fields[idx]) {
			return false
		}

		// field was moved, it is the same as before but different position
		if idxOrg != idx {
			return false
		}
	}

	return true
}

func (c *ContentType) EqualEditorInterface(n *sdk.EditorInterface) bool {

	if len(c.Fields) != len(n.Controls) {
		return false
	}

	filteredControls := pie.Filter(n.Controls, func(c sdk.EditorInterfaceControl) bool {
		return c.WidgetId != nil || c.WidgetNamespace != nil || c.Settings != nil
	})

	filteredFields := pie.Filter(c.Fields, func(f Field) bool {
		return f.Control != nil
	})

	if len(filteredControls) != len(filteredFields) {
		return false
	}

	for _, field := range filteredFields {
		idx := pie.FindFirstUsing(filteredControls, func(t sdk.EditorInterfaceControl) bool {
			return t.FieldId == field.Id.ValueString()
		})

		if idx == -1 {
			return false
		}
		control := filteredControls[idx]

		if field.Control.WidgetId.ValueString() != *control.WidgetId {
			return false
		}

		var namespace *string = nil
		if control.WidgetNamespace != nil {
			namespace = utils.Pointer(string(*control.WidgetNamespace))
		}
		if field.Control.WidgetNamespace.ValueStringPointer() != namespace {
			return false
		}

		if field.Control.Settings == nil && control.Settings != nil {
			return false
		}

		if field.Control.Settings != nil && !reflect.DeepEqual(field.Control.Settings.Draft(), control.Settings) {
			return false
		}
	}

	if n.Sidebar == nil && len(c.Sidebar) > 0 {
		return false
	}

	if len(c.Sidebar) != len(*n.Sidebar) {
		return false
	}

	sidebar := *n.Sidebar

	for idxOrg, s := range c.Sidebar {
		idx := pie.FindFirstUsing(sidebar, func(t sdk.EditorInterfaceSidebarItem) bool {
			return t.WidgetId == s.WidgetId.ValueStringPointer()
		})

		if idx == -1 {
			return false
		}

		// field was moved, it is the same as before but different position
		if idxOrg != idx {
			return false
		}

		sidebar := sidebar[idx]

		if sidebar.Disabled != s.Disabled.ValueBoolPointer() {
			return false
		}

		if sidebar.WidgetId != s.WidgetId.ValueStringPointer() {
			return false
		}

		var namespace *string = nil
		if !s.WidgetNamespace.IsNull() {
			namespace = utils.Pointer(s.WidgetNamespace.ValueString())
		}

		if namespace != s.WidgetNamespace.ValueStringPointer() {
			return false
		}

		a := make(map[string]string)

		s.Settings.Unmarshal(a)

		if !reflect.DeepEqual(sidebar.Settings, a) {
			return false
		}
	}

	return true
}

func (c *ContentType) DraftEditorInterface(n *sdk.EditorInterface) {

	n.Controls = pie.Map(c.Fields, func(field Field) sdk.EditorInterfaceControl {

		control := sdk.EditorInterfaceControl{
			FieldId: field.Id.ValueString(),
		}

		if field.Control != nil {
			control.WidgetId = field.Control.WidgetId.ValueStringPointer()
			control.WidgetNamespace = utils.Pointer(sdk.EditorInterfaceControlWidgetNamespace(field.Control.WidgetNamespace.ValueString()))

			if field.Control.Settings != nil {
				control.Settings = field.Control.Settings.Draft()
			}
		}

		return control

	})

	sidebar := pie.Map(c.Sidebar, func(t Sidebar) sdk.EditorInterfaceSidebarItem {

		var namespace sdk.EditorInterfaceSidebarItemWidgetNamespace
		if !t.WidgetNamespace.IsNull() {
			namespace = sdk.EditorInterfaceSidebarItemWidgetNamespace(t.WidgetNamespace.ValueString())
		}

		sidebar := sdk.EditorInterfaceSidebarItem{
			WidgetNamespace: &namespace,
			WidgetId:        t.WidgetId.ValueStringPointer(),
			Disabled:        t.Disabled.ValueBoolPointer(),
		}

		if !*sidebar.Disabled {
			settings := sdk.EditorInterfaceSettings{}

			t.Settings.Unmarshal(settings)
			sidebar.Settings = &settings
		}

		return sidebar
	})

	if len(sidebar) > 0 {
		n.Sidebar = &sidebar
	} else {
		n.Sidebar = nil
	}
}

func getValidations(contentfulValidations *[]sdk.FieldValidation) ([]Validation, error) {
	var validations []Validation

	if contentfulValidations == nil {
		return validations, nil
	}

	for _, validation := range *contentfulValidations {

		val, err := getValidation(validation)

		if err != nil {
			return nil, err
		}

		validations = append(validations, *val)
	}

	return validations, nil
}

func getValidation(cfVal sdk.FieldValidation) (*Validation, error) {

	if cfVal.AssetFileSize != nil {
		return &Validation{
			Range: &Size{
				Max: types.Float64PointerValue(cfVal.AssetFileSize.Max),
				Min: types.Float64PointerValue(cfVal.AssetFileSize.Min),
			},
			Message: types.StringPointerValue(cfVal.Message),
		}, nil
	}

	if cfVal.Regexp != nil {
		return &Validation{
			Regexp: &Regexp{
				Pattern: types.StringValue(cfVal.Regexp.Pattern),
			},
			Message: types.StringPointerValue(cfVal.Message),
		}, nil
	}

	if cfVal.LinkContentType != nil {
		return &Validation{
			LinkContentType: pie.Map(*cfVal.LinkContentType, func(t string) types.String {
				return types.StringValue(t)
			}),
			Message: types.StringPointerValue(cfVal.Message),
		}, nil
	}

	if cfVal.LinkMimetypeGroup != nil {
		return &Validation{
			LinkMimetypeGroup: pie.Map(*cfVal.LinkMimetypeGroup, func(t string) types.String {
				return types.StringValue(t)
			}),
			Message: types.StringPointerValue(cfVal.Message),
		}, nil
	}

	if cfVal.In != nil {
		return &Validation{
			In: pie.Map(*cfVal.In, func(t string) types.String {
				return types.StringValue(t)
			}),
			Message: types.StringPointerValue(cfVal.Message),
		}, nil
	}

	if cfVal.EnabledMarks != nil {
		return &Validation{
			EnabledMarks: pie.Map(*cfVal.EnabledMarks, func(t string) types.String {
				return types.StringValue(t)
			}),
			Message: types.StringPointerValue(cfVal.Message),
		}, nil
	}

	if cfVal.EnabledNodeTypes != nil {
		return &Validation{
			EnabledNodeTypes: pie.Map(*cfVal.EnabledNodeTypes, func(t string) types.String {
				return types.StringValue(t)
			}),
			Message: types.StringPointerValue(cfVal.Message),
		}, nil
	}

	if cfVal.AssetFileSize != nil {
		return &Validation{
			AssetFileSize: &Size{
				Max: types.Float64PointerValue(cfVal.AssetFileSize.Max),
				Min: types.Float64PointerValue(cfVal.AssetFileSize.Min),
			},
			Message: types.StringPointerValue(cfVal.Message),
		}, nil
	}

	if cfVal.Unique != nil {
		return &Validation{
			Unique:  types.BoolPointerValue(cfVal.Unique),
			Message: types.StringPointerValue(cfVal.Message),
		}, nil
	}

	return nil, fmt.Errorf("unsupported validation used, %s. Please implement", reflect.TypeOf(cfVal).String())
}

func compareValidations(a []Validation, b *[]sdk.FieldValidation) bool {
	if b == nil {
		return len(a) == 0
	}

	other := *b

	if len(a) != len(other) {
		return false
	}

	validations, err := createValidations(a)
	if err != nil {
		panic(err)
	}

	for idx, validation := range validations {
		cfVal := other[idx]

		if !reflect.DeepEqual(validation, cfVal) {
			return false
		}
	}

	return true
}

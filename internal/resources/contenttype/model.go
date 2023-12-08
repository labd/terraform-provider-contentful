package contenttype

import (
	"fmt"
	"github.com/elliotchance/pie/v2"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/labd/contentful-go"
	"github.com/labd/terraform-provider-contentful/internal/utils"
	"reflect"
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

func (d *DefaultValue) Draft() map[string]any {
	var default_values = map[string]any{}

	if !d.String.IsNull() && !d.String.IsUnknown() {

		for k, v := range d.String.Elements() {
			default_values[k] = v.(types.String).ValueString()
		}
	}

	if !d.Bool.IsNull() && !d.Bool.IsUnknown() {

		for k, v := range d.Bool.Elements() {
			default_values[k] = v.(types.Bool).ValueBool()
		}
	}

	return default_values
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

func (v Validation) Draft() contentful.FieldValidation {

	if !v.Unique.IsUnknown() && !v.Unique.IsNull() {
		return contentful.FieldValidationUnique{
			Unique: v.Unique.ValueBool(),
		}
	}

	if v.Size != nil {
		return contentful.FieldValidationSize{
			Size: &contentful.MinMax{
				Min: v.Size.Min.ValueFloat64Pointer(),
				Max: v.Size.Max.ValueFloat64Pointer(),
			},
			ErrorMessage: v.Message.ValueStringPointer(),
		}
	}

	if v.Range != nil {
		return contentful.FieldValidationRange{
			Range: &contentful.MinMax{
				Min: v.Range.Min.ValueFloat64Pointer(),
				Max: v.Range.Max.ValueFloat64Pointer(),
			},
			ErrorMessage: v.Message.ValueString(),
		}
	}

	if v.AssetFileSize != nil {
		return contentful.FieldValidationFileSize{
			Size: &contentful.MinMax{
				Min: v.AssetFileSize.Min.ValueFloat64Pointer(),
				Max: v.AssetFileSize.Max.ValueFloat64Pointer(),
			},
		}
	}

	if v.Regexp != nil {
		return contentful.FieldValidationRegex{
			Regex: &contentful.Regex{
				Pattern: v.Regexp.Pattern.ValueString(),
			},
			ErrorMessage: v.Message.ValueStringPointer(),
		}
	}

	if len(v.LinkContentType) > 0 {
		return contentful.FieldValidationLink{
			LinkContentType: pie.Map(v.LinkContentType, func(t types.String) string {
				return t.ValueString()
			}),
		}
	}

	if len(v.LinkMimetypeGroup) > 0 {
		return contentful.FieldValidationMimeType{
			MimeTypes: pie.Map(v.LinkMimetypeGroup, func(t types.String) string {
				return t.ValueString()
			}),
			ErrorMessage: v.Message.ValueStringPointer(),
		}
	}

	if len(v.In) > 0 {
		return contentful.FieldValidationPredefinedValues{
			In: pie.Map(v.In, func(t types.String) interface{} {
				return t.ValueString()
			}),
		}
	}

	if len(v.EnabledMarks) > 0 {
		return contentful.FieldValidationEnabledMarks{
			Marks: pie.Map(v.EnabledMarks, func(t types.String) string {
				return t.ValueString()
			}),
			ErrorMessage: v.Message.ValueStringPointer(),
		}
	}

	if len(v.EnabledNodeTypes) > 0 {
		return contentful.FieldValidationEnabledNodeTypes{
			NodeTypes: pie.Map(v.EnabledNodeTypes, func(t types.String) string {
				return t.ValueString()
			}),
			ErrorMessage: v.Message.ValueStringPointer(),
		}
	}

	return nil
}

type Size struct {
	Min types.Float64 `tfsdk:"min"`
	Max types.Float64 `tfsdk:"max"`
}

type Regexp struct {
	Pattern types.String `tfsdk:"pattern"`
}

func (f *Field) Equal(n *contentful.Field) bool {

	if n.Type != f.Type.ValueString() {
		return false
	}

	if n.ID != f.Id.ValueString() {
		return false
	}

	if n.Name != f.Name.ValueString() {
		return false
	}

	if n.LinkType != f.LinkType.ValueString() {
		return false
	}

	if n.Required != f.Required.ValueBool() {
		return false
	}

	if n.Omitted != f.Omitted.ValueBool() {
		return false
	}

	if n.Disabled != f.Disabled.ValueBool() {
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

	if len(f.Validations) != len(n.Validations) {
		return false
	}

	for idx, validation := range pie.Map(f.Validations, func(t Validation) contentful.FieldValidation {
		return t.Draft()
	}) {
		cfVal := n.Validations[idx]

		if !reflect.DeepEqual(validation, cfVal) {
			return false
		}

	}

	if f.DefaultValue != nil && !reflect.DeepEqual(f.DefaultValue.Draft(), n.DefaultValue) {
		return false
	}

	return true
}

func (f *Field) ToNative() (*contentful.Field, error) {

	contentfulField := &contentful.Field{
		ID:        f.Id.ValueString(),
		Name:      f.Name.ValueString(),
		Type:      f.Type.ValueString(),
		Localized: f.Localized.ValueBool(),
		Required:  f.Required.ValueBool(),
		Disabled:  f.Disabled.ValueBool(),
		Omitted:   f.Omitted.ValueBool(),
		Validations: pie.Map(f.Validations, func(t Validation) contentful.FieldValidation {
			return t.Draft()
		}),
	}

	if !f.LinkType.IsNull() && !f.LinkType.IsUnknown() {
		contentfulField.LinkType = f.LinkType.ValueString()
	}

	if contentfulField.Type == contentful.FieldTypeArray {
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

func getTypeOfMap(mapValues map[string]any) (*string, error) {
	for _, v := range mapValues {
		switch c := v.(type) {
		case string:
			t := "string"
			return &t, nil
		case bool:
			t := "bool"
			return &t, nil
		default:
			return nil, fmt.Errorf("The default type %T is not supported by the provider", c)
		}
	}

	return nil, nil
}

func (f *Field) Import(n *contentful.Field, c []contentful.Controls) error {
	f.Id = types.StringValue(n.ID)
	f.Name = types.StringValue(n.Name)
	f.Type = types.StringValue(n.Type)
	f.LinkType = utils.FromOptionalString(n.LinkType)
	f.Required = types.BoolValue(n.Required)
	f.Omitted = types.BoolValue(n.Omitted)
	f.Localized = types.BoolValue(n.Localized)
	f.Disabled = types.BoolValue(n.Disabled)

	defaultValueType, err := getTypeOfMap(n.DefaultValue)

	if err != nil {
		return err
	}

	if defaultValueType != nil {

		f.DefaultValue = &DefaultValue{
			Bool:   types.MapNull(types.BoolType),
			String: types.MapNull(types.StringType),
		}

		switch *defaultValueType {
		case "string":
			stringMap := map[string]attr.Value{}

			for k, v := range n.DefaultValue {
				stringMap[k] = types.StringValue(v.(string))
			}

			f.DefaultValue.String = types.MapValueMust(types.StringType, stringMap)
		case "bool":
			boolMap := map[string]attr.Value{}

			for k, v := range n.DefaultValue {
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

	if n.Type == contentful.FieldTypeArray {

		itemValidations, err := getValidations(n.Items.Validations)

		if err != nil {
			return err
		}

		f.Items = &Items{
			Type:        types.StringValue(n.Items.Type),
			LinkType:    types.StringPointerValue(n.Items.LinkType),
			Validations: itemValidations,
		}
	}

	idx := pie.FindFirstUsing(c, func(control contentful.Controls) bool {
		return n.ID == control.FieldID
	})

	if idx != -1 && c[idx].WidgetID != nil {

		var settings *Settings

		if c[idx].Settings != nil {
			settings = &Settings{}

			settings.Import(c[idx].Settings)
		}

		f.Control = &Control{
			WidgetId:        types.StringPointerValue(c[idx].WidgetID),
			WidgetNamespace: types.StringPointerValue(c[idx].WidgetNameSpace),
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

func (s *Settings) Import(settings *contentful.Settings) {
	s.HelpText = types.StringPointerValue(settings.HelpText)
	s.TrueLabel = types.StringPointerValue(settings.TrueLabel)
	s.FalseLabel = types.StringPointerValue(settings.FalseLabel)
	s.Stars = types.Int64PointerValue(settings.Stars)
	s.Format = types.StringPointerValue(settings.Format)
	s.TimeFormat = types.StringPointerValue(settings.AMPM)
	s.BulkEditing = types.BoolPointerValue(settings.BulkEditing)
	s.TrackingFieldId = types.StringPointerValue(settings.TrackingFieldId)
}

func (s *Settings) Draft() *contentful.Settings {
	settings := &contentful.Settings{}

	settings.HelpText = s.HelpText.ValueStringPointer()
	settings.TrueLabel = s.TrueLabel.ValueStringPointer()
	settings.FalseLabel = s.FalseLabel.ValueStringPointer()
	settings.Stars = s.Stars.ValueInt64Pointer()
	settings.Format = s.Format.ValueStringPointer()
	settings.AMPM = s.TimeFormat.ValueStringPointer()
	settings.BulkEditing = s.BulkEditing.ValueBoolPointer()
	settings.TrackingFieldId = s.TrackingFieldId.ValueStringPointer()
	return settings
}

type Items struct {
	Type        types.String `tfsdk:"type"`
	LinkType    types.String `tfsdk:"link_type"`
	Validations []Validation `tfsdk:"validations"`
}

func (i *Items) ToNative() (*contentful.FieldTypeArrayItem, error) {

	return &contentful.FieldTypeArrayItem{
		Type: i.Type.ValueString(),
		Validations: pie.Map(i.Validations, func(t Validation) contentful.FieldValidation {
			return t.Draft()
		}),
		LinkType: i.LinkType.ValueStringPointer(),
	}, nil
}

func (i *Items) Equal(n *contentful.FieldTypeArrayItem) bool {

	if n == nil {
		return false
	}

	if i.Type.ValueString() != n.Type {
		return false
	}

	if !utils.CompareStringPointer(i.LinkType, n.LinkType) {
		return false
	}

	if len(i.Validations) != len(n.Validations) {
		return false
	}

	for idx, validation := range pie.Map(i.Validations, func(t Validation) contentful.FieldValidation {
		return t.Draft()
	}) {
		cfVal := n.Validations[idx]

		if !reflect.DeepEqual(validation, cfVal) {
			return false
		}

	}

	return true
}

func (c *ContentType) Draft() (*contentful.ContentType, error) {

	var fields []*contentful.Field

	for _, field := range c.Fields {

		nativeField, err := field.ToNative()
		if err != nil {
			return nil, err
		}

		fields = append(fields, nativeField)
	}

	contentfulType := &contentful.ContentType{
		Name:         c.Name.ValueString(),
		DisplayField: c.DisplayField.ValueString(),
		Fields:       fields,
	}

	if !c.ID.IsUnknown() || !c.ID.IsNull() {
		contentfulType.Sys = &contentful.Sys{ID: c.ID.ValueString()}
	}

	if !c.Description.IsNull() && !c.Description.IsUnknown() {
		contentfulType.Description = c.Description.ValueStringPointer()
	}

	return contentfulType, nil

}

func (c *ContentType) Import(n *contentful.ContentType, e *contentful.EditorInterface) error {
	c.ID = types.StringValue(n.Sys.ID)
	c.Version = types.Int64Value(int64(n.Sys.Version))

	c.Description = types.StringPointerValue(n.Description)

	c.Name = types.StringValue(n.Name)
	c.DisplayField = types.StringValue(n.DisplayField)

	var fields []Field

	var controls []contentful.Controls

	if e != nil {
		controls = e.Controls
		c.VersionControls = types.Int64Value(int64(e.Sys.Version))
	}

	for _, nf := range n.Fields {
		field := &Field{}
		err := field.Import(nf, controls)
		if err != nil {
			return err
		}
		fields = append(fields, *field)
	}

	c.Fields = fields

	return nil

}

func (c *ContentType) Equal(n *contentful.ContentType) bool {

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
		idx := pie.FindFirstUsing(n.Fields, func(f *contentful.Field) bool {
			return f.ID == field.Id.ValueString()
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

func (c *ContentType) EqualControls(n []contentful.Controls) bool {

	if len(c.Fields) != len(n) {
		return false
	}

	filteredControls := pie.Filter(n, func(c contentful.Controls) bool {
		return c.WidgetID != nil || c.WidgetNameSpace != nil || c.Settings != nil
	})

	filteredFields := pie.Filter(c.Fields, func(f Field) bool {
		return f.Control != nil
	})

	if len(filteredControls) != len(filteredFields) {
		return false
	}

	for _, field := range filteredFields {
		idx := pie.FindFirstUsing(filteredControls, func(t contentful.Controls) bool {
			return t.FieldID == field.Id.ValueString()
		})

		if idx == -1 {
			return false
		}
		control := filteredControls[idx]

		if field.Control.WidgetId.ValueString() != *control.WidgetID {
			return false
		}

		if field.Control.WidgetNamespace.ValueString() != *control.WidgetNameSpace {
			return false
		}

		if field.Control.Settings == nil && control.Settings != nil {
			return false
		}

		if field.Control.Settings != nil && !reflect.DeepEqual(field.Control.Settings.Draft(), control.Settings) {
			return false
		}
	}

	return true
}

func (c *ContentType) DraftControls() []contentful.Controls {
	return pie.Map(c.Fields, func(field Field) contentful.Controls {

		control := contentful.Controls{
			FieldID: field.Id.ValueString(),
		}

		if field.Control != nil {
			control.WidgetID = field.Control.WidgetId.ValueStringPointer()
			control.WidgetNameSpace = field.Control.WidgetNamespace.ValueStringPointer()

			if field.Control.Settings != nil {
				control.Settings = field.Control.Settings.Draft()
			}
		}

		return control

	})
}

func getValidations(contentfulValidations []contentful.FieldValidation) ([]Validation, error) {
	var validations []Validation

	for _, validation := range contentfulValidations {

		val, err := getValidation(validation)

		if err != nil {
			return nil, err
		}

		validations = append(validations, *val)
	}

	return validations, nil
}

func getValidation(cfVal contentful.FieldValidation) (*Validation, error) {

	if v, ok := cfVal.(contentful.FieldValidationPredefinedValues); ok {

		return &Validation{
			In: pie.Map(v.In, func(t any) types.String {
				return types.StringValue(t.(string))
			}),
		}, nil
	}

	if v, ok := cfVal.(contentful.FieldValidationUnique); ok {

		return &Validation{
			Unique: types.BoolValue(v.Unique),
		}, nil
	}

	if v, ok := cfVal.(contentful.FieldValidationRegex); ok {

		return &Validation{
			Regexp: &Regexp{
				Pattern: types.StringValue(v.Regex.Pattern),
			},
			Message: types.StringPointerValue(v.ErrorMessage),
		}, nil
	}

	if v, ok := cfVal.(contentful.FieldValidationSize); ok {

		return &Validation{
			Size: &Size{
				Max: types.Float64PointerValue(v.Size.Max),
				Min: types.Float64PointerValue(v.Size.Min),
			},
		}, nil
	}

	if v, ok := cfVal.(contentful.FieldValidationLink); ok {

		return &Validation{
			LinkContentType: pie.Map(v.LinkContentType, func(t string) types.String {
				return types.StringValue(t)
			}),
		}, nil
	}

	if v, ok := cfVal.(contentful.FieldValidationMimeType); ok {

		return &Validation{
			LinkMimetypeGroup: pie.Map(v.MimeTypes, func(t string) types.String {
				return types.StringValue(t)
			}),
			Message: types.StringPointerValue(v.ErrorMessage),
		}, nil
	}

	if v, ok := cfVal.(contentful.FieldValidationRange); ok {

		return &Validation{
			Range: &Size{
				Max: types.Float64PointerValue(v.Range.Max),
				Min: types.Float64PointerValue(v.Range.Min),
			},
		}, nil
	}

	if v, ok := cfVal.(contentful.FieldValidationEnabledNodeTypes); ok {

		return &Validation{
			EnabledNodeTypes: pie.Map(v.NodeTypes, func(t string) types.String {
				return types.StringValue(t)
			}),
			Message: types.StringPointerValue(v.ErrorMessage),
		}, nil
	}

	if v, ok := cfVal.(contentful.FieldValidationEnabledMarks); ok {

		return &Validation{
			EnabledMarks: pie.Map(v.Marks, func(t string) types.String {
				return types.StringValue(t)
			}),
			Message: types.StringPointerValue(v.ErrorMessage),
		}, nil
	}

	if v, ok := cfVal.(contentful.FieldValidationFileSize); ok {
		return &Validation{
			AssetFileSize: &Size{
				Max: types.Float64PointerValue(v.Size.Max),
				Min: types.Float64PointerValue(v.Size.Min),
			},
		}, nil
	}

	return nil, fmt.Errorf("Unsupported validation used, %s. Please implement", reflect.TypeOf(cfVal).String())
}

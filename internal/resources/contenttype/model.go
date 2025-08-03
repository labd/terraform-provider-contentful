package contenttype

import (
	"fmt"
	"reflect"

	"github.com/elliotchance/pie/v2"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/labd/terraform-provider-contentful/internal/sdk"
	"github.com/labd/terraform-provider-contentful/internal/utils"
)

// ContentType is the main resource schema data
type ContentType struct {
	ID           types.String `tfsdk:"id"`
	SpaceId      types.String `tfsdk:"space_id"`
	Environment  types.String `tfsdk:"environment"`
	Name         types.String `tfsdk:"name"`
	DisplayField types.String `tfsdk:"display_field"`
	Description  types.String `tfsdk:"description"`
	Version      types.Int64  `tfsdk:"version"`
	Fields       []Field      `tfsdk:"fields"`
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
	DefaultValue *DefaultValue `tfsdk:"default_value"`
}

type DefaultValue struct {
	Bool   types.Map `tfsdk:"bool"`
	String types.Map `tfsdk:"string"`
}

// HasContent checks if the DefaultValue has any actual content
func (d *DefaultValue) HasContent() bool {
	if d == nil {
		return false
	}
	
	// Check if String map has content
	if !d.String.IsNull() && !d.String.IsUnknown() && len(d.String.Elements()) > 0 {
		return true
	}
	
	// Check if Bool map has content
	if !d.Bool.IsNull() && !d.Bool.IsUnknown() && len(d.Bool.Elements()) > 0 {
		return true
	}
	
	return false
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
	Nodes             *Nodes         `tfsdk:"nodes"`
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

	if v.LinkContentType != nil {
		value := pie.Map(v.LinkContentType, func(t types.String) string {
			return t.ValueString()
		})
		base.LinkContentType = &value
		return base, nil
	}

	if v.LinkMimetypeGroup != nil {
		value := pie.Map(v.LinkMimetypeGroup, func(t types.String) string {
			return t.ValueString()
		})
		base.LinkMimetypeGroup = &value
		return base, nil
	}

	if v.In != nil {
		value := pie.Map(v.In, func(t types.String) string {
			return t.ValueString()
		})
		base.In = &value
		return base, nil
	}

	if v.EnabledMarks != nil {
		value := pie.Map(v.EnabledMarks, func(t types.String) string {
			return t.ValueString()
		})
		base.EnabledMarks = &value
		return base, nil
	}

	if v.EnabledNodeTypes != nil {
		value := pie.Map(v.EnabledNodeTypes, func(t types.String) string {
			return t.ValueString()
		})
		base.EnabledNodeTypes = &value
		return base, nil
	}

	if v.Nodes != nil {
		value := v.Nodes.Draft()
		base.Nodes = value
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

type Nodes struct {
	AssetHyperlink         []AssetHyperlinkValidation        `tfsdk:"asset_hyperlink"`
	EntryHyperlink         []EntryHyperlinkValidation        `tfsdk:"entry_hyperlink"`
	EmbeddedAssetBlock     []EmbeddedAssetBlockValidation    `tfsdk:"embedded_asset_block"`
	EmbeddedEntryBlock     []EmbeddedEntryBlockValidation    `tfsdk:"embedded_entry_block"`
	EmbeddedEntryInline    []EmbeddedEntryInlineValidation   `tfsdk:"embedded_entry_inline"`
	EmbeddedResourceBlock  *EmbeddedResourceBlockValidation  `tfsdk:"embedded_resource_block"`
	EmbeddedResourceInline *EmbeddedResourceInlineValidation `tfsdk:"embedded_resource_inline"`
	ResourceHyperlink      *ResourceHyperlinkValidation      `tfsdk:"resource_hyperlink"`
}

type AssetHyperlinkValidation struct {
	Message types.String `tfsdk:"message"`
	Size    *Size        `tfsdk:"size"`
}

type EntryHyperlinkValidation struct {
	Message         types.String    `tfsdk:"message"`
	Size            *Size           `tfsdk:"size"`
	LinkContentType *[]types.String `tfsdk:"link_content_type"`
}

type EmbeddedAssetBlockValidation struct {
	Message types.String `tfsdk:"message"`
	Size    *Size        `tfsdk:"size"`
}

type EmbeddedEntryBlockValidation struct {
	Message         types.String    `tfsdk:"message"`
	Size            *Size           `tfsdk:"size"`
	LinkContentType *[]types.String `tfsdk:"link_content_type"`
}

type EmbeddedEntryInlineValidation struct {
	Message         types.String    `tfsdk:"message"`
	Size            *Size           `tfsdk:"size"`
	LinkContentType *[]types.String `tfsdk:"link_content_type"`
}

type EmbeddedResourceBlockValidation struct {
	Validations      *[]Validations     `tfsdk:"validations"`
	AllowedResources *[]AllowedResource `tfsdk:"allowed_resources"`
}

type EmbeddedResourceInlineValidation struct {
	Validations      *[]Validations     `tfsdk:"validations"`
	AllowedResources *[]AllowedResource `tfsdk:"allowed_resources"`
}

type ResourceHyperlinkValidation struct {
	Validations      *[]Validations     `tfsdk:"validations"`
	AllowedResources *[]AllowedResource `tfsdk:"allowed_resources"`
}

type Validations struct {
	Message types.String `tfsdk:"message"`
	Size    *Size        `tfsdk:"size"`
}

type AllowedResource struct {
	Type         types.String   `tfsdk:"type"`
	Source       types.String   `tfsdk:"source"`
	ContentTypes []types.String `tfsdk:"content_types"`
}

func (n *Nodes) Draft() *sdk.NodesValidation {
	base := &sdk.NodesValidation{}

	if n.AssetHyperlink != nil {
		value := pie.Map(n.AssetHyperlink, func(v AssetHyperlinkValidation) sdk.AssetHyperlinkValidation {
			var size *sdk.RangeMinMax
			if v.Size != nil {
				size = &sdk.RangeMinMax{
					Min: v.Size.Min.ValueFloat64Pointer(),
					Max: v.Size.Max.ValueFloat64Pointer(),
				}
			}

			return sdk.AssetHyperlinkValidation{
				Message: v.Message.ValueStringPointer(),
				Size:    size,
			}
		})
		base.AssetHyperlink = &value
	}

	if n.EntryHyperlink != nil {
		value := pie.Map(n.EntryHyperlink, func(v EntryHyperlinkValidation) sdk.EntryHyperlinkValidation {
			var linkContentType *[]string
			if v.LinkContentType != nil {
				v := pie.Map(*v.LinkContentType, func(t types.String) string { return t.ValueString() })
				linkContentType = &v
			}

			var size *sdk.RangeMinMax
			if v.Size != nil {
				size = &sdk.RangeMinMax{
					Min: v.Size.Min.ValueFloat64Pointer(),
					Max: v.Size.Max.ValueFloat64Pointer(),
				}
			}
			return sdk.EntryHyperlinkValidation{
				Message:         v.Message.ValueStringPointer(),
				LinkContentType: linkContentType,
				Size:            size,
			}
		})
		base.EntryHyperlink = &value
	}

	if n.EmbeddedAssetBlock != nil {
		value := pie.Map(n.EmbeddedAssetBlock, func(v EmbeddedAssetBlockValidation) sdk.EmbeddedAssetBlockValidation {
			var size *sdk.RangeMinMax
			if v.Size != nil {
				size = &sdk.RangeMinMax{
					Min: v.Size.Min.ValueFloat64Pointer(),
					Max: v.Size.Max.ValueFloat64Pointer(),
				}
			}
			return sdk.EmbeddedAssetBlockValidation{
				Message: v.Message.ValueStringPointer(),
				Size:    size,
			}
		})
		base.EmbeddedAssetBlock = &value
	}

	if n.EmbeddedEntryBlock != nil {
		value := pie.Map(n.EmbeddedEntryBlock, func(v EmbeddedEntryBlockValidation) sdk.EmbeddedEntryBlockValidation {
			var linkContentType *[]string
			if v.LinkContentType != nil {
				v := pie.Map(*v.LinkContentType, func(t types.String) string { return t.ValueString() })
				linkContentType = &v
			}
			var size *sdk.RangeMinMax
			if v.Size != nil {
				size = &sdk.RangeMinMax{
					Min: v.Size.Min.ValueFloat64Pointer(),
					Max: v.Size.Max.ValueFloat64Pointer(),
				}
			}

			return sdk.EmbeddedEntryBlockValidation{
				Message:         v.Message.ValueStringPointer(),
				LinkContentType: linkContentType,
				Size:            size,
			}
		})
		base.EmbeddedEntryBlock = &value
	}

	if n.EmbeddedEntryInline != nil {
		value := pie.Map(n.EmbeddedEntryInline, func(v EmbeddedEntryInlineValidation) sdk.EmbeddedEntryInlineValidation {
			var linkContentType *[]string
			if v.LinkContentType != nil {
				v := pie.Map(*v.LinkContentType, func(t types.String) string { return t.ValueString() })
				linkContentType = &v
			}
			var size *sdk.RangeMinMax
			if v.Size != nil {
				size = &sdk.RangeMinMax{
					Min: v.Size.Min.ValueFloat64Pointer(),
					Max: v.Size.Max.ValueFloat64Pointer(),
				}
			}

			return sdk.EmbeddedEntryInlineValidation{
				Message:         v.Message.ValueStringPointer(),
				LinkContentType: linkContentType,
				Size:            size,
			}
		})
		base.EmbeddedEntryInline = &value
	}

	if n.EmbeddedResourceBlock != nil {
		er := &sdk.EmbeddedResourceBlockValidation{}

		if n.EmbeddedResourceBlock.Validations != nil {
			v := pie.Map(*n.EmbeddedResourceBlock.Validations, func(v Validations) sdk.EmbeddedValidation {
				var size *sdk.RangeMinMax
				if v.Size != nil {
					size = &sdk.RangeMinMax{
						Min: v.Size.Min.ValueFloat64Pointer(),
						Max: v.Size.Max.ValueFloat64Pointer(),
					}
				}

				return sdk.EmbeddedValidation{
					Message: v.Message.ValueStringPointer(),
					Size:    size,
				}
			})

			er.Validations = &v
		}

		if n.EmbeddedResourceBlock.AllowedResources != nil {
			ar := pie.Map(*n.EmbeddedResourceBlock.AllowedResources, func(v AllowedResource) sdk.AllowedResource {
				return MapInternalAllowedResource(v)
			})
			er.AllowedResources = &ar
		}

		base.EmbeddedResourceBlock = er
	}

	if n.EmbeddedResourceInline != nil {
		ri := &sdk.EmbeddedResourceInlineValidation{}

		if n.EmbeddedResourceInline.Validations != nil {
			v := pie.Map(*n.EmbeddedResourceInline.Validations, func(v Validations) sdk.EmbeddedValidation {
				var size *sdk.RangeMinMax
				if v.Size != nil {
					size = &sdk.RangeMinMax{
						Min: v.Size.Min.ValueFloat64Pointer(),
						Max: v.Size.Max.ValueFloat64Pointer(),
					}
				}

				return sdk.EmbeddedValidation{
					Message: v.Message.ValueStringPointer(),
					Size:    size,
				}
			})
			ri.Validations = &v

		}

		if n.EmbeddedResourceInline.AllowedResources != nil {
			ar := pie.Map(*n.EmbeddedResourceInline.AllowedResources, func(v AllowedResource) sdk.AllowedResource {
				return MapInternalAllowedResource(v)
			})
			ri.AllowedResources = &ar
		}

		base.EmbeddedResourceInline = ri
	}

	if n.ResourceHyperlink != nil {
		rl := &sdk.ResourceHyperlinkValidation{}

		if n.ResourceHyperlink.Validations != nil {
			v := pie.Map(*n.ResourceHyperlink.Validations, func(v Validations) sdk.EmbeddedValidation {
				var size *sdk.RangeMinMax
				if v.Size != nil {
					size = &sdk.RangeMinMax{
						Min: v.Size.Min.ValueFloat64Pointer(),
						Max: v.Size.Max.ValueFloat64Pointer(),
					}
				}

				return sdk.EmbeddedValidation{
					Message: v.Message.ValueStringPointer(),
					Size:    size,
				}
			})
			rl.Validations = &v

		}

		if n.ResourceHyperlink.AllowedResources != nil {
			ar := pie.Map(*n.ResourceHyperlink.AllowedResources, func(v AllowedResource) sdk.AllowedResource {
				return MapInternalAllowedResource(v)
			})
			rl.AllowedResources = &ar
		}

		base.ResourceHyperlink = rl
	}

	return base
}

func MapInternalAllowedResource(allowedResource AllowedResource) sdk.AllowedResource {
	ar := sdk.AllowedResource{
		Type:   allowedResource.Type.ValueStringPointer(),
		Source: allowedResource.Source.ValueStringPointer(),
	}

	var contentTypes = pie.Map(allowedResource.ContentTypes, func(t types.String) string {
		return t.ValueString()
	})

	if len(contentTypes) != 0 {
		ar.ContentTypes = &contentTypes
	}

	return ar
}

func MapSdkAllowedResource(allowedResource sdk.AllowedResource) AllowedResource {
	var contentTypes = pie.Map(*allowedResource.ContentTypes, func(t string) types.String {
		return types.StringValue(t)
	})

	return AllowedResource{
		Type:         types.StringPointerValue(allowedResource.Type),
		Source:       types.StringPointerValue(allowedResource.Source),
		ContentTypes: contentTypes,
	}
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

	if f.DefaultValue != nil && f.DefaultValue.HasContent() {
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

func (f *Field) Import(n sdk.Field) error {
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

			var linkType = types.StringNull()
			if linkItem.LinkType != "" {
				linkType = types.StringValue(string(linkItem.LinkType))
			}

			f.Items = &Items{
				Type:        types.StringValue(itemType),
				LinkType:    linkType,
				Validations: itemValidations,
			}
		}
	}

	return nil
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
		DisplayField: c.DisplayField.ValueStringPointer(),
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
		DisplayField: c.DisplayField.ValueStringPointer(),
		Fields:       fields,
	}

	if !c.Description.IsNull() && !c.Description.IsUnknown() {
		contentfulType.Description = c.Description.ValueStringPointer()
	}

	return contentfulType, nil
}

func (c *ContentType) Import(n *sdk.ContentType) error {
	c.ID = types.StringValue(n.Sys.Id)
	c.Version = types.Int64Value(n.Sys.Version)

	c.Description = types.StringPointerValue(n.Description)

	c.Name = types.StringValue(n.Name)
	c.DisplayField = types.StringPointerValue(n.DisplayField)

	var fields []Field

	for _, nf := range n.Fields {
		field := &Field{}
		err := field.Import(nf)
		if err != nil {
			return fmt.Errorf("field import failed: %w", err)
		}
		fields = append(fields, *field)
	}

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

	if c.DisplayField.ValueStringPointer() != n.DisplayField {
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
			AssetFileSize: &Size{
				Max: types.Float64PointerValue(cfVal.AssetFileSize.Max),
				Min: types.Float64PointerValue(cfVal.AssetFileSize.Min),
			},
			Message: types.StringPointerValue(cfVal.Message),
		}, nil
	}

	if cfVal.Range != nil {
		return &Validation{
			Range: &Size{
				Max: types.Float64PointerValue(cfVal.Range.Max),
				Min: types.Float64PointerValue(cfVal.Range.Min),
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

	if cfVal.Unique != nil {
		return &Validation{
			Unique:  types.BoolPointerValue(cfVal.Unique),
			Message: types.StringPointerValue(cfVal.Message),
		}, nil
	}

	if cfVal.Size != nil {
		return &Validation{
			Size: &Size{
				Max: types.Float64PointerValue(cfVal.Size.Max),
				Min: types.Float64PointerValue(cfVal.Size.Min),
			},
			Message: types.StringPointerValue(cfVal.Message),
		}, nil
	}

	if cfVal.Nodes != nil {
		nodes := getNodesValidation(*cfVal.Nodes)

		return &Validation{
			Nodes:   &nodes,
			Message: types.StringPointerValue(cfVal.Message),
		}, nil
	}

	return nil, fmt.Errorf("unsupported validation used, %s. Please implement", reflect.TypeOf(cfVal).String())
}

func getNodesValidation(cfVal sdk.NodesValidation) Nodes {
	nodes := Nodes{}

	if cfVal.AssetHyperlink != nil {
		nodes.AssetHyperlink = pie.Map(*cfVal.AssetHyperlink, func(v sdk.AssetHyperlinkValidation) AssetHyperlinkValidation {
			var size *Size
			if v.Size != nil {
				size = &Size{
					Min: types.Float64PointerValue(v.Size.Min),
					Max: types.Float64PointerValue(v.Size.Max),
				}
			}
			return AssetHyperlinkValidation{
				Message: types.StringPointerValue(v.Message),
				Size:    size,
			}
		})
	}

	if cfVal.EntryHyperlink != nil {
		nodes.EntryHyperlink = pie.Map(*cfVal.EntryHyperlink, func(v sdk.EntryHyperlinkValidation) EntryHyperlinkValidation {
			var size *Size
			if v.Size != nil {
				size = &Size{
					Min: types.Float64PointerValue(v.Size.Min),
					Max: types.Float64PointerValue(v.Size.Max),
				}
			}

			var linkContentType []types.String
			if v.LinkContentType != nil {
				linkContentType = pie.Map(*v.LinkContentType, func(t string) types.String {
					return types.StringValue(t)
				})
			}

			return EntryHyperlinkValidation{
				Message:         types.StringPointerValue(v.Message),
				Size:            size,
				LinkContentType: &linkContentType,
			}
		})
	}

	if cfVal.EmbeddedAssetBlock != nil {
		nodes.EmbeddedAssetBlock = pie.Map(*cfVal.EmbeddedAssetBlock, func(v sdk.EmbeddedAssetBlockValidation) EmbeddedAssetBlockValidation {
			var size *Size
			if v.Size != nil {
				size = &Size{
					Min: types.Float64PointerValue(v.Size.Min),
					Max: types.Float64PointerValue(v.Size.Max),
				}
			}
			return EmbeddedAssetBlockValidation{
				Message: types.StringPointerValue(v.Message),
				Size:    size,
			}
		})
	}

	if cfVal.EmbeddedEntryBlock != nil {
		nodes.EmbeddedEntryBlock = pie.Map(*cfVal.EmbeddedEntryBlock, func(v sdk.EmbeddedEntryBlockValidation) EmbeddedEntryBlockValidation {
			var size *Size
			if v.Size != nil {
				size = &Size{
					Min: types.Float64PointerValue(v.Size.Min),
					Max: types.Float64PointerValue(v.Size.Max),
				}
			}

			var linkContentType []types.String
			if v.LinkContentType != nil {
				linkContentType = pie.Map(*v.LinkContentType, func(t string) types.String {
					return types.StringValue(t)
				})
			}

			return EmbeddedEntryBlockValidation{
				Message:         types.StringPointerValue(v.Message),
				Size:            size,
				LinkContentType: &linkContentType,
			}
		})
	}

	if cfVal.EmbeddedEntryInline != nil {
		nodes.EmbeddedEntryInline = pie.Map(*cfVal.EmbeddedEntryInline, func(v sdk.EmbeddedEntryInlineValidation) EmbeddedEntryInlineValidation {
			var size *Size
			if v.Size != nil {
				size = &Size{
					Min: types.Float64PointerValue(v.Size.Min),
					Max: types.Float64PointerValue(v.Size.Max),
				}
			}

			var linkContentType []types.String
			if v.LinkContentType != nil {
				linkContentType = pie.Map(*v.LinkContentType, func(t string) types.String {
					return types.StringValue(t)
				})
			}

			return EmbeddedEntryInlineValidation{
				Message:         types.StringPointerValue(v.Message),
				Size:            size,
				LinkContentType: &linkContentType,
			}
		})
	}

	if cfVal.EmbeddedResourceBlock != nil {
		var allowedResources []AllowedResource
		if cfVal.EmbeddedResourceBlock.AllowedResources != nil {
			allowedResources = pie.Map(*cfVal.EmbeddedResourceBlock.AllowedResources, MapSdkAllowedResource)
		}

		var validations = pie.Map(*cfVal.EmbeddedResourceBlock.Validations, func(v sdk.EmbeddedValidation) Validations {
			var size *Size
			if v.Size != nil {
				size = &Size{
					Min: types.Float64PointerValue(v.Size.Min),
					Max: types.Float64PointerValue(v.Size.Max),
				}
			}

			return Validations{
				Message: types.StringPointerValue(v.Message),
				Size:    size,
			}
		})

		nodes.EmbeddedResourceBlock = &EmbeddedResourceBlockValidation{
			AllowedResources: &allowedResources,
			Validations:      &validations,
		}
	}

	if cfVal.EmbeddedResourceInline != nil {
		var allowedResources []AllowedResource
		if cfVal.EmbeddedResourceInline.AllowedResources != nil {
			allowedResources = pie.Map(*cfVal.EmbeddedResourceInline.AllowedResources, MapSdkAllowedResource)
		}

		var validations = pie.Map(*cfVal.EmbeddedResourceInline.Validations, func(v sdk.EmbeddedValidation) Validations {
			var size *Size
			if v.Size != nil {
				size = &Size{
					Min: types.Float64PointerValue(v.Size.Min),
					Max: types.Float64PointerValue(v.Size.Max),
				}
			}

			return Validations{
				Message: types.StringPointerValue(v.Message),
				Size:    size,
			}
		})

		nodes.EmbeddedResourceInline = &EmbeddedResourceInlineValidation{
			AllowedResources: &allowedResources,
			Validations:      &validations,
		}
	}

	if cfVal.ResourceHyperlink != nil {

		var allowedResources []AllowedResource
		if cfVal.ResourceHyperlink.AllowedResources != nil {
			allowedResources = pie.Map(*cfVal.ResourceHyperlink.AllowedResources, MapSdkAllowedResource)
		}

		var validations = pie.Map(*cfVal.ResourceHyperlink.Validations, func(v sdk.EmbeddedValidation) Validations {
			var size *Size
			if v.Size != nil {
				size = &Size{
					Min: types.Float64PointerValue(v.Size.Min),
					Max: types.Float64PointerValue(v.Size.Max),
				}
			}

			return Validations{
				Message: types.StringPointerValue(v.Message),
				Size:    size,
			}
		})

		nodes.ResourceHyperlink = &ResourceHyperlinkValidation{
			AllowedResources: &allowedResources,
			Validations:      &validations,
		}
	}

	return nodes
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

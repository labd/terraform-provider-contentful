package entry

import (
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/iancoleman/orderedmap"

	"github.com/labd/terraform-provider-contentful/internal/sdk"
	"github.com/labd/terraform-provider-contentful/internal/utils"
)

// Entry is the main resource schema data
type Entry struct {
	ID            types.String `tfsdk:"id"`
	EntryID       types.String `tfsdk:"entry_id"`
	Version       types.Int64  `tfsdk:"version"`
	SpaceID       types.String `tfsdk:"space_id"`
	Environment   types.String `tfsdk:"environment"`
	ContentTypeID types.String `tfsdk:"contenttype_id"`
	Field         []Field      `tfsdk:"field"`
	Published     types.Bool   `tfsdk:"published"`
	Archived      types.Bool   `tfsdk:"archived"`
}

// Field represents a content field in an Entry
type Field struct {
	ID      types.String `tfsdk:"id"`
	Content types.String `tfsdk:"content"`
	Locale  types.String `tfsdk:"locale"`
}

// Import populates the Entry struct from an SDK entry object
func (e *Entry) Import(entry *sdk.Entry) {
	e.ID = types.StringValue(entry.Sys.Id)
	e.EntryID = types.StringValue(entry.Sys.Id)
	e.Version = types.Int64Value(entry.Sys.Version)
	e.SpaceID = types.StringValue(entry.Sys.Space.Sys.Id)
	e.Environment = types.StringValue(entry.Sys.Environment.Sys.Id)
	e.ContentTypeID = types.StringValue(entry.Sys.ContentType.Sys.Id)
	e.Published = types.BoolValue(entry.Sys.PublishedAt != nil)
	e.Archived = types.BoolValue(entry.Sys.ArchivedAt != nil)

	e.BuildFieldsFromAPIResponse(entry)
}

// DraftForCreate creates an EntryCreate object for creating a new entry
func (e *Entry) Draft() sdk.EntryDraft {
	fieldProperties := orderedmap.New()

	for _, field := range e.Field {
		fieldID := field.ID.ValueString()
		locale := field.Locale.ValueString()
		content := ParseContentValue(field.Content.ValueString())

		prop, ok := fieldProperties.Get(fieldID)
		if !ok {
			prop = map[string]any{}
			fieldProperties.Set(fieldID, prop)
		}

		prop.(map[string]any)[locale] = content
	}

	return sdk.EntryDraft{
		Fields: fieldProperties,
	}
}

// parseContentValue tries to parse a string as JSON, otherwise returns the original value
func ParseContentValue(value string) interface{} {
	var content any
	err := json.Unmarshal([]byte(value), &content)
	if err != nil {
		content = value
	}

	return utils.SortOrderedMapRecursively(content)
}

// BuildFieldsFromAPIResponse builds the Field array from API response
func (e *Entry) BuildFieldsFromAPIResponse(entry *sdk.Entry) {
	e.Field = []Field{}

	// If no fields are present in the response, return early
	if len(entry.Fields.Keys()) == 0 {
		return
	}

	// Convert fields from the API response to the Field structure
	fields := entry.Fields
	for _, fieldID := range fields.Keys() {
		fieldValue, _ := fields.Get(fieldID)

		subFields := fieldValue.(orderedmap.OrderedMap)

		for _, locale := range subFields.Keys() {
			content, _ := subFields.Get(locale)

			// Convert the content back to string representation for storage
			contentStr := ""
			switch v := content.(type) {
			case string:
				contentStr = v
			default:
				v = utils.SortOrderedMapRecursively(v)
				// Try to marshal complex types back to JSON string
				if jsonBytes, err := json.Marshal(v); err == nil {
					contentStr = string(jsonBytes)
				} else {
					contentStr = fmt.Sprintf("%v", v)
				}
			}

			e.Field = append(e.Field, Field{
				ID:      types.StringValue(fieldID),
				Locale:  types.StringValue(locale),
				Content: types.StringValue(contentStr),
			})
		}
	}
}

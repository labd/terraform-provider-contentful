package asset

import (
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/labd/terraform-provider-contentful/internal/sdk"
)

// Asset is the main resource schema data
type Asset struct {
	ID          types.String `tfsdk:"id"`
	AssetID     types.String `tfsdk:"asset_id"`
	Version     types.Int64  `tfsdk:"version"`
	SpaceID     types.String `tfsdk:"space_id"`
	Environment types.String `tfsdk:"environment"`
	Fields      *AssetFields `tfsdk:"fields"`
	Published   types.Bool   `tfsdk:"published"`
	Archived    types.Bool   `tfsdk:"archived"`
}

type AssetFields struct {
	Title       []LocalizedField    `tfsdk:"title"`
	Description []LocalizedField    `tfsdk:"description"`
	File        []LocalizedFileItem `tfsdk:"file"`
}

type LocalizedField struct {
	Content types.String `tfsdk:"content"`
	Locale  types.String `tfsdk:"locale"`
}

type LocalizedFileItem struct {
	Locale      types.String `tfsdk:"locale"`
	Upload      types.String `tfsdk:"upload"`
	URL         types.String `tfsdk:"url"`
	FileName    types.String `tfsdk:"file_name"`
	ContentType types.String `tfsdk:"content_type"`
	FileSize    types.Int64  `tfsdk:"filesize"`
	ImageWidth  types.Int64  `tfsdk:"image_width"`
	ImageHeight types.Int64  `tfsdk:"image_height"`
}

// Import populates the Asset struct from an SDK asset object
func (a *Asset) Import(asset *sdk.Asset) {
	a.ID = types.StringValue(asset.Sys.Id)
	a.AssetID = types.StringValue(asset.Sys.Id)
	a.Version = types.Int64Value(int64(asset.Sys.Version))
	a.SpaceID = types.StringValue(asset.Sys.Space.Sys.Id)
	a.Environment = types.StringValue(asset.Sys.Environment.Sys.Id)
	a.Published = types.BoolValue(asset.Sys.PublishedAt != nil)
	a.Archived = types.BoolValue(asset.Sys.ArchivedAt != nil)
	a.Version = types.Int64Value(asset.Sys.Version)

	// Import fields
	a.Fields = &AssetFields{
		Title:       []LocalizedField{},
		Description: []LocalizedField{},
		File:        []LocalizedFileItem{},
	}

	// Title
	for locale, content := range asset.Fields.Title {
		a.Fields.Title = append(a.Fields.Title, LocalizedField{
			Content: types.StringValue(content),
			Locale:  types.StringValue(locale),
		})
	}

	// Description
	for locale, content := range asset.Fields.Description {
		a.Fields.Description = append(a.Fields.Description, LocalizedField{
			Content: types.StringValue(content),
			Locale:  types.StringValue(locale),
		})
	}

	// File
	for locale, file := range asset.Fields.File {
		fileItem := LocalizedFileItem{
			Locale:      types.StringValue(locale),
			ContentType: types.StringValue(file.ContentType),
			Upload:      types.StringPointerValue(file.Upload),
			URL:         types.StringPointerValue(file.Url),
			FileName:    types.StringValue(file.FileName),
		}

		if file.Details != nil {
			fileItem.FileSize = types.Int64PointerValue(file.Details.Size)
			if file.Details.Image != nil {
				fileItem.ImageWidth = types.Int64PointerValue(file.Details.Image.Width)
				fileItem.ImageHeight = types.Int64PointerValue(file.Details.Image.Height)
			}
		}

		// Add to files collection
		a.Fields.File = append(a.Fields.File, fileItem)
	}
}

func (a *Asset) CopyInputValues(plan *Asset) {
	uploads := map[string]string{}
	for _, file := range plan.Fields.File {
		if !file.Upload.IsNull() {
			uploads[file.Locale.ValueString()] = file.Upload.ValueString()
		}
	}
	contentTypes := map[string]string{}
	for _, file := range plan.Fields.File {
		if !file.ContentType.IsNull() {
			contentTypes[file.Locale.ValueString()] = file.ContentType.ValueString()
		}
	}

	for i, file := range a.Fields.File {
		a.Fields.File[i].Upload = types.StringValue(uploads[file.Locale.ValueString()])
	}

	for i, file := range a.Fields.File {
		a.Fields.File[i].ContentType = types.StringValue(contentTypes[file.Locale.ValueString()])
	}
}

// DraftForCreate creates an AssetCreate object for the API
func (a *Asset) DraftForCreate() *sdk.AssetCreate {
	localizedTitle := map[string]string{}
	for _, item := range a.Fields.Title {
		localizedTitle[item.Locale.ValueString()] = item.Content.ValueString()
	}

	localizedDescription := map[string]string{}
	for _, item := range a.Fields.Description {
		localizedDescription[item.Locale.ValueString()] = item.Content.ValueString()
	}

	fileData := map[string]sdk.AssetFile{}
	for _, item := range a.Fields.File {
		key := item.Locale.ValueString()
		fileData[key] = sdk.AssetFile{
			Upload:      item.Upload.ValueString(),
			FileName:    item.FileName.ValueString(),
			ContentType: item.ContentType.ValueString(),
		}
	}

	return &sdk.AssetCreate{
		Fields: &sdk.AssetField{
			Title:       localizedTitle,
			Description: localizedDescription,
			File:        fileData,
		},
	}
}

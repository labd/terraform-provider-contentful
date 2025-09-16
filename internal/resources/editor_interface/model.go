package editor_interface

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/elliotchance/pie/v2"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/labd/terraform-provider-contentful/internal/sdk"
	"github.com/labd/terraform-provider-contentful/internal/utils"
)

// EditorInterface represents a Contentful Editor Interface
type EditorInterface struct {
	ID          types.String `tfsdk:"id"`
	SpaceID     types.String `tfsdk:"space_id"`
	Environment types.String `tfsdk:"environment"`
	ContentType types.String `tfsdk:"content_type"`
	Version     types.Int64  `tfsdk:"version"`
	Controls    []Control    `tfsdk:"controls"`
	Sidebar     []Sidebar    `tfsdk:"sidebar"`
	Editors     []Editor     `tfsdk:"editors"`
}

type Sidebar struct {
	WidgetId        types.String         `tfsdk:"widget_id"`
	WidgetNamespace types.String         `tfsdk:"widget_namespace"`
	Settings        jsontypes.Normalized `tfsdk:"settings"`
	Disabled        types.Bool           `tfsdk:"disabled"`
}

type Editor struct {
	WidgetId        types.String         `tfsdk:"widget_id"`
	WidgetNamespace types.String         `tfsdk:"widget_namespace"`
	Settings        jsontypes.Normalized `tfsdk:"settings"`
	Disabled        types.Bool           `tfsdk:"disabled"`
}

// Control represents a field control in the editor interface
type Control struct {
	FieldID         types.String `tfsdk:"field_id"`
	WidgetID        types.String `tfsdk:"widget_id"`
	WidgetNamespace types.String `tfsdk:"widget_namespace"`
	Settings        *Settings    `tfsdk:"settings"`
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

// ToUpdateBody converts the EditorInterface to an SDK update request body
func (e *EditorInterface) ToUpdateBody() sdk.EditorInterfaceUpdate {
	result := sdk.EditorInterfaceUpdate{}

	controls := make([]sdk.EditorInterfaceControl, 0, len(e.Controls))
	for _, control := range e.Controls {
		newControl := sdk.EditorInterfaceControl{
			FieldId:         control.FieldID.ValueString(),
			WidgetId:        control.WidgetID.ValueStringPointer(),
			WidgetNamespace: utils.Pointer(sdk.EditorInterfaceControlWidgetNamespaceBuiltin),
		}

		if !control.WidgetNamespace.IsNull() && !control.WidgetNamespace.IsUnknown() {
			newControl.WidgetNamespace = utils.Pointer(sdk.EditorInterfaceControlWidgetNamespace(control.WidgetNamespace.ValueString()))
		}

		if control.Settings != nil {
			newControl.Settings = control.Settings.Draft()
		}

		controls = append(controls, newControl)
	}
	result.Controls = controls

	sidebar := pie.Map(e.Sidebar, func(t Sidebar) sdk.EditorInterfaceSidebarItem {
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

	if sidebar != nil {
		result.Sidebar = &sidebar
	} else {
		result.Sidebar = nil
	}

	editors := pie.Map(e.Editors, func(t Editor) sdk.EditorInterfaceEditor {
		var namespace sdk.EditorInterfaceEditorWidgetNamespace
		if !t.WidgetNamespace.IsNull() {
			namespace = sdk.EditorInterfaceEditorWidgetNamespace(t.WidgetNamespace.ValueString())
		}

		editor := sdk.EditorInterfaceEditor{
			WidgetNamespace: &namespace,
			WidgetId:        t.WidgetId.ValueStringPointer(),
			Disabled:        t.Disabled.ValueBoolPointer(),
		}

		if !*editor.Disabled {
			settings := sdk.EditorInterfaceSettings{}

			t.Settings.Unmarshal(settings)
			editor.Settings = &settings
		}

		return editor
	})

	if editors != nil {
		result.Editors = &editors
	} else {
		result.Editors = nil
	}

	return result
}

// Import populates the EditorInterface from an SDK response
func (e *EditorInterface) Import(editorInterface *sdk.EditorInterface) {
	e.SpaceID = types.StringValue(editorInterface.Sys.Space.Sys.Id)
	e.Environment = types.StringValue(editorInterface.Sys.Environment.Sys.Id)
	e.ContentType = types.StringValue(editorInterface.Sys.ContentType.Sys.Id)

	e.ID = types.StringValue(fmt.Sprintf("%s:%s:%s",
		e.SpaceID.ValueString(),
		e.Environment.ValueString(),
		e.ContentType.ValueString()))
	e.Version = types.Int64Value(editorInterface.Sys.Version)

	controls := make([]Control, 0, len(editorInterface.Controls))
	for _, control := range editorInterface.Controls {
		widgetNamespace := types.StringValue("")
		if control.WidgetNamespace != nil {
			widgetNamespace = types.StringValue(string(*control.WidgetNamespace))
		}

		newControl := Control{
			FieldID:         types.StringValue(control.FieldId),
			WidgetID:        types.StringPointerValue(control.WidgetId),
			WidgetNamespace: widgetNamespace,
		}

		if control.Settings != nil {
			newControl.Settings = &Settings{}
			newControl.Settings.Import(control.Settings)
		}

		controls = append(controls, newControl)
	}

	e.Controls = controls

	if editorInterface.Sidebar != nil {
		e.Sidebar = pie.Map(*editorInterface.Sidebar, func(t sdk.EditorInterfaceSidebarItem) Sidebar {

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
	} else {
		e.Sidebar = nil
	}

	if editorInterface.Editors != nil {
		e.Editors = pie.Map(*editorInterface.Editors, func(t sdk.EditorInterfaceEditor) Editor {

			settings := jsontypes.NewNormalizedValue("{}")

			if t.Settings != nil {
				data, _ := json.Marshal(t.Settings)
				settings = jsontypes.NewNormalizedValue(string(data))
			}
			return Editor{
				WidgetId:        types.StringValue(*t.WidgetId),
				WidgetNamespace: types.StringValue(string(*t.WidgetNamespace)),
				Settings:        settings,
				Disabled:        types.BoolPointerValue(t.Disabled),
			}
		})
	} else {
		e.Editors = nil
	}

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

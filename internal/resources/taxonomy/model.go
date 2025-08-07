package taxonomy

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/labd/terraform-provider-contentful/internal/sdk"
	"github.com/labd/terraform-provider-contentful/internal/utils"
)

// TaxonomyConcept is the main resource schema data
type TaxonomyConcept struct {
	ID           types.String `tfsdk:"id"`
	Version      types.Int64  `tfsdk:"version"`
	SpaceID      types.String `tfsdk:"space_id"`
	Environment  types.String `tfsdk:"environment"`
	ConceptScheme types.String `tfsdk:"concept_scheme_id"`
	PrefLabel    types.Map    `tfsdk:"pref_label"`
	AltLabel     types.Map    `tfsdk:"alt_label"`
	Definition   types.Map    `tfsdk:"definition"`
	Note         types.Map    `tfsdk:"note"`
	Notations    types.List   `tfsdk:"notations"`
	Broader      types.List   `tfsdk:"broader"`
	Narrower     types.List   `tfsdk:"narrower"`
	Related      types.List   `tfsdk:"related"`
}

// Import populates the TaxonomyConcept struct from an SDK taxonomy concept object
func (t *TaxonomyConcept) Import(concept *sdk.TaxonomyConcept) {
	if concept.Sys.Id != nil {
		t.ID = types.StringValue(*concept.Sys.Id)
	} else {
		t.ID = types.StringNull()
	}

	if concept.Sys.Space != nil && concept.Sys.Space.Sys != nil {
		t.SpaceID = types.StringValue(concept.Sys.Space.Sys.Id)
	} else {
		t.SpaceID = types.StringNull()
	}

	if concept.Sys.Version != nil {
		t.Version = types.Int64Value(*concept.Sys.Version)
	} else {
		t.Version = types.Int64Null()
	}

	if concept.Sys.Environment != nil && concept.Sys.Environment.Sys != nil {
		t.Environment = types.StringValue(concept.Sys.Environment.Sys.Id)
	} else {
		t.Environment = types.StringNull()
	}

	// ConceptScheme
	if concept.ConceptScheme.Sys.Id != nil {
		t.ConceptScheme = types.StringValue(*concept.ConceptScheme.Sys.Id)
	} else {
		t.ConceptScheme = types.StringNull()
	}

	// PrefLabel
	if concept.PrefLabel != nil && len(concept.PrefLabel) > 0 {
		prefLabelMap := make(map[string]attr.Value)
		for k, v := range concept.PrefLabel {
			prefLabelMap[k] = types.StringValue(v)
		}
		t.PrefLabel = types.MapValueMust(types.StringType, prefLabelMap)
	} else {
		t.PrefLabel = types.MapNull(types.StringType)
	}

	// AltLabel (map of string arrays)
	if concept.AltLabel != nil && len(*concept.AltLabel) > 0 {
		altLabelMap := make(map[string]attr.Value)
		for k, v := range *concept.AltLabel {
			stringValues := make([]attr.Value, len(v))
			for i, str := range v {
				stringValues[i] = types.StringValue(str)
			}
			altLabelMap[k] = types.ListValueMust(types.StringType, stringValues)
		}
		t.AltLabel = types.MapValueMust(types.ListType{ElemType: types.StringType}, altLabelMap)
	} else {
		t.AltLabel = types.MapNull(types.ListType{ElemType: types.StringType})
	}

	// Definition
	if concept.Definition != nil && len(*concept.Definition) > 0 {
		definitionMap := make(map[string]attr.Value)
		for k, v := range *concept.Definition {
			definitionMap[k] = types.StringValue(v)
		}
		t.Definition = types.MapValueMust(types.StringType, definitionMap)
	} else {
		t.Definition = types.MapNull(types.StringType)
	}

	// Note
	if concept.Note != nil && len(*concept.Note) > 0 {
		noteMap := make(map[string]attr.Value)
		for k, v := range *concept.Note {
			noteMap[k] = types.StringValue(v)
		}
		t.Note = types.MapValueMust(types.StringType, noteMap)
	} else {
		t.Note = types.MapNull(types.StringType)
	}

	// Notations
	if concept.Notations != nil && len(*concept.Notations) > 0 {
		notationsValues := make([]attr.Value, len(*concept.Notations))
		for i, notation := range *concept.Notations {
			notationsValues[i] = types.StringValue(notation)
		}
		t.Notations = types.ListValueMust(types.StringType, notationsValues)
	} else {
		t.Notations = types.ListNull(types.StringType)
	}

	// Broader concepts
	if concept.Broader != nil && len(*concept.Broader) > 0 {
		broaderValues := make([]attr.Value, len(*concept.Broader))
		for i, link := range *concept.Broader {
			if link.Sys.Id != nil {
				broaderValues[i] = types.StringValue(*link.Sys.Id)
			} else {
				broaderValues[i] = types.StringValue("")
			}
		}
		t.Broader = types.ListValueMust(types.StringType, broaderValues)
	} else {
		t.Broader = types.ListNull(types.StringType)
	}

	// Narrower concepts
	if concept.Narrower != nil && len(*concept.Narrower) > 0 {
		narrowerValues := make([]attr.Value, len(*concept.Narrower))
		for i, link := range *concept.Narrower {
			if link.Sys.Id != nil {
				narrowerValues[i] = types.StringValue(*link.Sys.Id)
			} else {
				narrowerValues[i] = types.StringValue("")
			}
		}
		t.Narrower = types.ListValueMust(types.StringType, narrowerValues)
	} else {
		t.Narrower = types.ListNull(types.StringType)
	}

	// Related concepts
	if concept.Related != nil && len(*concept.Related) > 0 {
		relatedValues := make([]attr.Value, len(*concept.Related))
		for i, link := range *concept.Related {
			if link.Sys.Id != nil {
				relatedValues[i] = types.StringValue(*link.Sys.Id)
			} else {
				relatedValues[i] = types.StringValue("")
			}
		}
		t.Related = types.ListValueMust(types.StringType, relatedValues)
	} else {
		t.Related = types.ListNull(types.StringType)
	}
}

// DraftForCreate creates a TaxonomyConceptCreate object for creating a new taxonomy concept
func (t *TaxonomyConcept) DraftForCreate() sdk.TaxonomyConceptCreate {
	conceptCreate := sdk.TaxonomyConceptCreate{}

	// ConceptScheme
	if !t.ConceptScheme.IsNull() && !t.ConceptScheme.IsUnknown() {
		conceptCreate.ConceptScheme = sdk.TaxonomyConceptScheme{
			Sys: struct {
				Id       *string                                `json:"id,omitempty"`
				LinkType *sdk.TaxonomyConceptSchemeSysLinkType `json:"linkType,omitempty"`
				Type     *sdk.TaxonomyConceptSchemeSysType     `json:"type,omitempty"`
			}{
				Id:       utils.Pointer(t.ConceptScheme.ValueString()),
				LinkType: utils.Pointer(sdk.TaxonomyConceptSchemeSysLinkTypeTaxonomyConceptScheme),
				Type:     utils.Pointer(sdk.TaxonomyConceptSchemeSysTypeLink),
			},
		}
	}

	// PrefLabel
	if !t.PrefLabel.IsNull() && !t.PrefLabel.IsUnknown() {
		prefLabelMap := make(map[string]string)
		for k, v := range t.PrefLabel.Elements() {
			prefLabelMap[k] = v.(types.String).ValueString()
		}
		conceptCreate.PrefLabel = prefLabelMap
	}

	// AltLabel
	if !t.AltLabel.IsNull() && !t.AltLabel.IsUnknown() {
		altLabelMap := make(map[string][]string)
		for k, v := range t.AltLabel.Elements() {
			list := v.(types.List)
			stringArray := make([]string, len(list.Elements()))
			for i, elem := range list.Elements() {
				stringArray[i] = elem.(types.String).ValueString()
			}
			altLabelMap[k] = stringArray
		}
		conceptCreate.AltLabel = &altLabelMap
	}

	// Definition
	if !t.Definition.IsNull() && !t.Definition.IsUnknown() {
		definitionMap := make(map[string]string)
		for k, v := range t.Definition.Elements() {
			definitionMap[k] = v.(types.String).ValueString()
		}
		conceptCreate.Definition = &definitionMap
	}

	// Note
	if !t.Note.IsNull() && !t.Note.IsUnknown() {
		noteMap := make(map[string]string)
		for k, v := range t.Note.Elements() {
			noteMap[k] = v.(types.String).ValueString()
		}
		conceptCreate.Note = &noteMap
	}

	// Notations
	if !t.Notations.IsNull() && !t.Notations.IsUnknown() {
		notations := make([]string, len(t.Notations.Elements()))
		for i, elem := range t.Notations.Elements() {
			notations[i] = elem.(types.String).ValueString()
		}
		conceptCreate.Notations = &notations
	}

	// Broader
	if !t.Broader.IsNull() && !t.Broader.IsUnknown() {
		broaderList := make([]sdk.TaxonomyConceptLink, len(t.Broader.Elements()))
		for i, elem := range t.Broader.Elements() {
			id := elem.(types.String).ValueString()
			broaderList[i] = sdk.TaxonomyConceptLink{
				Sys: struct {
					Id       *string                               `json:"id,omitempty"`
					LinkType *sdk.TaxonomyConceptLinkSysLinkType  `json:"linkType,omitempty"`
					Type     *sdk.TaxonomyConceptLinkSysType      `json:"type,omitempty"`
				}{
					Id:       utils.Pointer(id),
					LinkType: utils.Pointer(sdk.TaxonomyConceptLinkSysLinkTypeTaxonomyConcept),
					Type:     utils.Pointer(sdk.TaxonomyConceptLinkSysTypeLink),
				},
			}
		}
		conceptCreate.Broader = &broaderList
	}

	// Narrower
	if !t.Narrower.IsNull() && !t.Narrower.IsUnknown() {
		narrowerList := make([]sdk.TaxonomyConceptLink, len(t.Narrower.Elements()))
		for i, elem := range t.Narrower.Elements() {
			id := elem.(types.String).ValueString()
			narrowerList[i] = sdk.TaxonomyConceptLink{
				Sys: struct {
					Id       *string                               `json:"id,omitempty"`
					LinkType *sdk.TaxonomyConceptLinkSysLinkType  `json:"linkType,omitempty"`
					Type     *sdk.TaxonomyConceptLinkSysType      `json:"type,omitempty"`
				}{
					Id:       utils.Pointer(id),
					LinkType: utils.Pointer(sdk.TaxonomyConceptLinkSysLinkTypeTaxonomyConcept),
					Type:     utils.Pointer(sdk.TaxonomyConceptLinkSysTypeLink),
				},
			}
		}
		conceptCreate.Narrower = &narrowerList
	}

	// Related
	if !t.Related.IsNull() && !t.Related.IsUnknown() {
		relatedList := make([]sdk.TaxonomyConceptLink, len(t.Related.Elements()))
		for i, elem := range t.Related.Elements() {
			id := elem.(types.String).ValueString()
			relatedList[i] = sdk.TaxonomyConceptLink{
				Sys: struct {
					Id       *string                               `json:"id,omitempty"`
					LinkType *sdk.TaxonomyConceptLinkSysLinkType  `json:"linkType,omitempty"`
					Type     *sdk.TaxonomyConceptLinkSysType      `json:"type,omitempty"`
				}{
					Id:       utils.Pointer(id),
					LinkType: utils.Pointer(sdk.TaxonomyConceptLinkSysLinkTypeTaxonomyConcept),
					Type:     utils.Pointer(sdk.TaxonomyConceptLinkSysTypeLink),
				},
			}
		}
		conceptCreate.Related = &relatedList
	}

	return conceptCreate
}

// DraftForUpdate creates a TaxonomyConceptUpdate object for updating an existing taxonomy concept
func (t *TaxonomyConcept) DraftForUpdate() sdk.TaxonomyConceptUpdate {
	conceptUpdate := sdk.TaxonomyConceptUpdate{}

	// ConceptScheme
	if !t.ConceptScheme.IsNull() && !t.ConceptScheme.IsUnknown() {
		conceptUpdate.ConceptScheme = sdk.TaxonomyConceptScheme{
			Sys: struct {
				Id       *string                                `json:"id,omitempty"`
				LinkType *sdk.TaxonomyConceptSchemeSysLinkType `json:"linkType,omitempty"`
				Type     *sdk.TaxonomyConceptSchemeSysType     `json:"type,omitempty"`
			}{
				Id:       utils.Pointer(t.ConceptScheme.ValueString()),
				LinkType: utils.Pointer(sdk.TaxonomyConceptSchemeSysLinkTypeTaxonomyConceptScheme),
				Type:     utils.Pointer(sdk.TaxonomyConceptSchemeSysTypeLink),
			},
		}
	}

	// PrefLabel
	if !t.PrefLabel.IsNull() && !t.PrefLabel.IsUnknown() {
		prefLabelMap := make(map[string]string)
		for k, v := range t.PrefLabel.Elements() {
			prefLabelMap[k] = v.(types.String).ValueString()
		}
		conceptUpdate.PrefLabel = prefLabelMap
	}

	// AltLabel
	if !t.AltLabel.IsNull() && !t.AltLabel.IsUnknown() {
		altLabelMap := make(map[string][]string)
		for k, v := range t.AltLabel.Elements() {
			list := v.(types.List)
			stringArray := make([]string, len(list.Elements()))
			for i, elem := range list.Elements() {
				stringArray[i] = elem.(types.String).ValueString()
			}
			altLabelMap[k] = stringArray
		}
		conceptUpdate.AltLabel = &altLabelMap
	}

	// Definition
	if !t.Definition.IsNull() && !t.Definition.IsUnknown() {
		definitionMap := make(map[string]string)
		for k, v := range t.Definition.Elements() {
			definitionMap[k] = v.(types.String).ValueString()
		}
		conceptUpdate.Definition = &definitionMap
	}

	// Note
	if !t.Note.IsNull() && !t.Note.IsUnknown() {
		noteMap := make(map[string]string)
		for k, v := range t.Note.Elements() {
			noteMap[k] = v.(types.String).ValueString()
		}
		conceptUpdate.Note = &noteMap
	}

	// Notations
	if !t.Notations.IsNull() && !t.Notations.IsUnknown() {
		notations := make([]string, len(t.Notations.Elements()))
		for i, elem := range t.Notations.Elements() {
			notations[i] = elem.(types.String).ValueString()
		}
		conceptUpdate.Notations = &notations
	}

	// Broader
	if !t.Broader.IsNull() && !t.Broader.IsUnknown() {
		broaderList := make([]sdk.TaxonomyConceptLink, len(t.Broader.Elements()))
		for i, elem := range t.Broader.Elements() {
			id := elem.(types.String).ValueString()
			broaderList[i] = sdk.TaxonomyConceptLink{
				Sys: struct {
					Id       *string                               `json:"id,omitempty"`
					LinkType *sdk.TaxonomyConceptLinkSysLinkType  `json:"linkType,omitempty"`
					Type     *sdk.TaxonomyConceptLinkSysType      `json:"type,omitempty"`
				}{
					Id:       utils.Pointer(id),
					LinkType: utils.Pointer(sdk.TaxonomyConceptLinkSysLinkTypeTaxonomyConcept),
					Type:     utils.Pointer(sdk.TaxonomyConceptLinkSysTypeLink),
				},
			}
		}
		conceptUpdate.Broader = &broaderList
	}

	// Narrower
	if !t.Narrower.IsNull() && !t.Narrower.IsUnknown() {
		narrowerList := make([]sdk.TaxonomyConceptLink, len(t.Narrower.Elements()))
		for i, elem := range t.Narrower.Elements() {
			id := elem.(types.String).ValueString()
			narrowerList[i] = sdk.TaxonomyConceptLink{
				Sys: struct {
					Id       *string                               `json:"id,omitempty"`
					LinkType *sdk.TaxonomyConceptLinkSysLinkType  `json:"linkType,omitempty"`
					Type     *sdk.TaxonomyConceptLinkSysType      `json:"type,omitempty"`
				}{
					Id:       utils.Pointer(id),
					LinkType: utils.Pointer(sdk.TaxonomyConceptLinkSysLinkTypeTaxonomyConcept),
					Type:     utils.Pointer(sdk.TaxonomyConceptLinkSysTypeLink),
				},
			}
		}
		conceptUpdate.Narrower = &narrowerList
	}

	// Related
	if !t.Related.IsNull() && !t.Related.IsUnknown() {
		relatedList := make([]sdk.TaxonomyConceptLink, len(t.Related.Elements()))
		for i, elem := range t.Related.Elements() {
			id := elem.(types.String).ValueString()
			relatedList[i] = sdk.TaxonomyConceptLink{
				Sys: struct {
					Id       *string                               `json:"id,omitempty"`
					LinkType *sdk.TaxonomyConceptLinkSysLinkType  `json:"linkType,omitempty"`
					Type     *sdk.TaxonomyConceptLinkSysType      `json:"type,omitempty"`
				}{
					Id:       utils.Pointer(id),
					LinkType: utils.Pointer(sdk.TaxonomyConceptLinkSysLinkTypeTaxonomyConcept),
					Type:     utils.Pointer(sdk.TaxonomyConceptLinkSysTypeLink),
				},
			}
		}
		conceptUpdate.Related = &relatedList
	}

	return conceptUpdate
}
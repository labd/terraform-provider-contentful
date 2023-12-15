package app_installation

import (
	"encoding/json"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/labd/contentful-go"
)

// AppInstallation is the main resource schema data
type AppInstallation struct {
	ID              types.String         `tfsdk:"id"`
	AppDefinitionID types.String         `tfsdk:"app_definition_id"`
	SpaceId         types.String         `tfsdk:"space_id"`
	Environment     types.String         `tfsdk:"environment"`
	Parameters      jsontypes.Normalized `tfsdk:"parameters"`
	AcceptedTerms   []types.String       `tfsdk:"accepted_terms"`
}

func (a *AppInstallation) Draft() *contentful.AppInstallation {

	app := &contentful.AppInstallation{}

	parameters := make(map[string]any)

	a.Parameters.Unmarshal(&parameters)

	app.Parameters = parameters

	return app
}

func (a *AppInstallation) Equal(n *contentful.AppInstallation) bool {

	data, _ := json.Marshal(n.Parameters)

	if a.Parameters.ValueString() != string(data) {
		return false
	}

	return true
}

func (a *AppInstallation) Import(n *contentful.AppInstallation) {
	a.AppDefinitionID = types.StringValue(n.Sys.AppDefinition.Sys.ID)
	a.ID = types.StringValue(n.Sys.AppDefinition.Sys.ID)
	data, _ := json.Marshal(n.Parameters)
	a.Parameters = jsontypes.NewNormalizedValue(string(data))
}

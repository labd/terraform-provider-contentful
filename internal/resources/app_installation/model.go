package app_installation

import (
	"encoding/json"
	"github.com/elliotchance/pie/v2"
	"github.com/flaconi/contentful-go/pkgs/model"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
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

func (a *AppInstallation) Draft() *model.AppInstallation {

	app := &model.AppInstallation{}

	parameters := make(map[string]any)

	a.Parameters.Unmarshal(&parameters)

	if !a.AppDefinitionID.IsUnknown() || !a.AppDefinitionID.IsNull() {

		appDefinitionSys := struct {
			Sys model.BaseSys
		}{
			Sys: model.BaseSys{
				ID: a.AppDefinitionID.ValueString(),
			},
		}

		app.Sys = &model.AppInstallationSys{AppDefinition: (*struct {
			Sys model.BaseSys `json:"sys,omitempty"`
		})(&appDefinitionSys)}
	}

	app.Terms = pie.Map(a.AcceptedTerms, func(t types.String) string {
		return t.ValueString()
	})
	app.Parameters = parameters

	return app
}

func (a *AppInstallation) Equal(n *model.AppInstallation) bool {

	data, _ := json.Marshal(n.Parameters)

	if a.Parameters.ValueString() != string(data) {
		return false
	}

	return true
}

func (a *AppInstallation) Import(n *model.AppInstallation) {
	a.AppDefinitionID = types.StringValue(n.Sys.AppDefinition.Sys.ID)
	a.ID = types.StringValue(n.Sys.AppDefinition.Sys.ID)
	data, _ := json.Marshal(n.Parameters)
	a.Parameters = jsontypes.NewNormalizedValue(string(data))
}

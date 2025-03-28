package app_installation

import (
	"encoding/json"
	"strings"

	"github.com/elliotchance/pie/v2"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/labd/terraform-provider-contentful/internal/sdk"
	"github.com/labd/terraform-provider-contentful/internal/utils"
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

func (a *AppInstallation) Draft() *sdk.AppInstallationUpsert {

	parameters := make(map[string]any)
	a.Parameters.Unmarshal(&parameters)

	app := &sdk.AppInstallationUpsert{
		Parameters: parameters,
	}

	return app
}

func (a *AppInstallation) AcceptedTermsHeader() *string {
	terms := pie.Map(a.AcceptedTerms, func(t types.String) string {
		return t.ValueString()
	})
	if len(terms) == 0 {
		return nil
	}
	return utils.Pointer(strings.Join(terms, ", "))
}

func (a *AppInstallation) Equal(n *sdk.AppInstallation) bool {

	data, _ := json.Marshal(n.Parameters)

	if a.Parameters.ValueString() != string(data) {
		return false
	}

	return true
}

func (a *AppInstallation) Import(n *sdk.AppInstallation) {
	a.AppDefinitionID = types.StringValue(n.Sys.AppDefinition.Sys.Id)
	a.ID = types.StringValue(n.Sys.AppDefinition.Sys.Id)
	data, _ := json.Marshal(n.Parameters)
	a.Parameters = jsontypes.NewNormalizedValue(string(data))
}

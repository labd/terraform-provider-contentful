package oauth_application

import (
	"github.com/elliotchance/pie/v2"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/labd/terraform-provider-contentful/internal/sdk"
)

// OAuthApplication is the Terraform state shape for a Contentful OAuth application.
type OAuthApplication struct {
	ID           types.String   `tfsdk:"id"`
	Name         types.String   `tfsdk:"name"`
	Description  types.String   `tfsdk:"description"`
	Scopes       []types.String `tfsdk:"scopes"`
	RedirectURI  types.String   `tfsdk:"redirect_uri"`
	Confidential types.Bool     `tfsdk:"confidential"`
	ClientID     types.String   `tfsdk:"client_id"`
	ClientSecret types.String   `tfsdk:"client_secret"`
}

func (a *OAuthApplication) Draft() *sdk.OAuthApplicationDraft {
	return &sdk.OAuthApplicationDraft{
		Name:         a.Name.ValueString(),
		Description:  a.Description.ValueString(),
		RedirectUri:  a.RedirectURI.ValueString(),
		Confidential: a.Confidential.ValueBool(),
		Scopes: pie.Map(a.Scopes, func(s types.String) sdk.OAuthApplicationScope {
			return sdk.OAuthApplicationScope(s.ValueString())
		}),
	}
}

// Import populates the model from an API response. `ClientSecret` is intentionally
// not touched here — the API never returns it outside of the create response, so
// preserving the existing state value is the caller's responsibility.
func (a *OAuthApplication) Import(n *sdk.OAuthApplication) {
	a.ID = types.StringValue(n.Sys.Id)
	a.Name = types.StringValue(n.Name)
	a.Description = types.StringValue(n.Description)
	a.RedirectURI = types.StringValue(n.RedirectUri)
	a.Confidential = types.BoolValue(n.Confidential)
	a.ClientID = types.StringValue(n.ClientId)
	a.Scopes = pie.Map(n.Scopes, func(s sdk.OAuthApplicationScope) types.String {
		return types.StringValue(string(s))
	})
}

package api_key

import (
	"github.com/elliotchance/pie/v2"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/labd/terraform-provider-contentful/internal/sdk"
)

// ApiKey is the main resource schema data
type ApiKey struct {
	ID           types.String   `tfsdk:"id"`
	PreviewID    types.String   `tfsdk:"preview_id"`
	SpaceId      types.String   `tfsdk:"space_id"`
	Name         types.String   `tfsdk:"name"`
	Description  types.String   `tfsdk:"description"`
	Version      types.Int64    `tfsdk:"version"`
	AccessToken  types.String   `tfsdk:"access_token"`
	PreviewToken types.String   `tfsdk:"preview_token"`
	Environments []types.String `tfsdk:"environments"`
}

func (a *ApiKey) Import(n *sdk.ApiKey) {
	a.ID = types.StringValue(*n.Sys.Id)
	a.PreviewID = types.StringValue(*n.PreviewApiKey.Sys.Id)
	a.SpaceId = types.StringValue(n.Sys.Space.Sys.Id)
	a.Version = types.Int64Value(int64(*n.Sys.Version))
	a.Name = types.StringValue(n.Name)
	a.Description = types.StringNull()
	if n.Description != "" {
		a.Description = types.StringValue(n.Description)
	}

	a.Environments = pie.Map(n.Environments, func(t sdk.EnvironmentSystemProperties) types.String {
		return types.StringValue(t.Sys.Id)
	})

	a.AccessToken = types.StringValue(n.AccessToken)
}

func (a *ApiKey) Draft() *sdk.ApiKeyDraft {
	draft := &sdk.ApiKeyDraft{}

	if !a.Description.IsNull() && !a.Description.IsUnknown() {
		draft.Description = a.Description.ValueString()
	}

	envs := pie.Map(a.Environments, func(t types.String) sdk.EnvironmentSystemProperties {
		return sdk.EnvironmentSystemProperties{
			Sys: &sdk.EnvironmentSystemPropertiesSys{
				Id:       t.ValueString(),
				Type:     "Link",
				LinkType: "Environment",
			},
		}
	})
	draft.Environments = &envs

	draft.Name = a.Name.ValueString()

	return draft
}

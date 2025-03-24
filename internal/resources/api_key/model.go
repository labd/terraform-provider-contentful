package api_key

import (
	"github.com/elliotchance/pie/v2"
	"github.com/flaconi/contentful-go/pkgs/model"
	"github.com/hashicorp/terraform-plugin-framework/types"
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

func (a *ApiKey) Import(n *model.APIKey) {
	a.ID = types.StringValue(n.Sys.ID)
	a.PreviewID = types.StringValue(n.PreviewAPIKey.Sys.ID)
	a.SpaceId = types.StringValue(n.Sys.Space.Sys.ID)
	a.Version = types.Int64Value(int64(n.Sys.Version))
	a.Name = types.StringValue(n.Name)
	a.Description = types.StringNull()
	if n.Description != "" {
		a.Description = types.StringValue(n.Description)
	}

	a.Environments = pie.Map(n.Environments, func(t model.Environments) types.String {
		return types.StringValue(t.Sys.ID)
	})

	a.AccessToken = types.StringValue(n.AccessToken)
}

func (a *ApiKey) Draft() *model.APIKey {
	draft := &model.APIKey{}

	if !a.ID.IsUnknown() || !a.ID.IsNull() {
		draft.Sys = &model.SpaceSys{CreatedSys: model.CreatedSys{BaseSys: model.BaseSys{ID: a.ID.ValueString(), Version: int(a.Version.ValueInt64())}}}

	}

	if !a.Description.IsNull() && !a.Description.IsUnknown() {
		draft.Description = a.Description.ValueString()
	}

	draft.Environments = pie.Map(a.Environments, func(t types.String) model.Environments {
		return model.Environments{
			Sys: model.BaseSys{
				ID:       t.ValueString(),
				Type:     "Link",
				LinkType: "Environment",
			},
		}
	})

	draft.Name = a.Name.ValueString()

	return draft
}

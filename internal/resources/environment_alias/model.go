package environment_alias

import (
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/labd/terraform-provider-contentful/internal/sdk"
)

// EnvironmentAlias is the main resource schema data
type EnvironmentAlias struct {
	ID            types.String `tfsdk:"id"`
	Version       types.Int64  `tfsdk:"version"`
	SpaceID       types.String `tfsdk:"space_id"`
	EnvironmentID types.String `tfsdk:"environment_id"`
}

// Import populates the EnvironmentAlias struct from an SDK environment alias object
func (e *EnvironmentAlias) Import(environmentAlias *sdk.EnvironmentAlias) {
	e.ID = types.StringValue(environmentAlias.Sys.Id)
	e.Version = types.Int64Value(int64(environmentAlias.Sys.Version))
	e.SpaceID = types.StringValue(environmentAlias.Sys.Space.Sys.Id)
	e.EnvironmentID = types.StringValue(environmentAlias.Environment.Sys.Id)
}

// DraftForUpdate creates an EnvironmentAliasUpdate object for creating or updating an environment alias
func (e *EnvironmentAlias) DraftForUpdate() sdk.EnvironmentAliasUpdate {
	return sdk.EnvironmentAliasUpdate{
		Environment: sdk.EnvironmentSystemProperties{
			Sys: &sdk.EnvironmentSystemPropertiesSys{
				Id:       e.EnvironmentID.ValueString(),
				LinkType: sdk.EnvironmentSystemPropertiesSysLinkTypeEnvironment,
				Type:     sdk.EnvironmentSystemPropertiesSysTypeLink,
			},
		},
	}
}

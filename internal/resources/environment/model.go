package environment

import (
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/labd/terraform-provider-contentful/internal/sdk"
)

// Environment is the main resource schema data
type Environment struct {
	ID      types.String `tfsdk:"id"`
	Version types.Int64  `tfsdk:"version"`
	SpaceId types.String `tfsdk:"space_id"`
	Name    types.String `tfsdk:"name"`
}

// Import populates the Environment struct from an SDK environment object
func (e *Environment) Import(environment *sdk.Environment) {
	e.ID = types.StringValue(environment.Sys.Id)
	e.Version = types.Int64Value(int64(environment.Sys.Version))
	e.SpaceId = types.StringValue(environment.Sys.Space.Sys.Id)
	e.Name = types.StringValue(environment.Name)
}

// DraftForCreate creates an EnvironmentCreate object for creating a new environment
func (e *Environment) DraftForCreate() sdk.EnvironmentCreate {
	return sdk.EnvironmentCreate{
		Name: e.Name.ValueString(),
	}
}

// DraftForUpdate creates an EnvironmentUpdate object for updating an existing environment
func (e *Environment) DraftForUpdate() sdk.EnvironmentUpdate {
	return sdk.EnvironmentUpdate{
		Name: e.Name.ValueString(),
	}
}

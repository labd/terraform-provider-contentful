package team

import (
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/labd/terraform-provider-contentful/internal/sdk"
)

// Team is the main resource schema data
type Team struct {
	ID          types.String `tfsdk:"id"`
	Version     types.Int64  `tfsdk:"version"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
}

// Import populates the Team struct from an SDK team object
func (t *Team) Import(team *sdk.Team) {
	t.ID = types.StringValue(team.Sys.Id)
	t.Version = types.Int64Value(int64(team.Sys.Version))
	t.Name = types.StringValue(team.Name)

	if team.Description != nil {
		t.Description = types.StringValue(*team.Description)
	} else {
		t.Description = types.StringNull()
	}
}

// DraftForCreate creates a TeamCreate object for creating a new team
func (t *Team) DraftForCreate() sdk.TeamCreate {
	draft := sdk.TeamCreate{
		Name: t.Name.ValueString(),
	}

	if !t.Description.IsNull() && !t.Description.IsUnknown() {
		description := t.Description.ValueString()
		draft.Description = &description
	}

	return draft
}

// DraftForUpdate creates a TeamUpdate object for updating an existing team
func (t *Team) DraftForUpdate() sdk.TeamUpdate {
	draft := sdk.TeamUpdate{
		Name: t.Name.ValueString(),
	}

	if !t.Description.IsNull() && !t.Description.IsUnknown() {
		description := t.Description.ValueString()
		draft.Description = &description
	}

	return draft
}

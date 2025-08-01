package preview_environment

import (
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/labd/terraform-provider-contentful/internal/sdk"
)

// PreviewEnvironment is the main resource schema data
type PreviewEnvironment struct {
	ID            types.String    `tfsdk:"id"`
	SpaceID       types.String    `tfsdk:"space_id"`
	Name          types.String    `tfsdk:"name"`
	Description   types.String    `tfsdk:"description"`
	Version       types.Int64     `tfsdk:"version"`
	Configuration []Configuration `tfsdk:"configuration"`
}

// Configuration represents a content type preview configuration
type Configuration struct {
	ContentType types.String `tfsdk:"content_type"`
	Enabled     types.Bool   `tfsdk:"enabled"`
	URL         types.String `tfsdk:"url"`
}

// Import populates the PreviewEnvironment struct from an SDK response
func (p *PreviewEnvironment) Import(env *sdk.PreviewEnvironment) {
	p.ID = types.StringValue(env.Sys.Id)
	p.SpaceID = types.StringValue(env.Sys.Space.Sys.Id)
	p.Name = types.StringValue(env.Name)
	p.Version = types.Int64PointerValue(env.Sys.Version)
	if env.Description != nil && *env.Description != "" {
		p.Description = types.StringPointerValue(env.Description)
	}

	// Map configurations
	var configurations []Configuration
	for _, config := range env.Configurations {
		configurations = append(configurations, Configuration{
			ContentType: types.StringValue(config.ContentType),
			Enabled:     types.BoolValue(config.Enabled),
			URL:         types.StringValue(config.Url),
		})
	}
	p.Configuration = configurations
}

// Draft creates a PreviewEnvironmentDraft for API operations
func (p *PreviewEnvironment) Draft() *sdk.PreviewEnvironmentInput {
	draft := &sdk.PreviewEnvironmentInput{
		Name: p.Name.ValueString(),
	}

	if !p.Description.IsNull() && !p.Description.IsUnknown() {
		description := p.Description.ValueStringPointer()
		draft.Description = description
	}

	// Map configurations
	if len(p.Configuration) > 0 {
		config := make([]sdk.PreviewConfiguration, 0, len(p.Configuration))
		for _, cfg := range p.Configuration {
			config = append(config, sdk.PreviewConfiguration{
				ContentType: cfg.ContentType.ValueString(),
				Enabled:     cfg.Enabled.ValueBool(),
				Url:         cfg.URL.ValueString(),
			})
		}
		draft.Configurations = config
	}

	return draft
}

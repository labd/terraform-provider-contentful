package tag

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"

	"github.com/labd/terraform-provider-contentful/internal/sdk"
)

func TestTag_Import(t *testing.T) {
	tag := Tag{}
	tag.Import(&sdk.Tag{
		Name: "Campaign",
		Sys: sdk.SystemPropertiesTag{
			Id:          "campaign",
			Version:     2,
			Visibility:  sdk.TagVisibilityPublic,
			Space:       sdk.SystemPropertiesReference{Sys: sdk.SystemPropertiesLink{Id: "space"}},
			Environment: sdk.SystemPropertiesReference{Sys: sdk.SystemPropertiesLink{Id: "master"}},
		},
	})

	assert.Equal(t, types.StringValue("campaign"), tag.ID)
	assert.Equal(t, types.Int64Value(2), tag.Version)
	assert.Equal(t, types.StringValue("space"), tag.SpaceID)
	assert.Equal(t, types.StringValue("master"), tag.Environment)
	assert.Equal(t, types.StringValue("Campaign"), tag.Name)
	assert.Equal(t, types.StringValue("public"), tag.Visibility)
}

func TestTag_Draft(t *testing.T) {
	tag := Tag{Name: types.StringValue("Campaign")}
	assert.Equal(t, sdk.TagDraft{Name: "Campaign"}, tag.Draft())
}

func TestTag_VisibilityForCreate(t *testing.T) {
	tests := map[string]struct {
		visibility types.String
		expected   sdk.TagVisibility
	}{
		"null defaults to private": {
			visibility: types.StringNull(),
			expected:   sdk.TagVisibilityPrivate,
		},
		"unknown defaults to private": {
			visibility: types.StringUnknown(),
			expected:   sdk.TagVisibilityPrivate,
		},
		"configured visibility is preserved": {
			visibility: types.StringValue("public"),
			expected:   sdk.TagVisibilityPublic,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			tag := Tag{Visibility: test.visibility}
			assert.Equal(t, test.expected, tag.VisibilityForCreate())
		})
	}
}

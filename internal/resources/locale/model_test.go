package locale

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"

	"github.com/labd/terraform-provider-contentful/internal/sdk"
)

func TestLocale_ImportUsesEnvironmentFromResponse(t *testing.T) {
	locale := Locale{}
	err := locale.Import(&sdk.Locale{
		Code:                 "fr-FR",
		ContentDeliveryApi:   true,
		ContentManagementApi: true,
		Name:                 "French",
		Sys: sdk.SystemPropertiesResource{
			Id:          "locale-id",
			Space:       sdk.SystemPropertiesReference{Sys: sdk.SystemPropertiesLink{Id: "space-id"}},
			Environment: &sdk.SystemPropertiesReference{Sys: sdk.SystemPropertiesLink{Id: "staging"}},
			Version:     1,
		},
	})

	assert.NoError(t, err)
	assert.Equal(t, types.StringValue("locale-id"), locale.ID)
	assert.Equal(t, types.StringValue("space-id"), locale.SpaceID)
	assert.Equal(t, types.StringValue("staging"), locale.Environment)
	assert.Equal(t, types.Int64Value(1), locale.Version)
	assert.Equal(t, types.StringValue("French"), locale.Name)
	assert.Equal(t, types.StringValue("fr-FR"), locale.Code)
}

func TestLocale_ImportWithoutEnvironmentFails(t *testing.T) {
	locale := Locale{}
	err := locale.Import(&sdk.Locale{
		Code: "fr-FR",
		Name: "French",
		Sys: sdk.SystemPropertiesResource{
			Id:    "locale-id",
			Space: sdk.SystemPropertiesReference{Sys: sdk.SystemPropertiesLink{Id: "space-id"}},
		},
	})

	assert.ErrorContains(t, err, "locale-id")
	assert.True(t, locale.Environment.IsNull())
}

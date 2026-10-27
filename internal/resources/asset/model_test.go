package asset

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/labd/terraform-provider-contentful/internal/sdk"
)

func TestAsset_ImportWithoutEnvironmentFails(t *testing.T) {
	asset := Asset{}
	err := asset.Import(&sdk.Asset{
		Sys: sdk.SystemPropertiesContent{
			Id:    "asset-id",
			Space: sdk.SystemPropertiesReference{Sys: sdk.SystemPropertiesLink{Id: "space-id"}},
		},
	})

	assert.ErrorContains(t, err, "asset-id")
	assert.True(t, asset.Environment.IsNull())
}

package entry

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/labd/terraform-provider-contentful/internal/sdk"
)

func TestEntry_ImportWithoutEnvironmentFails(t *testing.T) {
	entry := Entry{}
	err := entry.Import(&sdk.Entry{
		Sys: sdk.SystemPropertiesEntry{
			Id:          "entry-id",
			Space:       sdk.SystemPropertiesReference{Sys: sdk.SystemPropertiesLink{Id: "space-id"}},
			ContentType: sdk.SystemPropertiesReference{Sys: sdk.SystemPropertiesLink{Id: "content-type-id"}},
		},
	})

	assert.ErrorContains(t, err, "entry-id")
	assert.True(t, entry.Environment.IsNull())
}

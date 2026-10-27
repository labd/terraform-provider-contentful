package editor_interface

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/labd/terraform-provider-contentful/internal/sdk"
)

func TestEditorInterface_ImportWithoutEnvironmentFails(t *testing.T) {
	editorInterface := EditorInterface{}
	err := editorInterface.Import(&sdk.EditorInterface{
		Sys: sdk.SystemPropertiesEntry{
			Id:          "editor-interface-id",
			Space:       sdk.SystemPropertiesReference{Sys: sdk.SystemPropertiesLink{Id: "space-id"}},
			ContentType: sdk.SystemPropertiesReference{Sys: sdk.SystemPropertiesLink{Id: "content-type-id"}},
		},
	})

	assert.ErrorContains(t, err, "content-type-id")
	assert.True(t, editorInterface.Environment.IsNull())
}

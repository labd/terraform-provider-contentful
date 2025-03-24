package contentful

import (
	"fmt"
	"testing"

	"github.com/flaconi/contentful-go/pkgs/common"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/stretchr/testify/assert"
)

func TestParseError_Nil(t *testing.T) {
	d := parseError(nil)
	assert.Nil(t, d)
}

func TestParseError_RegularErr(t *testing.T) {
	d := parseError(fmt.Errorf("regular error"))
	assert.True(t, d.HasError())
	assert.Equal(t, d[0].Summary, "regular error")
}

func TestParseError_WithoutWarning(t *testing.T) {
	d := parseError(&common.ErrorResponse{
		Message: "error message",
	})
	assert.True(t, d.HasError())
	assert.Equal(t, len(d), 1)
	assert.Equal(t, d[0].Summary, "error message")
	assert.Equal(t, d[0].Severity, diag.Error)
}

func TestParseError_WithWarning_WithoutPath(t *testing.T) {
	d := parseError(common.ErrorResponse{
		Message: "error message",
		Details: &common.ErrorDetails{
			Errors: []*common.ErrorDetail{
				{
					Details: "error detail",
				},
			},
		},
	})
	assert.True(t, d.HasError())
	assert.Equal(t, len(d), 2)
	assert.Equal(t, d[0].Summary, "error detail ()")
	assert.Equal(t, d[0].Severity, diag.Warning)
	assert.Equal(t, d[1].Summary, "error message")
	assert.Equal(t, d[1].Severity, diag.Error)
}

func TestParseError_WithWarning_WithPath(t *testing.T) {
	d := parseError(common.ErrorResponse{
		Message: "error message",
		Details: &common.ErrorDetails{
			Errors: []*common.ErrorDetail{
				{
					Path:    []interface{}{"path", "to", "error"},
					Details: "error detail",
				},
			},
		},
	})
	assert.True(t, d.HasError())
	assert.Equal(t, len(d), 2)
	assert.Equal(t, d[0].Summary, "error detail (path.to.error)")
	assert.Equal(t, d[0].Severity, diag.Warning)
	assert.Equal(t, d[1].Summary, "error message")
	assert.Equal(t, d[1].Severity, diag.Error)
}

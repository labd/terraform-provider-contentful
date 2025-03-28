package contentful

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

func parseError(err error) diag.Diagnostics {
	return diag.FromErr(err)
}

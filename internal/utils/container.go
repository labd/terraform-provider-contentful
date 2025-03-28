package utils

import (
	"github.com/labd/terraform-provider-contentful/internal/sdk"
)

type ProviderData struct {
	Client         *sdk.ClientWithResponses
	ClientUpload   *sdk.ClientWithResponses
	OrganizationId string
}

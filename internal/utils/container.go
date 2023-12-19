package utils

import (
	"github.com/flaconi/contentful-go"
	"github.com/flaconi/contentful-go/service/common"
)

type ProviderData struct {
	Client         *contentful.Client
	CMAClient      common.SpaceIdClientBuilder
	OrganizationId string
}

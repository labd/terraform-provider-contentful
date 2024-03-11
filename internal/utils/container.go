package utils

import (
	"github.com/flaconi/contentful-go"
	"github.com/flaconi/contentful-go/service/cma"
)

type ProviderData struct {
	Client         *contentful.Client
	CMAClient      cma.SpaceIdClientBuilder
	OrganizationId string
}

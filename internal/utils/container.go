package utils

import (
	"github.com/labd/contentful-go"
	"github.com/labd/contentful-go/service/cma"
)

type ProviderData struct {
	Client         *contentful.Client
	CMAClient      cma.SpaceIdClientBuilder
	OrganizationId string
}

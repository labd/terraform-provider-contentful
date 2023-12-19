package utils

import "github.com/flaconi/contentful-go"

type ProviderData struct {
	Client         *contentful.Client
	OrganizationId string
}

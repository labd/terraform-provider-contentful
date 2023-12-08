package utils

import "github.com/labd/contentful-go"

type ProviderData struct {
	Client         *contentful.Client
	OrganizationId string
}

package utils

import (
	"context"
	"fmt"
	"net/http"

	"github.com/deepmap/oapi-codegen/pkg/securityprovider"
	"github.com/hashicorp/go-retryablehttp"

	"github.com/labd/terraform-provider-contentful/internal/sdk"
)

func CreateClient(url string, token string) (*sdk.ClientWithResponses, error) {

	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = 3

	httpClient := retryClient.StandardClient()
	httpClient.Transport = NewDebugTransport(httpClient.Transport)

	authProvider, err := securityprovider.NewSecurityProviderBearerToken(token)
	if err != nil {
		return nil, fmt.Errorf("Unable to Create Storyblok API Client %s", err.Error())
	}

	setVersionHeader := func(ctx context.Context, req *http.Request) error {
		if req.Header.Get("Content-Type") == "application/json" {
			req.Header.Set("Content-Type", "application/vnd.contentful.management.v1+json")
		}
		return nil
	}

	client, err := sdk.NewClientWithResponses(
		url,
		sdk.WithHTTPClient(httpClient),
		sdk.WithRequestEditorFn(setVersionHeader),
		sdk.WithRequestEditorFn(authProvider.Intercept))

	if err != nil {
		return nil, err
	}

	return client, nil

}

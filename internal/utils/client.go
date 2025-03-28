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

type Response interface {
	StatusCode() int
}

// Use a generic constraint on the value type, not a pointer
func CheckClientResponse(resp Response, err error, statusCode int) error {

	if err != nil {
		return fmt.Errorf("Error while interacting with Contentful API: %w", err)
	}

	if resp.StatusCode() == http.StatusConflict {
		return fmt.Errorf("Conflict while interacting with Contentful API: 409 Conflict")
	}

	if resp.StatusCode() != statusCode {
		return fmt.Errorf("Unexpected HTTP status code, expected %d, got %d", statusCode, resp.StatusCode())
	}

	return nil
}

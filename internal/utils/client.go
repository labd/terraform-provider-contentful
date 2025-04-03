package utils

import (
	"context"
	"fmt"
	"net/http"
	"reflect"

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

	if resp.StatusCode() == statusCode {
		return nil
	}

	if err := ExtractErrorResponse(resp); err != nil {
		return err
	}

	if resp.StatusCode() == http.StatusConflict {
		return fmt.Errorf("Conflict while interacting with Contentful API: 409 Conflict")
	}

	return fmt.Errorf("Unexpected response from Contentful API: %d (expected: %d)", resp.StatusCode(), statusCode)
}

func ExtractErrorResponse(resp Response) error {

	// Use reflection to check if the response has a JSON422 field
	respValue := reflect.ValueOf(resp)

	// Handle pointer types by dereferencing
	if respValue.Kind() == reflect.Ptr {
		respValue = respValue.Elem()
	}

	// Extract the body
	body := respValue.FieldByName("Body")
	if body.IsValid() && !body.IsZero() {
		value := body.Interface()
		if v, ok := value.([]byte); ok {
			return fmt.Errorf("response from Contentful API (%d):\n\n  %s", resp.StatusCode(), string(v))
		}
	}

	return nil
}

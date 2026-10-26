package environment

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/labd/terraform-provider-contentful/internal/sdk"
)

const (
	environmentStatusReady   = "ready"
	environmentStatusFailure = "failure"
)

var (
	environmentReadyTimeout      = 5 * time.Minute
	environmentReadyPollInterval = time.Second
)

type environmentGetter interface {
	GetEnvironmentWithResponse(ctx context.Context, spaceId sdk.SpaceId, environmentId sdk.EnvironmentId, reqEditors ...sdk.RequestEditorFn) (*sdk.GetEnvironmentResponse, error)
}

// environmentStatus is the part of the environment response that the
// generated SDK does not model: sys.status.sys.id (queued, ready, failure).
type environmentStatus struct {
	Sys struct {
		Status *struct {
			Sys struct {
				ID string `json:"id"`
			} `json:"sys"`
		} `json:"status"`
	} `json:"sys"`
}

// waitForEnvironmentReady polls the environment until Contentful finished
// creating it. While an environment is still being created its version keeps
// changing, so updating or deleting it with the version returned on create
// fails with a 409 version mismatch.
func waitForEnvironmentReady(ctx context.Context, client environmentGetter, spaceID, environmentID string) (*sdk.Environment, error) {
	ctx, cancel := context.WithTimeout(ctx, environmentReadyTimeout)
	defer cancel()

	for {
		resp, err := client.GetEnvironmentWithResponse(ctx, spaceID, environmentID)
		if err != nil {
			if ctx.Err() != nil {
				return nil, fmt.Errorf("timed out waiting for environment %s to become ready: %w", environmentID, ctx.Err())
			}
			return nil, fmt.Errorf("could not read environment: %w", err)
		}

		// A freshly created environment can briefly return 404
		if resp.StatusCode() != http.StatusNotFound {
			if resp.StatusCode() != http.StatusOK || resp.JSON200 == nil {
				return nil, fmt.Errorf("received unexpected status code: %d", resp.StatusCode())
			}

			var status environmentStatus
			if err := json.Unmarshal(resp.Body, &status); err != nil {
				return nil, fmt.Errorf("could not parse environment status: %w", err)
			}

			// Environments without a status (e.g. master) are always ready
			if status.Sys.Status == nil {
				return resp.JSON200, nil
			}

			switch status.Sys.Status.Sys.ID {
			case environmentStatusReady:
				return resp.JSON200, nil
			case environmentStatusFailure:
				return nil, fmt.Errorf("environment %s failed to be created", environmentID)
			}
		}

		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("timed out waiting for environment %s to become ready: %w", environmentID, ctx.Err())
		case <-time.After(environmentReadyPollInterval):
		}
	}
}

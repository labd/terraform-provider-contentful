package environment

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/labd/terraform-provider-contentful/internal/sdk"
)

func environmentBody(version int, status string) string {
	statusJSON := ""
	if status != "" {
		statusJSON = fmt.Sprintf(`,"status":{"sys":{"type":"Link","linkType":"Status","id":%q}}`, status)
	}
	return fmt.Sprintf(`{"name":"env","sys":{"id":"env","type":"Environment","version":%d,"space":{"sys":{"type":"Link","linkType":"Space","id":"space"}},"createdBy":{"sys":{"type":"Link","linkType":"User","id":"user"}}%s}}`, version, statusJSON)
}

func newTestClient(t *testing.T, responses []string) (*sdk.ClientWithResponses, *int) {
	t.Helper()
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/spaces/space/environments/env", r.URL.Path)
		body := responses[min(calls, len(responses)-1)]
		calls++
		w.Header().Set("Content-Type", "application/vnd.contentful.management.v1+json")
		if body == "" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(server.Close)

	client, err := sdk.NewClientWithResponses(server.URL)
	require.NoError(t, err)
	return client, &calls
}

func withFastPolling(t *testing.T, timeout time.Duration) {
	t.Helper()
	interval, origTimeout := environmentReadyPollInterval, environmentReadyTimeout
	environmentReadyPollInterval, environmentReadyTimeout = time.Millisecond, timeout
	t.Cleanup(func() {
		environmentReadyPollInterval, environmentReadyTimeout = interval, origTimeout
	})
}

func TestWaitForEnvironmentReady_PollsUntilReady(t *testing.T) {
	withFastPolling(t, time.Minute)
	client, calls := newTestClient(t, []string{
		"",
		environmentBody(1, "queued"),
		environmentBody(2, "queued"),
		environmentBody(3, "ready"),
	})

	env, err := waitForEnvironmentReady(context.Background(), client, "space", "env")
	require.NoError(t, err)
	assert.EqualValues(t, 3, env.Sys.Version)
	assert.Equal(t, 4, *calls)
}

func TestWaitForEnvironmentReady_NoStatus(t *testing.T) {
	withFastPolling(t, time.Minute)
	client, calls := newTestClient(t, []string{environmentBody(5, "")})

	env, err := waitForEnvironmentReady(context.Background(), client, "space", "env")
	require.NoError(t, err)
	assert.EqualValues(t, 5, env.Sys.Version)
	assert.Equal(t, 1, *calls)
}

func TestWaitForEnvironmentReady_Failure(t *testing.T) {
	withFastPolling(t, time.Minute)
	client, _ := newTestClient(t, []string{environmentBody(1, "queued"), environmentBody(2, "failure")})

	_, err := waitForEnvironmentReady(context.Background(), client, "space", "env")
	assert.ErrorContains(t, err, "failed to be created")
}

func TestWaitForEnvironmentReady_Timeout(t *testing.T) {
	withFastPolling(t, 20*time.Millisecond)
	client, _ := newTestClient(t, []string{environmentBody(1, "queued")})

	_, err := waitForEnvironmentReady(context.Background(), client, "space", "env")
	assert.ErrorContains(t, err, "timed out waiting for environment env")
}

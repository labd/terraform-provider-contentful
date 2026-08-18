package webhook_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/labd/terraform-provider-contentful/internal/resources/webhook"
)

// A failed API call leaves the parsed response body empty; mapping it should
// return an error instead of panicking on a nil dereference.
func TestWebhookMapFromSDKNil(t *testing.T) {
	w := &webhook.Webhook{}
	err := w.MapFromSDK(nil)
	assert.Error(t, err)
}

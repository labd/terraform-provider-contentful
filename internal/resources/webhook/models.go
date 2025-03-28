package webhook

import (
	"github.com/elliotchance/pie/v2"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/labd/terraform-provider-contentful/internal/sdk"
	"github.com/labd/terraform-provider-contentful/internal/utils"
)

// Webhook is the main resource schema data
type Webhook struct {
	ID                    types.String            `tfsdk:"id"`
	SpaceId               types.String            `tfsdk:"space_id"`
	Version               types.Int64             `tfsdk:"version"`
	Name                  types.String            `tfsdk:"name"`
	URL                   types.String            `tfsdk:"url"`
	HttpBasicAuthUsername types.String            `tfsdk:"http_basic_auth_username"`
	HttpBasicAuthPassword types.String            `tfsdk:"http_basic_auth_password"`
	Headers               map[string]types.String `tfsdk:"headers"`
	Topics                []types.String          `tfsdk:"topics"`
}

// Import populates the Webhook struct from an SDK webhook object
func (w *Webhook) Import(webhook *sdk.Webhook) {
	w.ID = types.StringValue(*webhook.Sys.Id)
	w.SpaceId = types.StringValue(webhook.Sys.Space.Sys.Id)
	w.Version = types.Int64Value(int64(*webhook.Sys.Version))
	w.Name = types.StringValue(webhook.Name)
	w.URL = types.StringValue(webhook.Url)

	// Handle nullable fields with appropriate defaults
	w.HttpBasicAuthUsername = types.StringValue("")
	if webhook.HttpBasicUsername != nil {
		w.HttpBasicAuthUsername = types.StringValue(*webhook.HttpBasicUsername)
	}

	// Convert headers
	w.Headers = make(map[string]types.String)
	if webhook.Headers != nil {
		for _, header := range webhook.Headers {
			w.Headers[header.Key] = types.StringValue(header.Value)
		}
	}

	// Convert topics
	w.Topics = pie.Map(webhook.Topics, func(t string) types.String {
		return types.StringValue(t)
	})
}

// DraftForCreate creates a WebhookCreate object for creating a new webhook
func (w *Webhook) DraftForCreate() sdk.WebhookCreate {
	return sdk.WebhookCreate{
		Name:              w.Name.ValueString(),
		Url:               w.URL.ValueString(),
		Topics:            w.topicsToSDK(),
		Headers:           w.headersToSDK(),
		HttpBasicUsername: utils.Pointer(w.HttpBasicAuthUsername.ValueString()),
		HttpBasicPassword: utils.Pointer(w.HttpBasicAuthPassword.ValueString()),
	}
}

// DraftForUpdate creates a WebhookUpdate object for updating an existing webhook
func (w *Webhook) DraftForUpdate() sdk.WebhookUpdate {
	return sdk.WebhookUpdate{
		Name:              w.Name.ValueString(),
		Url:               w.URL.ValueString(),
		Topics:            w.topicsToSDK(),
		Headers:           w.headersToSDK(),
		HttpBasicUsername: utils.Pointer(w.HttpBasicAuthUsername.ValueString()),
		HttpBasicPassword: utils.Pointer(w.HttpBasicAuthPassword.ValueString()),
	}
}

// Convert topics from Terraform types to SDK string slice
func (w *Webhook) topicsToSDK() []string {
	return pie.Map(w.Topics, func(t types.String) string {
		return t.ValueString()
	})
}

// Convert headers from Terraform map to SDK WebhookHeader slice
func (w *Webhook) headersToSDK() *[]sdk.WebhookHeader {
	var headers []sdk.WebhookHeader

	for key, value := range w.Headers {
		headers = append(headers, sdk.WebhookHeader{
			Key:   key,
			Value: value.ValueString(),
		})
	}

	if len(headers) == 0 {
		return nil
	}

	return &headers
}

package webhook

import (
	"encoding/json"
	"errors"

	"github.com/elliotchance/pie/v2"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/labd/terraform-provider-contentful/internal/sdk"
	"github.com/labd/terraform-provider-contentful/internal/utils"
)

// WebhookHeader represents a single webhook HTTP header with an optional secret flag.
type WebhookHeader struct {
	Key    types.String `tfsdk:"key"`
	Value  types.String `tfsdk:"value"`
	Secret types.Bool   `tfsdk:"secret"`
}

// Webhook is the main resource schema data
type Webhook struct {
	ID                    types.String    `tfsdk:"id"`
	SpaceId               types.String    `tfsdk:"space_id"`
	Version               types.Int64     `tfsdk:"version"`
	Name                  types.String    `tfsdk:"name"`
	URL                   types.String    `tfsdk:"url"`
	HttpBasicAuthUsername types.String    `tfsdk:"http_basic_auth_username"`
	HttpBasicAuthPassword types.String    `tfsdk:"http_basic_auth_password"`
	Headers               []WebhookHeader `tfsdk:"headers"`
	Topics                []types.String  `tfsdk:"topics"`
	Filters               types.String    `tfsdk:"filters"`
	Active                types.Bool      `tfsdk:"active"`
}

// MapFromSDK populates the Webhook struct from an SDK webhook object
func (w *Webhook) MapFromSDK(webhook *sdk.Webhook) error {
	w.ID = types.StringValue(*webhook.Sys.Id)
	w.SpaceId = types.StringValue(webhook.Sys.Space.Sys.Id)
	w.Version = types.Int64Value(int64(*webhook.Sys.Version))
	w.Name = types.StringValue(webhook.Name)
	w.URL = types.StringValue(webhook.Url)

	// Handle nullable fields with appropriate defaults
	w.HttpBasicAuthUsername = types.StringNull()
	if webhook.HttpBasicUsername != nil {
		w.HttpBasicAuthUsername = types.StringValue(*webhook.HttpBasicUsername)
	}

	// Convert headers
	w.Headers = make([]WebhookHeader, 0)
	if webhook.Headers != nil {
		for _, header := range webhook.Headers {
			secret := header.Secret != nil && *header.Secret
			w.Headers = append(w.Headers, WebhookHeader{
				Key:    types.StringValue(header.Key),
				Value:  types.StringValue(header.Value),
				Secret: types.BoolValue(secret),
			})
		}
	}

	// Convert topics
	w.Topics = pie.Map(webhook.Topics, func(t string) types.String {
		return types.StringValue(t)
	})

	w.Active = types.BoolPointerValue(webhook.Active)

	var filters string
	if webhook.Filters != nil {
		filtersBytes, err := json.Marshal(webhook.Filters)
		if err != nil {
			return errors.New("failed to marshal filters: " + err.Error())
		}
		filters = string(filtersBytes)
	}

	w.Filters = types.StringPointerValue(&filters)

	return nil
}

// PreserveSecretHeaderValues preserves the values (and secret flags) of secret headers from the prior state,
// since the Contentful API does not return secret header values in GET responses.
// If the prior state marked a header as secret but the API response does not include the secret flag,
// the prior secret flag and value are preserved to avoid spurious diffs.
// Note: duplicate header keys are not supported; if multiple prior headers share the same key,
// only the last one is used for value preservation.
func (w *Webhook) PreserveSecretHeaderValues(priorHeaders []WebhookHeader) {
	// Build a lookup map from the prior state headers by key
	priorByKey := make(map[string]WebhookHeader, len(priorHeaders))
	for _, h := range priorHeaders {
		priorByKey[h.Key.ValueString()] = h
	}

	for i, h := range w.Headers {
		prior, ok := priorByKey[h.Key.ValueString()]
		if !ok {
			continue
		}
		// Preserve secret flag and value when either the current API response or the
		// prior state marks the header as secret. This handles cases where the API
		// does not return the secret flag explicitly (nil → false) even though it was set.
		if h.Secret.ValueBool() || prior.Secret.ValueBool() {
			w.Headers[i].Secret = prior.Secret
			w.Headers[i].Value = prior.Value
		}
	}
}

func (w *Webhook) DraftForCreate() (sdk.WebhookCreate, error) {
	filters, err := w.filtersToSdk()
	if err != nil {
		return sdk.WebhookCreate{}, err
	}

	var draft = sdk.WebhookCreate{
		Name:    w.Name.ValueString(),
		Url:     w.URL.ValueString(),
		Topics:  w.topicsToSDK(),
		Headers: w.headersToSDK(),
		Active:  w.Active.ValueBoolPointer(),
		Filters: filters,
	}

	if !w.HttpBasicAuthPassword.IsNull() {
		draft.HttpBasicPassword = utils.Pointer(w.HttpBasicAuthPassword.ValueString())
	}
	if !w.HttpBasicAuthUsername.IsNull() {
		draft.HttpBasicUsername = utils.Pointer(w.HttpBasicAuthUsername.ValueString())
	}

	return draft, nil
}

// DraftForUpdate creates a WebhookUpdate object for updating an existing webhook
func (w *Webhook) DraftForUpdate() (sdk.WebhookUpdate, error) {
	filters, err := w.filtersToSdk()
	if err != nil {
		return sdk.WebhookUpdate{}, err
	}

	var update = sdk.WebhookUpdate{
		Name:    w.Name.ValueString(),
		Url:     w.URL.ValueString(),
		Topics:  w.topicsToSDK(),
		Headers: w.headersToSDK(),
		Active:  w.Active.ValueBoolPointer(),
		Filters: filters,
	}

	if !w.HttpBasicAuthPassword.IsNull() {
		update.HttpBasicPassword = utils.Pointer(w.HttpBasicAuthPassword.ValueString())
	}
	if !w.HttpBasicAuthUsername.IsNull() {
		update.HttpBasicUsername = utils.Pointer(w.HttpBasicAuthUsername.ValueString())
	}

	return update, nil
}

// Convert topics from Terraform types to SDK string slice
func (w *Webhook) topicsToSDK() []string {
	return pie.Map(w.Topics, func(t types.String) string {
		return t.ValueString()
	})
}

// Convert filters from Terraform types to SDK string slice
func (w *Webhook) filtersToSdk() (*[]map[string]interface{}, error) {
	var filterContent = make([]map[string]interface{}, 0)

	filter := w.Filters.ValueString()
	if filter == "" {
		return &filterContent, nil
	}

	err := json.Unmarshal([]byte(filter), &filterContent)
	if err != nil {
		return nil, errors.New("failed to unmarshal filters: " + err.Error())
	}

	return &filterContent, nil
}

// Convert headers from Terraform list to SDK WebhookHeader slice
func (w *Webhook) headersToSDK() *[]sdk.WebhookHeader {
	var headers []sdk.WebhookHeader

	for _, header := range w.Headers {
		h := sdk.WebhookHeader{
			Key:   header.Key.ValueString(),
			Value: header.Value.ValueString(),
		}
		if !header.Secret.IsNull() && !header.Secret.IsUnknown() {
			h.Secret = utils.Pointer(header.Secret.ValueBool())
		}
		headers = append(headers, h)
	}

	if len(headers) == 0 {
		return nil
	}

	return &headers
}

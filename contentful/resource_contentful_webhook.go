package contentful

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/labd/terraform-provider-contentful/internal/sdk"
	"github.com/labd/terraform-provider-contentful/internal/utils"
)

func resourceContentfulWebhook() *schema.Resource {
	return &schema.Resource{
		Description: "A Contentful Webhook represents a webhook that can be used to notify external services of changes in a space.",

		CreateContext: resourceCreateWebhook,
		ReadContext:   resourceReadWebhook,
		UpdateContext: resourceUpdateWebhook,
		DeleteContext: resourceDeleteWebhook,

		Schema: map[string]*schema.Schema{
			"version": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"space_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			// Webhook specific props
			"url": {
				Type:     schema.TypeString,
				Required: true,
			},
			"http_basic_auth_username": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},
			"http_basic_auth_password": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},
			"headers": {
				Type:     schema.TypeMap,
				Optional: true,
			},
			"topics": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				MinItems: 1,
				Required: true,
			},
		},
	}
}

func resourceCreateWebhook(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(utils.ProviderData).Client
	spaceID := d.Get("space_id").(string)

	body := sdk.WebhookCreate{
		Name:              d.Get("name").(string),
		Url:               d.Get("url").(string),
		Topics:            transformTopicsToContentfulFormat(d.Get("topics").([]any)),
		Headers:           transformHeadersToContentfulFormat(d.Get("headers")),
		HttpBasicUsername: utils.Pointer(d.Get("http_basic_auth_username").(string)),
		HttpBasicPassword: utils.Pointer(d.Get("http_basic_auth_password").(string)),
	}

	resp, err := client.CreateWebhookWithResponse(ctx, spaceID, body)
	if err != nil {
		return parseError(err)
	}

	if resp.StatusCode() != 201 {
		return diag.Errorf("Failed to create webhook")
	}

	webhook := resp.JSON201
	err = setWebhookProperties(d, webhook)
	if err != nil {
		return parseError(err)
	}

	d.SetId(*webhook.Sys.Id)

	return nil
}

func resourceUpdateWebhook(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(utils.ProviderData).Client
	spaceID := d.Get("space_id").(string)
	webhookID := d.Id()

	resp, err := client.GetWebhookWithResponse(ctx, spaceID, webhookID)
	if err != nil {
		return parseError(err)
	}

	if resp.StatusCode() != 200 {
		return diag.Errorf("Webhook not found")
	}

	params := &sdk.UpdateWebhookParams{
		XContentfulVersion: int64(d.Get("version").(int)),
	}

	update := sdk.WebhookUpdate{
		Name:              d.Get("name").(string),
		Url:               d.Get("url").(string),
		Topics:            transformTopicsToContentfulFormat(d.Get("topics").([]any)),
		Headers:           transformHeadersToContentfulFormat(d.Get("headers")),
		HttpBasicUsername: utils.Pointer(d.Get("http_basic_auth_username").(string)),
		HttpBasicPassword: utils.Pointer(d.Get("http_basic_auth_password").(string)),
	}

	updateResp, err := client.UpdateWebhookWithResponse(ctx, spaceID, webhookID, params, update)
	if err != nil {
		return parseError(err)
	}

	if updateResp.StatusCode() != 200 {
		return diag.Errorf("Failed to update webhook")
	}

	webhook := updateResp.JSON200
	err = setWebhookProperties(d, webhook)
	if err != nil {
		return parseError(err)
	}

	d.SetId(*webhook.Sys.Id)

	return nil
}

func resourceReadWebhook(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(utils.ProviderData).Client
	spaceID := d.Get("space_id").(string)
	webhookID := d.Id()

	resp, err := client.GetWebhookWithResponse(ctx, spaceID, webhookID)
	if err != nil {
		return parseError(err)
	}

	if resp.StatusCode() != 200 {
		d.SetId("")
		return nil
	}

	webhook := resp.JSON200
	err = setWebhookProperties(d, webhook)
	if err != nil {
		return parseError(err)
	}

	return nil
}

func resourceDeleteWebhook(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(utils.ProviderData).Client
	spaceID := d.Get("space_id").(string)
	webhookID := d.Id()

	params := &sdk.DeleteWebhookParams{
		XContentfulVersion: int64(d.Get("version").(int)),
	}

	resp, err := client.DeleteWebhookWithResponse(ctx, spaceID, webhookID, params)
	if err != nil {
		return parseError(err)
	}

	if resp.StatusCode() != 204 {
		return diag.Errorf("Failed to delete webhook")
	}

	return nil
}

func setWebhookProperties(d *schema.ResourceData, webhook *sdk.Webhook) (err error) {
	headers := make(map[string]string)
	for _, entry := range webhook.Headers {
		headers[entry.Key] = entry.Value
	}

	err = d.Set("headers", headers)
	if err != nil {
		return err
	}

	err = d.Set("space_id", webhook.Sys.Space.Sys.Id)
	if err != nil {
		return err
	}

	err = d.Set("version", webhook.Sys.Version)
	if err != nil {
		return err
	}

	err = d.Set("name", webhook.Name)
	if err != nil {
		return err
	}

	err = d.Set("url", webhook.Url)
	if err != nil {
		return err
	}

	err = d.Set("http_basic_auth_username", webhook.HttpBasicUsername)
	if err != nil {
		return err
	}

	err = d.Set("topics", webhook.Topics)
	if err != nil {
		return err
	}

	return nil
}

func transformHeadersToContentfulFormat(headersTerraform any) *[]sdk.WebhookHeader {
	var headers []sdk.WebhookHeader

	for k, v := range headersTerraform.(map[string]interface{}) {
		headers = append(headers, sdk.WebhookHeader{
			Key:   k,
			Value: v.(string),
		})
	}
	if len(headers) == 0 {
		return nil
	}

	return &headers
}

func transformTopicsToContentfulFormat(topicsTerraform []interface{}) []string {
	var topics []string

	for _, v := range topicsTerraform {
		topics = append(topics, v.(string))
	}

	return topics
}

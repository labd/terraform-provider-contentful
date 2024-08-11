package contentful

import (
	"context"
	"errors"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/labd/contentful-go"
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
	time.Sleep(consistencyBuffer)
	client := m.(*contentful.Client)
	spaceID := d.Get("space_id").(string)

	webhook := &contentful.Webhook{
		Name:              d.Get("name").(string),
		URL:               d.Get("url").(string),
		Topics:            transformTopicsToContentfulFormat(d.Get("topics").([]interface{})),
		Headers:           transformHeadersToContentfulFormat(d.Get("headers")),
		HTTPBasicUsername: d.Get("http_basic_auth_username").(string),
		HTTPBasicPassword: d.Get("http_basic_auth_password").(string),
	}

	err := client.Webhooks.Upsert(spaceID, webhook)
	if err != nil {
		return parseError(err)
	}

	err = setWebhookProperties(d, webhook)
	if err != nil {
		return parseError(err)
	}

	d.SetId(webhook.Sys.ID)

	return nil
}

func resourceUpdateWebhook(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	time.Sleep(consistencyBuffer)
	client := m.(*contentful.Client)
	spaceID := d.Get("space_id").(string)
	webhookID := d.Id()

	webhook, err := client.Webhooks.Get(spaceID, webhookID)
	if err != nil {
		return parseError(err)
	}

	webhook.Name = d.Get("name").(string)
	webhook.URL = d.Get("url").(string)
	webhook.Topics = transformTopicsToContentfulFormat(d.Get("topics").([]interface{}))
	webhook.Headers = transformHeadersToContentfulFormat(d.Get("headers"))
	webhook.HTTPBasicUsername = d.Get("http_basic_auth_username").(string)
	webhook.HTTPBasicPassword = d.Get("http_basic_auth_password").(string)

	err = client.Webhooks.Upsert(spaceID, webhook)
	if err != nil {
		return parseError(err)
	}

	err = setWebhookProperties(d, webhook)
	if err != nil {
		return parseError(err)
	}

	d.SetId(webhook.Sys.ID)

	return nil
}

func resourceReadWebhook(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	time.Sleep(consistencyBuffer)
	client := m.(*contentful.Client)
	spaceID := d.Get("space_id").(string)
	webhookID := d.Id()

	webhook, err := client.Webhooks.Get(spaceID, webhookID)
	var notFoundError contentful.NotFoundError
	if errors.As(err, &notFoundError) {
		d.SetId("")
		return nil
	}

	if err != nil {
		return parseError(err)
	}

	err = setWebhookProperties(d, webhook)
	if err != nil {
		return parseError(err)
	}

	return nil
}

func resourceDeleteWebhook(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	time.Sleep(consistencyBuffer)
	client := m.(*contentful.Client)
	spaceID := d.Get("space_id").(string)
	webhookID := d.Id()

	webhook, err := client.Webhooks.Get(spaceID, webhookID)
	if err != nil {
		return parseError(err)
	}

	err = client.Webhooks.Delete(spaceID, webhook)
	if _, ok := err.(contentful.NotFoundError); ok {
		return nil
	}

	return parseError(err)
}

func setWebhookProperties(d *schema.ResourceData, webhook *contentful.Webhook) (err error) {
	headers := make(map[string]string)
	for _, entry := range webhook.Headers {
		headers[entry.Key] = entry.Value
	}

	err = d.Set("headers", headers)
	if err != nil {
		return err
	}

	err = d.Set("space_id", webhook.Sys.Space.Sys.ID)
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

	err = d.Set("url", webhook.URL)
	if err != nil {
		return err
	}

	err = d.Set("http_basic_auth_username", webhook.HTTPBasicUsername)
	if err != nil {
		return err
	}

	err = d.Set("topics", webhook.Topics)
	if err != nil {
		return err
	}

	return nil
}

func transformHeadersToContentfulFormat(headersTerraform interface{}) []*contentful.WebhookHeader {
	var headers []*contentful.WebhookHeader

	for k, v := range headersTerraform.(map[string]interface{}) {
		headers = append(headers, &contentful.WebhookHeader{
			Key:   k,
			Value: v.(string),
		})
	}

	return headers
}

func transformTopicsToContentfulFormat(topicsTerraform []interface{}) []string {
	var topics []string

	for _, v := range topicsTerraform {
		topics = append(topics, v.(string))
	}

	return topics
}

package contentful

import (
	"context"
	"errors"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/labd/contentful-go"
)

func resourceContentfulAPIKey() *schema.Resource {
	return &schema.Resource{
		Description: "A Contentful API Key represents a token that can be used to authenticate against the Contentful Content Delivery API and Content Preview API.",

		CreateContext: resourceCreateAPIKey,
		ReadContext:   resourceReadAPIKey,
		UpdateContext: resourceUpdateAPIKey,
		DeleteContext: resourceDeleteAPIKey,

		Schema: map[string]*schema.Schema{
			"version": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"access_token": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"space_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"environment_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceCreateAPIKey(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	time.Sleep(consistencyBuffer)
	client := m.(*contentful.Client)

	apiKey := &contentful.APIKey{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
	}

	environmentID := d.Get("environment_id").(string)
	if environmentID != "" {
		apiKey.Environments = append(apiKey.Environments, environmentLink(environmentID))
	}

	err := client.APIKeys.Upsert(d.Get("space_id").(string), apiKey)
	if err != nil {
		return parseError(err)
	}

	if err := setAPIKeyProperties(d, apiKey); err != nil {
		return parseError(err)
	}

	d.SetId(apiKey.Sys.ID)

	return nil
}

func resourceUpdateAPIKey(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	time.Sleep(consistencyBuffer)
	client := m.(*contentful.Client)
	spaceID := d.Get("space_id").(string)
	apiKeyID := d.Id()

	apiKey, err := client.APIKeys.Get(spaceID, apiKeyID)
	if err != nil {
		return parseError(err)
	}

	apiKey.Name = d.Get("name").(string)
	apiKey.Description = d.Get("description").(string)

	err = client.APIKeys.Upsert(spaceID, apiKey)
	if err != nil {
		return parseError(err)
	}

	if err := setAPIKeyProperties(d, apiKey); err != nil {
		return parseError(err)
	}

	d.SetId(apiKey.Sys.ID)

	return nil
}

func resourceReadAPIKey(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	time.Sleep(consistencyBuffer)
	client := m.(*contentful.Client)
	spaceID := d.Get("space_id").(string)
	apiKeyID := d.Id()

	apiKey, err := client.APIKeys.Get(spaceID, apiKeyID)
	var notFoundError contentful.NotFoundError
	if errors.As(err, &notFoundError) {
		d.SetId("")
		return nil
	}

	err = setAPIKeyProperties(d, apiKey)
	if err != nil {
		return parseError(err)
	}

	return nil
}

func resourceDeleteAPIKey(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	time.Sleep(consistencyBuffer)
	client := m.(*contentful.Client)
	spaceID := d.Get("space_id").(string)
	apiKeyID := d.Id()

	apiKey, err := client.APIKeys.Get(spaceID, apiKeyID)
	if err != nil {
		return parseError(err)
	}

	err = client.APIKeys.Delete(spaceID, apiKey)
	if err != nil {
		return parseError(err)
	}

	return nil
}

func setAPIKeyProperties(d *schema.ResourceData, apiKey *contentful.APIKey) error {
	if err := d.Set("space_id", apiKey.Sys.Space.Sys.ID); err != nil {
		return err
	}

	if err := d.Set("version", apiKey.Sys.Version); err != nil {
		return err
	}

	if err := d.Set("name", apiKey.Name); err != nil {
		return err
	}

	if err := d.Set("description", apiKey.Description); err != nil {
		return err
	}

	if err := d.Set("access_token", apiKey.AccessToken); err != nil {
		return err
	}

	return nil
}

func environmentLink(environmentID string) contentful.Environments {
	return contentful.Environments{
		Sys: contentful.Sys{
			Type:     "Link",
			LinkType: "Environment",
			ID:       environmentID,
		},
	}
}

package contentful

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/labd/contentful-go"
	"github.com/labd/contentful-go/pkgs/common"

	"github.com/labd/terraform-provider-contentful/internal/utils"
)

func resourceContentfulSpace() *schema.Resource {
	return &schema.Resource{
		Description: "A Contentful Space represents a space in Contentful.",

		CreateContext: resourceSpaceCreate,
		ReadContext:   resourceSpaceRead,
		UpdateContext: resourceSpaceUpdate,
		DeleteContext: resourceSpaceDelete,

		Schema: map[string]*schema.Schema{
			"version": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			// Space specific props
			"default_locale": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "en",
			},
		},
	}
}

func resourceSpaceCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(utils.ProviderData).Client

	space := &contentful.Space{
		Name:          d.Get("name").(string),
		DefaultLocale: d.Get("default_locale").(string),
	}

	err := client.Spaces.Upsert(space)
	if err != nil {
		return parseError(err)
	}

	err = updateSpaceProperties(d, space)
	if err != nil {
		return parseError(err)
	}

	d.SetId(space.Sys.ID)

	return nil
}

func resourceSpaceRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(utils.ProviderData).Client
	spaceID := d.Id()

	_, err := client.Spaces.Get(spaceID)
	if _, ok := err.(common.NotFoundError); ok {
		d.SetId("")
		return nil
	}

	return parseError(err)
}

func resourceSpaceUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(utils.ProviderData).Client
	spaceID := d.Id()

	space, err := client.Spaces.Get(spaceID)
	if err != nil {
		return parseError(err)
	}

	space.Name = d.Get("name").(string)

	err = client.Spaces.Upsert(space)
	if err != nil {
		return parseError(err)
	}

	err = updateSpaceProperties(d, space)
	if err != nil {
		return parseError(err)
	}

	return nil
}

func resourceSpaceDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(utils.ProviderData).Client
	spaceID := d.Id()

	space, err := client.Spaces.Get(spaceID)
	if err != nil {
		return parseError(err)
	}

	err = client.Spaces.Delete(space)
	if _, ok := err.(common.NotFoundError); ok {
		return nil
	}

	return parseError(err)
}

func updateSpaceProperties(d *schema.ResourceData, space *contentful.Space) error {
	err := d.Set("version", space.Sys.Version)
	if err != nil {
		return err
	}

	err = d.Set("name", space.Name)
	if err != nil {
		return err
	}

	return nil
}

package contentful

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/labd/terraform-provider-contentful/internal/sdk"
	"github.com/labd/terraform-provider-contentful/internal/utils"
)

func resourceContentfulSpace() *schema.Resource {
	return &schema.Resource{
		Description: "A Contentful Space represents a space in Contentful.",

		CreateContext: resourceSpaceCreate,
		ReadContext:   resourceSpaceRead,
		UpdateContext: resourceSpaceUpdate,
		DeleteContext: resourceSpaceDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

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

	body := sdk.SpaceCreate{
		Name:          d.Get("name").(string),
		DefaultLocale: utils.Pointer(d.Get("default_locale").(string)),
	}

	resp, err := client.CreateSpaceWithResponse(ctx, nil, body)
	if err != nil {
		return parseError(err)
	}

	if resp.StatusCode() != 201 {
		return diag.Errorf("Failed to create space")
	}

	space := resp.JSON201

	err = updateSpaceProperties(d, space)
	if err != nil {
		return parseError(err)
	}

	d.SetId(space.Sys.Id)

	return nil
}

func resourceSpaceRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(utils.ProviderData).Client
	spaceID := d.Id()

	resp, err := client.GetSpaceWithResponse(ctx, spaceID)
	if err != nil {
		return parseError(err)
	}

	if resp.StatusCode() == 404 {
		d.SetId("")
		return nil
	}

	if resp.StatusCode() != 200 {
		return diag.Errorf("Failed to retrieve space")
	}

	err = updateSpaceProperties(d, resp.JSON200)
	if err != nil {
		return parseError(err)
	}

	return nil
}

func resourceSpaceUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(utils.ProviderData).Client
	spaceID := d.Id()

	getResp, err := client.GetSpaceWithResponse(ctx, spaceID)
	if err != nil {
		return parseError(err)
	}

	if getResp.StatusCode() != 200 {
		return diag.Errorf("Failed to retrieve space")
	}

	space := getResp.JSON200

	// TODO: we can't update the default locale here, we need to do that via the
	// locales endpoint, searching for the default locale
	update := sdk.SpaceUpdate{
		Name: d.Get("name").(string),
	}

	params := &sdk.UpdateSpaceParams{
		XContentfulVersion: int64(d.Get("version").(int)),
	}

	resp, err := client.UpdateSpaceWithResponse(ctx, space.Sys.Id, params, update)
	if err != nil {
		return parseError(err)
	}

	if resp.StatusCode() != 200 {
		return diag.Errorf("Failed to update space")
	}

	space = resp.JSON200

	err = updateSpaceProperties(d, space)
	if err != nil {
		return parseError(err)
	}

	return nil
}

func resourceSpaceDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(utils.ProviderData).Client
	spaceID := d.Id()

	getResp, err := client.GetSpaceWithResponse(ctx, spaceID)
	if err != nil {
		return parseError(err)
	}

	if getResp.StatusCode() != 200 {
		return diag.Errorf("Failed to retrieve space")
	}

	params := &sdk.DeleteSpaceParams{
		XContentfulVersion: int64(d.Get("version").(int)),
	}

	space := getResp.JSON200
	resp, err := client.DeleteSpaceWithResponse(ctx, space.Sys.Id, params)
	if err != nil {
		return parseError(err)
	}

	if resp.StatusCode() != 204 {
		return diag.Errorf("Failed to delete space")
	}

	return parseError(err)
}

func updateSpaceProperties(d *schema.ResourceData, space *sdk.Space) error {
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

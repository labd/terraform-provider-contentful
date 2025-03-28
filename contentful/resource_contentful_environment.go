package contentful

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/labd/terraform-provider-contentful/internal/sdk"
	"github.com/labd/terraform-provider-contentful/internal/utils"
)

func resourceContentfulEnvironment() *schema.Resource {
	return &schema.Resource{
		Description:   "A Contentful Environment represents a space environment.",
		CreateContext: resourceCreateEnvironment,
		ReadContext:   resourceReadEnvironment,
		UpdateContext: resourceUpdateEnvironment,
		DeleteContext: resourceDeleteEnvironment,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

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
		},
	}
}

func resourceCreateEnvironment(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(utils.ProviderData).Client
	spaceID := d.Get("space_id").(string)

	body := sdk.EnvironmentCreate{
		Name: d.Get("name").(string),
	}

	resp, err := client.CreateEnvironmentWithResponse(ctx, spaceID, body)
	if err != nil {
		return parseError(err)
	}

	if resp.StatusCode() != 201 {
		return diag.Errorf("Failed to create environment")
	}

	environment := resp.JSON201
	if err := setEnvironmentProperties(d, environment); err != nil {
		return parseError(err)
	}

	d.SetId(environment.Sys.Id)

	return nil
}

func resourceUpdateEnvironment(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(utils.ProviderData).Client
	spaceID := d.Get("space_id").(string)
	environmentID := d.Id()

	body := sdk.EnvironmentUpdate{
		Name: d.Get("name").(string),
	}

	params := &sdk.UpdateEnvironmentParams{
		XContentfulVersion: int64(d.Get("version").(int)),
	}

	resp, err := client.UpdateEnvironmentWithResponse(ctx, spaceID, environmentID, params, body)
	if err != nil {
		return parseError(err)
	}

	if resp.StatusCode() != 200 {
		return diag.Errorf("Failed to update environment")
	}

	environment := resp.JSON200
	if err := setEnvironmentProperties(d, environment); err != nil {
		return parseError(err)
	}

	d.SetId(environment.Sys.Id)

	return nil
}

func resourceReadEnvironment(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(utils.ProviderData).Client
	spaceID := d.Get("space_id").(string)
	environmentID := d.Id()

	resp, err := client.GetEnvironmentWithResponse(ctx, spaceID, environmentID)
	if err != nil {
		return parseError(err)
	}

	if resp.StatusCode() == 404 {
		d.SetId("")
		return nil
	}

	if resp.StatusCode() != 200 {
		return diag.Errorf("Failed to read environment")
	}

	environment := resp.JSON200
	err = setEnvironmentProperties(d, environment)
	if err != nil {
		return parseError(err)
	}

	return nil
}

func resourceDeleteEnvironment(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(utils.ProviderData).Client
	spaceID := d.Get("space_id").(string)
	environmentID := d.Id()

	params := &sdk.DeleteEnvironmentParams{
		XContentfulVersion: int64(d.Get("version").(int)),
	}

	resp, err := client.DeleteEnvironmentWithResponse(ctx, spaceID, environmentID, params)
	if err != nil {
		return parseError(err)
	}

	if resp.StatusCode() != 204 {
		return diag.Errorf("Failed to delete environment")
	}

	return nil
}

func setEnvironmentProperties(d *schema.ResourceData, environment *sdk.Environment) error {
	if err := d.Set("space_id", environment.Sys.Space.Sys.Id); err != nil {
		return err
	}

	if err := d.Set("version", environment.Sys.Version); err != nil {
		return err
	}

	if err := d.Set("name", environment.Name); err != nil {
		return err
	}

	return nil
}

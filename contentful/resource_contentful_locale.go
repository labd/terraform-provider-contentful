package contentful

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/labd/terraform-provider-contentful/internal/sdk"
	"github.com/labd/terraform-provider-contentful/internal/utils"
)

func resourceContentfulLocaleParseImportId(id string) (string, string, string, error) {
	parts := strings.SplitN(id, ":", 3)

	if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		return "", "", "", fmt.Errorf("unexpected format of ID (%s), expected localeId:env:spaceId", id)
	}

	return parts[0], parts[1], parts[2], nil
}

func resourceContentfulLocale() *schema.Resource {
	return &schema.Resource{
		Description: "A Contentful Locale represents a language and region combination.",

		CreateContext: resourceCreateLocale,
		ReadContext:   resourceReadLocale,
		UpdateContext: resourceUpdateLocale,
		DeleteContext: resourceDeleteLocale,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
				id, environment, spaceId, err := resourceContentfulLocaleParseImportId(d.Id())

				if err != nil {
					return nil, err
				}

				err = d.Set("environment", environment)
				if err != nil {
					return nil, err
				}

				err = d.Set("space_id", spaceId)
				if err != nil {
					return nil, err
				}
				d.SetId(id)

				return []*schema.ResourceData{d}, nil
			},
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
			"code": {
				Type:     schema.TypeString,
				Required: true,
			},
			"fallback_code": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"optional": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"cda": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"cma": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"environment": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceCreateLocale(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := m.(utils.ProviderData).Client
	spaceID := d.Get("space_id").(string)
	environmentID := d.Get("environment").(string)

	body := sdk.LocaleCreate{
		Name:                 d.Get("name").(string),
		Code:                 d.Get("code").(string),
		FallbackCode:         nil,
		Optional:             utils.Pointer(d.Get("optional").(bool)),
		ContentDeliveryApi:   utils.Pointer(d.Get("cda").(bool)),
		ContentManagementApi: utils.Pointer(d.Get("cma").(bool)),
	}

	if fallbackCode, ok := d.GetOk("fallback_code"); ok {
		fallbackCodeStr := fallbackCode.(string)
		body.FallbackCode = &fallbackCodeStr
	}

	resp, err := client.CreateLocaleWithResponse(ctx, spaceID, environmentID, body)
	if err != nil {
		return parseError(err)
	}

	if resp.StatusCode() != 201 {
		return diag.Errorf("Failed to create locale")
	}

	locale := resp.JSON201
	diagErr := setLocaleProperties(d, locale)
	if diagErr != nil {
		return diagErr
	}

	d.SetId(*locale.Sys.Id)

	return nil
}

func resourceReadLocale(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(utils.ProviderData).Client
	spaceID := d.Get("space_id").(string)
	environmentId := d.Get("environment").(string)
	localeId := d.Id()

	locale, err := getLocale(ctx, client, spaceID, environmentId, localeId)
	if err != nil {
		return parseError(err)
	}

	if locale == nil {
		d.SetId("")
		return nil
	}

	diagErr := setLocaleProperties(d, locale)
	if diagErr != nil {
		return diagErr
	}

	return nil
}

func resourceUpdateLocale(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(utils.ProviderData).Client
	spaceID := d.Get("space_id").(string)
	environmentId := d.Get("environment").(string)
	localeId := d.Id()

	params := &sdk.UpdateLocaleParams{
		XContentfulVersion: int64(d.Get("version").(int)),
	}
	body := sdk.LocaleUpdate{
		Name:                 d.Get("name").(string),
		Code:                 d.Get("code").(string),
		FallbackCode:         nil,
		Optional:             utils.Pointer(d.Get("optional").(bool)),
		ContentDeliveryApi:   utils.Pointer(d.Get("cda").(bool)),
		ContentManagementApi: utils.Pointer(d.Get("cma").(bool)),
	}

	if fallbackCode, ok := d.GetOk("fallback_code"); ok {
		fallbackCodeStr := fallbackCode.(string)
		body.FallbackCode = &fallbackCodeStr
	}

	resp, err := client.UpdateLocaleWithResponse(ctx, spaceID, environmentId, localeId, params, body)
	if err != nil {
		return parseError(err)
	}

	if resp.StatusCode() != 200 {
		return diag.Errorf("Failed to update locale")
	}

	locale := resp.JSON200
	diagErr := setLocaleProperties(d, locale)
	if diagErr != nil {
		return diagErr
	}

	return nil
}

func resourceDeleteLocale(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(utils.ProviderData).Client
	spaceID := d.Get("space_id").(string)
	environmentId := d.Get("environment").(string)
	localeId := d.Id()

	resp, err := client.DeleteLocaleWithResponse(ctx, spaceID, environmentId, localeId)
	if err != nil {
		return parseError(err)
	}

	if resp.StatusCode() != 204 {
		return diag.Errorf("Failed to delete locale")
	}

	return nil
}

func setLocaleProperties(d *schema.ResourceData, locale *sdk.Locale) diag.Diagnostics {
	err := d.Set("name", locale.Name)
	if err != nil {
		return diag.FromErr(err)
	}

	err = d.Set("code", locale.Code)
	if err != nil {
		return diag.FromErr(err)
	}

	err = d.Set("fallback_code", locale.FallbackCode)
	if err != nil {
		return diag.FromErr(err)
	}

	err = d.Set("optional", locale.Optional)
	if err != nil {
		return diag.FromErr(err)
	}

	err = d.Set("cda", locale.ContentDeliveryApi)
	if err != nil {
		return diag.FromErr(err)
	}

	err = d.Set("cma", locale.ContentManagementApi)
	if err != nil {
		return diag.FromErr(err)
	}

	err = d.Set("version", locale.Sys.Version)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func getLocale(ctx context.Context, client *sdk.ClientWithResponses, spaceID, environmentId, localeId string) (*sdk.Locale, error) {
	resp, err := client.GetLocaleWithResponse(ctx, spaceID, environmentId, localeId)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode() == 404 {
		return nil, nil
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("Failed to read locale")
	}

	return resp.JSON200, nil
}

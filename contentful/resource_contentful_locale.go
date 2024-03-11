package contentful

import (
	"context"
	"errors"
	"fmt"
	"github.com/flaconi/contentful-go/pkgs/common"
	"github.com/flaconi/contentful-go/pkgs/model"
	"github.com/flaconi/contentful-go/service/cma"
	"github.com/flaconi/terraform-provider-contentful/internal/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"strings"
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

func resourceCreateLocale(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(utils.ProviderData).CMAClient
	spaceID := d.Get("space_id").(string)
	localesClient := client.WithSpaceId(spaceID).WithEnvironment(d.Get("environment").(string)).Locales()
	var err error

	locale := &model.Locale{
		Name:         d.Get("name").(string),
		Code:         d.Get("code").(string),
		FallbackCode: nil,
		Optional:     d.Get("optional").(bool),
		CDA:          d.Get("cda").(bool),
		CMA:          d.Get("cma").(bool),
	}

	if fallbackCode, ok := d.GetOk("fallback_code"); ok {

		fallbackCodeStr := fallbackCode.(string)

		locale.FallbackCode = &fallbackCodeStr
	}

	err = localesClient.Upsert(context.Background(), locale)

	if err != nil {
		return diag.FromErr(err)
	}

	diagErr := setLocaleProperties(d, locale)
	if diagErr != nil {
		return diagErr
	}

	d.SetId(locale.Sys.ID)

	return nil
}

func resourceReadLocale(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(utils.ProviderData).CMAClient
	spaceID := d.Get("space_id").(string)
	localesClient := client.WithSpaceId(spaceID).WithEnvironment(d.Get("environment").(string)).Locales()

	locale, err := getLocale(d, localesClient)

	var notFoundError *common.NotFoundError
	if errors.As(err, &notFoundError) {
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.FromErr(err)
	}

	return setLocaleProperties(d, locale)
}

func resourceUpdateLocale(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(utils.ProviderData).CMAClient
	spaceID := d.Get("space_id").(string)
	localesClient := client.WithSpaceId(spaceID).WithEnvironment(d.Get("environment").(string)).Locales()

	locale, err := getLocale(d, localesClient)

	if err != nil {
		return diag.FromErr(err)
	}

	locale.Name = d.Get("name").(string)
	locale.Code = d.Get("code").(string)

	locale.FallbackCode = nil

	if fallbackCode, ok := d.GetOk("fallback_code"); ok {

		fallbackCodeStr := fallbackCode.(string)

		locale.FallbackCode = &fallbackCodeStr
	}

	locale.Optional = d.Get("optional").(bool)
	locale.CDA = d.Get("cda").(bool)
	locale.CMA = d.Get("cma").(bool)

	err = localesClient.Upsert(context.Background(), locale)

	if err != nil {
		return diag.FromErr(err)
	}

	return setLocaleProperties(d, locale)
}

func resourceDeleteLocale(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(utils.ProviderData).CMAClient
	spaceID := d.Get("space_id").(string)
	localesClient := client.WithSpaceId(spaceID).WithEnvironment(d.Get("environment").(string)).Locales()

	locale, err := getLocale(d, localesClient)

	if err != nil {
		return diag.FromErr(err)
	}

	err = localesClient.Delete(context.Background(), locale)

	var notFoundError *common.NotFoundError
	if errors.As(err, &notFoundError) {
		return nil
	}

	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func setLocaleProperties(d *schema.ResourceData, locale *model.Locale) diag.Diagnostics {
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

	err = d.Set("cda", locale.CDA)
	if err != nil {
		return diag.FromErr(err)
	}

	err = d.Set("cma", locale.CMA)
	if err != nil {
		return diag.FromErr(err)
	}

	err = d.Set("version", locale.Sys.Version)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func getLocale(d *schema.ResourceData, client cma.Locales) (*model.Locale, error) {
	return client.Get(context.Background(), d.Id())
}

package contentful

import (
	"context"
	"errors"
	"fmt"
	"github.com/flaconi/contentful-go"
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
				Optional: true,
			},
		},
	}
}

func resourceCreateLocale(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*contentful.Client)
	spaceID := d.Get("space_id").(string)
	var err error

	locale := &contentful.Locale{
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

	if environment, ok := d.GetOk("environment"); ok {

		env, envErr := client.Environments.Get(spaceID, environment.(string))

		if envErr != nil {
			return diag.FromErr(envErr)
		}

		err = client.Locales.UpsertWithEnv(env, locale)
	} else {
		err = client.Locales.Upsert(spaceID, locale)
	}

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
	client := m.(*contentful.Client)

	locale, err := getLocale(d, client)

	var notFoundError *contentful.NotFoundError
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
	client := m.(*contentful.Client)
	spaceID := d.Get("space_id").(string)

	locale, err := getLocale(d, client)

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

	if environment, ok := d.GetOk("environment"); ok {
		env, envErr := client.Environments.Get(spaceID, environment.(string))

		if envErr != nil {
			return diag.FromErr(envErr)
		}

		err = client.Locales.UpsertWithEnv(env, locale)

	} else {
		err = client.Locales.Upsert(spaceID, locale)
	}

	if err != nil {
		return diag.FromErr(err)
	}

	return setLocaleProperties(d, locale)
}

func resourceDeleteLocale(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*contentful.Client)
	spaceID := d.Get("space_id").(string)

	locale, err := getLocale(d, client)

	if err != nil {
		return diag.FromErr(err)
	}

	if environment, ok := d.GetOk("environment"); ok {
		env, envErr := client.Environments.Get(spaceID, environment.(string))

		if envErr != nil {
			return diag.FromErr(envErr)
		}

		err = client.Locales.DeleteWithEnv(env, locale)

	} else {
		err = client.Locales.Delete(spaceID, locale)
	}

	var notFoundError *contentful.NotFoundError
	if errors.As(err, &notFoundError) {
		return nil
	}

	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func setLocaleProperties(d *schema.ResourceData, locale *contentful.Locale) diag.Diagnostics {
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

func getLocale(d *schema.ResourceData, client *contentful.Client) (*contentful.Locale, error) {
	spaceID := d.Get("space_id").(string)
	if environment, ok := d.GetOk("environment"); ok {
		env, envErr := client.Environments.Get(spaceID, environment.(string))

		if envErr != nil {
			return nil, envErr
		}

		return client.Locales.GetWithEnv(env, d.Id())
	} else {
		return client.Locales.Get(spaceID, d.Id())
	}
}

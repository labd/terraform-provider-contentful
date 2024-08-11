package contentful

import (
	"context"
	"errors"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/labd/contentful-go"
)

func resourceContentfulLocale() *schema.Resource {
	return &schema.Resource{
		Description: "A Contentful Locale represents a language and region combination.",

		CreateContext: resourceCreateLocale,
		ReadContext:   resourceReadLocale,
		UpdateContext: resourceUpdateLocale,
		DeleteContext: resourceDeleteLocale,

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
				Default:  "en-US",
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
		},
	}
}

func resourceCreateLocale(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	time.Sleep(consistencyBuffer)
	client := m.(*contentful.Client)
	spaceID := d.Get("space_id").(string)

	locale := &contentful.Locale{
		Name:         d.Get("name").(string),
		Code:         d.Get("code").(string),
		FallbackCode: d.Get("fallback_code").(string),
		Optional:     d.Get("optional").(bool),
		CDA:          d.Get("cda").(bool),
		CMA:          d.Get("cma").(bool),
	}

	err := client.Locales.Upsert(spaceID, locale)
	if err != nil {
		return parseError(err)
	}

	err = setLocaleProperties(d, locale)
	if err != nil {
		return parseError(err)
	}

	d.SetId(locale.Sys.ID)

	return nil
}

func resourceReadLocale(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	time.Sleep(consistencyBuffer)
	client := m.(*contentful.Client)
	spaceID := d.Get("space_id").(string)
	localeID := d.Id()

	locale, err := client.Locales.Get(spaceID, localeID)
	var notFoundError *contentful.NotFoundError
	if errors.As(err, &notFoundError) {
		d.SetId("")
		return nil
	}

	if err != nil {
		return parseError(err)
	}

	err = setLocaleProperties(d, locale)
	if err != nil {
		return parseError(err)
	}

	return nil
}

func resourceUpdateLocale(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	time.Sleep(consistencyBuffer)
	client := m.(*contentful.Client)
	spaceID := d.Get("space_id").(string)
	localeID := d.Id()

	locale, err := client.Locales.Get(spaceID, localeID)
	if err != nil {
		return parseError(err)
	}

	locale.Name = d.Get("name").(string)
	locale.Code = d.Get("code").(string)
	locale.FallbackCode = d.Get("fallback_code").(string)
	locale.Optional = d.Get("optional").(bool)
	locale.CDA = d.Get("cda").(bool)
	locale.CMA = d.Get("cma").(bool)

	err = client.Locales.Upsert(spaceID, locale)
	if err != nil {
		return parseError(err)
	}

	err = setLocaleProperties(d, locale)
	if err != nil {
		return parseError(err)
	}

	return nil
}

func resourceDeleteLocale(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	time.Sleep(consistencyBuffer)
	client := m.(*contentful.Client)
	spaceID := d.Get("space_id").(string)
	localeID := d.Id()

	locale, err := client.Locales.Get(spaceID, localeID)
	if err != nil {
		return parseError(err)
	}

	err = client.Locales.Delete(spaceID, locale)
	var notFoundError *contentful.NotFoundError
	if errors.As(err, &notFoundError) {
		return nil
	}

	if err != nil {
		return parseError(err)
	}

	return nil
}

func setLocaleProperties(d *schema.ResourceData, locale *contentful.Locale) error {
	err := d.Set("name", locale.Name)
	if err != nil {
		return err
	}

	err = d.Set("code", locale.Code)
	if err != nil {
		return err
	}

	err = d.Set("fallback_code", locale.FallbackCode)
	if err != nil {
		return err
	}

	err = d.Set("optional", locale.Optional)
	if err != nil {
		return err
	}

	err = d.Set("cda", locale.CDA)
	if err != nil {
		return err
	}

	err = d.Set("cma", locale.CMA)
	if err != nil {
		return err
	}

	return nil
}

package contentful

import (
	"context"
	"errors"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/labd/contentful-go"
)

func resourceContentfulEnvironment() *schema.Resource {
	return &schema.Resource{
		Description:   "A Contentful Environment represents a space environment.",
		CreateContext: resourceCreateEnvironment,
		ReadContext:   resourceReadEnvironment,
		UpdateContext: resourceUpdateEnvironment,
		DeleteContext: resourceDeleteEnvironment,

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

func resourceCreateEnvironment(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	time.Sleep(consistencyBuffer)
	client := m.(*contentful.Client)

	environment := &contentful.Environment{
		Name: d.Get("name").(string),
	}

	err := client.Environments.Upsert(d.Get("space_id").(string), environment)
	if err != nil {
		return parseError(err)
	}

	if err := setEnvironmentProperties(d, environment); err != nil {
		return parseError(err)
	}

	d.SetId(environment.Name)

	return nil
}

func resourceUpdateEnvironment(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	time.Sleep(consistencyBuffer)
	client := m.(*contentful.Client)
	spaceID := d.Get("space_id").(string)
	environmentID := d.Id()

	environment, err := client.Environments.Get(spaceID, environmentID)
	if err != nil {
		return parseError(err)
	}

	environment.Name = d.Get("name").(string)

	err = client.Environments.Upsert(spaceID, environment)
	if err != nil {
		return parseError(err)
	}

	if err := setEnvironmentProperties(d, environment); err != nil {
		return parseError(err)
	}

	d.SetId(environment.Sys.ID)

	return nil
}

func resourceReadEnvironment(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	time.Sleep(consistencyBuffer)
	client := m.(*contentful.Client)
	spaceID := d.Get("space_id").(string)
	environmentID := d.Id()

	environment, err := client.Environments.Get(spaceID, environmentID)
	var notFoundError contentful.NotFoundError
	if errors.As(err, &notFoundError) {
		d.SetId("")
		return nil
	}

	err = setEnvironmentProperties(d, environment)
	if err != nil {
		return parseError(err)
	}

	return nil
}

func resourceDeleteEnvironment(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	time.Sleep(consistencyBuffer)
	client := m.(*contentful.Client)
	spaceID := d.Get("space_id").(string)
	environmentID := d.Id()

	environment, err := client.Environments.Get(spaceID, environmentID)
	if err != nil {
		return parseError(err)
	}

	err = client.Environments.Delete(spaceID, environment)
	if err != nil {
		return parseError(err)
	}

	return nil
}

func setEnvironmentProperties(d *schema.ResourceData, environment *contentful.Environment) error {
	if err := d.Set("space_id", environment.Sys.Space.Sys.ID); err != nil {
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

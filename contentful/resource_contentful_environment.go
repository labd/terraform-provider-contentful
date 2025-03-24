package contentful

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/labd/contentful-go/pkgs/common"
	"github.com/labd/contentful-go/pkgs/model"

	"github.com/labd/terraform-provider-contentful/internal/utils"
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
	client := m.(utils.ProviderData).CMAClient

	environment := &model.Environment{
		Name: d.Get("name").(string),
	}

	err := client.WithSpaceId(spaceID).Environments().Upsert(context.Background(), environment, nil)
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
	client := m.(utils.ProviderData).CMAClient
	spaceID := d.Get("space_id").(string)
	environmentID := d.Id()

	envClient := client.WithSpaceId(spaceID).Environments()

	environment, err := envClient.Get(context.Background(), environmentID)
	if err != nil {
		return parseError(err)
	}

	environment.Name = d.Get("name").(string)

	err = envClient.Upsert(context.Background(), environment, nil)
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
	client := m.(utils.ProviderData).CMAClient
	spaceID := d.Get("space_id").(string)
	environmentID := d.Id()

	environment, err := client.WithSpaceId(spaceID).Environments().Get(context.Background(), environmentID)
	if _, ok := err.(common.NotFoundError); ok {
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
	client := m.(utils.ProviderData).CMAClient
	spaceID := d.Get("space_id").(string)
	environmentID := d.Id()

	envClient := client.WithSpaceId(spaceID).Environments()

	environment, err := envClient.Get(context.Background(), environmentID)
	if err != nil {
		return parseError(err)
	}

	err = envClient.Delete(context.Background(), environment)
	if err != nil {
		return parseError(err)
	}

	return nil
}

func setEnvironmentProperties(d *schema.ResourceData, environment *model.Environment) error {
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

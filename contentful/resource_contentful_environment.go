package contentful

import (
	"context"
	"github.com/flaconi/contentful-go/pkgs/common"
	"github.com/flaconi/contentful-go/pkgs/model"
	"github.com/flaconi/terraform-provider-contentful/internal/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceContentfulEnvironment() *schema.Resource {
	return &schema.Resource{
		Create: resourceCreateEnvironment,
		Read:   resourceReadEnvironment,
		Update: resourceUpdateEnvironment,
		Delete: resourceDeleteEnvironment,

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

func resourceCreateEnvironment(d *schema.ResourceData, m interface{}) (err error) {
	client := m.(utils.ProviderData).CMAClient

	environment := &model.Environment{
		Name: d.Get("name").(string),
	}

	err = client.WithSpaceId(spaceID).Environments().Upsert(context.Background(), environment, nil)
	if err != nil {
		return err
	}

	if err := setEnvironmentProperties(d, environment); err != nil {
		return err
	}

	d.SetId(environment.Name)

	return nil
}

func resourceUpdateEnvironment(d *schema.ResourceData, m interface{}) (err error) {
	client := m.(utils.ProviderData).CMAClient
	spaceID := d.Get("space_id").(string)
	environmentID := d.Id()

	envClient := client.WithSpaceId(spaceID).Environments()

	environment, err := envClient.Get(context.Background(), environmentID)
	if err != nil {
		return err
	}

	environment.Name = d.Get("name").(string)

	err = envClient.Upsert(context.Background(), environment, nil)
	if err != nil {
		return err
	}

	if err := setEnvironmentProperties(d, environment); err != nil {
		return err
	}

	d.SetId(environment.Sys.ID)

	return nil
}

func resourceReadEnvironment(d *schema.ResourceData, m interface{}) (err error) {
	client := m.(utils.ProviderData).CMAClient
	spaceID := d.Get("space_id").(string)
	environmentID := d.Id()

	environment, err := client.WithSpaceId(spaceID).Environments().Get(context.Background(), environmentID)
	if _, ok := err.(common.NotFoundError); ok {
		d.SetId("")
		return nil
	}

	return setEnvironmentProperties(d, environment)
}

func resourceDeleteEnvironment(d *schema.ResourceData, m interface{}) (err error) {
	client := m.(utils.ProviderData).CMAClient
	spaceID := d.Get("space_id").(string)
	environmentID := d.Id()

	envClient := client.WithSpaceId(spaceID).Environments()

	environment, err := envClient.Get(context.Background(), environmentID)
	if err != nil {
		return err
	}

	return envClient.Delete(context.Background(), environment)
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

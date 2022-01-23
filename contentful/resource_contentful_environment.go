package contentful

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	contentful "github.com/labd/contentful-go"
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
	client := m.(*contentful.Client)

	environment := &contentful.Environment{
		Name: d.Get("name").(string),
	}

	err = client.Environments.Upsert(d.Get("space_id").(string), environment)
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
	client := m.(*contentful.Client)
	spaceID := d.Get("space_id").(string)
	environmentID := d.Id()

	environment, err := client.Environments.Get(spaceID, environmentID)
	if err != nil {
		return err
	}

	environment.Name = d.Get("name").(string)

	err = client.Environments.Upsert(spaceID, environment)
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
	client := m.(*contentful.Client)
	spaceID := d.Get("space_id").(string)
	environmentID := d.Id()

	environment, err := client.Environments.Get(spaceID, environmentID)
	if _, ok := err.(contentful.NotFoundError); ok {
		d.SetId("")
		return nil
	}

	return setEnvironmentProperties(d, environment)
}

func resourceDeleteEnvironment(d *schema.ResourceData, m interface{}) (err error) {
	client := m.(*contentful.Client)
	spaceID := d.Get("space_id").(string)
	environmentID := d.Id()

	environment, err := client.Environments.Get(spaceID, environmentID)
	if err != nil {
		return err
	}

	return client.Environments.Delete(spaceID, environment)
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

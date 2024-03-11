package contentful

import (
	contentful "github.com/flaconi/contentful-go"
	"github.com/flaconi/contentful-go/pkgs/common"
	"github.com/flaconi/terraform-provider-contentful/internal/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceContentfulSpace() *schema.Resource {
	return &schema.Resource{
		Create: resourceSpaceCreate,
		Read:   resourceSpaceRead,
		Update: resourceSpaceUpdate,
		Delete: resourceSpaceDelete,

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

func resourceSpaceCreate(d *schema.ResourceData, m interface{}) (err error) {
	client := m.(utils.ProviderData).Client

	space := &contentful.Space{
		Name:          d.Get("name").(string),
		DefaultLocale: d.Get("default_locale").(string),
	}

	err = client.Spaces.Upsert(space)
	if err != nil {
		return err
	}

	err = updateSpaceProperties(d, space)
	if err != nil {
		return err
	}

	d.SetId(space.Sys.ID)

	return nil
}

func resourceSpaceRead(d *schema.ResourceData, m interface{}) error {
	client := m.(utils.ProviderData).Client
	spaceID := d.Id()

	_, err := client.Spaces.Get(spaceID)
	if _, ok := err.(common.NotFoundError); ok {
		d.SetId("")
		return nil
	}

	return err
}

func resourceSpaceUpdate(d *schema.ResourceData, m interface{}) (err error) {
	client := m.(utils.ProviderData).Client
	spaceID := d.Id()

	space, err := client.Spaces.Get(spaceID)
	if err != nil {
		return err
	}

	space.Name = d.Get("name").(string)

	err = client.Spaces.Upsert(space)
	if err != nil {
		return err
	}

	return updateSpaceProperties(d, space)
}

func resourceSpaceDelete(d *schema.ResourceData, m interface{}) (err error) {
	client := m.(utils.ProviderData).Client
	spaceID := d.Id()

	space, err := client.Spaces.Get(spaceID)
	if err != nil {
		return err
	}

	err = client.Spaces.Delete(space)
	if _, ok := err.(common.NotFoundError); ok {
		return nil
	}

	return err
}

func updateSpaceProperties(d *schema.ResourceData, space *contentful.Space) error {
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

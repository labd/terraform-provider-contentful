package contentful

import (
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	contentful "github.com/labd/contentful-go"
)

func resourceContentfulUpload() *schema.Resource {
	return &schema.Resource{
		Create: resourceCreateUpload,
		Read:   resourceReadUpload,
		Update: nil,
		Delete: resourceDeleteUpload,

		Schema: map[string]*schema.Schema{
			"space_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"file_path": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"asset_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"locale": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"title": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"description": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceCreateUpload(d *schema.ResourceData, m interface{}) (err error) {
	uploadClient := contentful.NewResourceClient(CMAToken)
	uploadClient.SetOrganization(orgID)
	client := m.(*contentful.Client)

	upload := contentful.Resource{}

	fmt.Println(d.Get("space_id").(string))
	fmt.Println(d.Get("file_path").(string))

	response := uploadClient.Resources.Create(d.Get("space_id").(string), d.Get("file_path").(string))
	err = json.Unmarshal([]byte(response.Error()), &upload)

	d.SetId(upload.Sys.ID)

	if err := setUploadProperties(d, &upload); err != nil {
		return err
	}

	asset := &contentful.Asset{
		Sys: &contentful.Sys{
			ID:      d.Get("asset_id").(string),
			Version: 0,
		},
		Locale: d.Get("locale").(string),
		Fields: &contentful.AssetFields{
			Title: map[string]string{
				d.Get("locale").(string): d.Get("title").(string),
			},
			Description: map[string]string{
				d.Get("locale").(string): d.Get("description").(string),
			},
			File: map[string]*contentful.File{
				d.Get("locale").(string): {
					UploadFrom: &contentful.UploadFrom{
						Sys: &contentful.Sys{
							ID:       "upload.Sys.ID",
							LinkType: "Upload",
						},
					},
				},
			},
		},
	}

	err = client.Assets.Upsert(d.Get("space_id").(string), asset)
	if err != nil {
		return err
	}

	err = client.Assets.Process(d.Get("space_id").(string), asset)
	if err != nil {
		return err
	}

	err = client.Assets.Publish(spaceID, asset)
	if err != nil {
		return err
	}

	return err
}

func resourceReadUpload(d *schema.ResourceData, m interface{}) (err error) {
	client := m.(*contentful.Client)
	spaceID := d.Get("space_id").(string)
	uploadID := d.Id()

	upload, err := client.Resources.Get(spaceID, uploadID)
	if _, ok := err.(contentful.NotFoundError); ok {
		d.SetId("")
		return err
	}

	return setUploadProperties(d, upload)
}

func resourceDeleteUpload(d *schema.ResourceData, m interface{}) (err error) {
	client := m.(*contentful.Client)
	spaceID := d.Get("space_id").(string)
	uploadID := d.Id()

	_, err = client.Resources.Get(spaceID, uploadID)
	if err != nil {
		return err
	}

	return client.Resources.Delete(spaceID, uploadID)
}

func setUploadProperties(d *schema.ResourceData, resource *contentful.Resource) (err error) {
	if err = d.Set("space_id", resource.Sys.Space.Sys.ID); err != nil {
		return err
	}

	return err
}

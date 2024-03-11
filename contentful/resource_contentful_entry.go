package contentful

import (
	"context"
	"github.com/flaconi/contentful-go/pkgs/common"
	"github.com/flaconi/contentful-go/pkgs/model"
	"github.com/flaconi/terraform-provider-contentful/internal/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceContentfulEntry() *schema.Resource {
	return &schema.Resource{
		Create: resourceCreateEntry,
		Read:   resourceReadEntry,
		Update: resourceUpdateEntry,
		Delete: resourceDeleteEntry,

		Schema: map[string]*schema.Schema{
			"entry_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"version": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"space_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"environment": {
				Type:     schema.TypeString,
				Required: true,
			},
			"contenttype_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"field": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"content": {
							Type:     schema.TypeString,
							Required: true,
						},
						"locale": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"published": {
				Type:     schema.TypeBool,
				Required: true,
			},
			"archived": {
				Type:     schema.TypeBool,
				Required: true,
			},
		},
	}
}

func resourceCreateEntry(d *schema.ResourceData, m interface{}) (err error) {
	client := m.(utils.ProviderData).CMAClient

	fieldProperties := map[string]interface{}{}
	rawField := d.Get("field").([]interface{})
	for i := 0; i < len(rawField); i++ {
		field := rawField[i].(map[string]interface{})
		fieldProperties[field["id"].(string)] = map[string]interface{}{}
		fieldProperties[field["id"].(string)].(map[string]interface{})[field["locale"].(string)] = field["content"].(string)
	}

	entry := &model.Entry{
		Fields: fieldProperties,
		Sys: &model.PublishSys{
			EnvironmentSys: model.EnvironmentSys{
				SpaceSys: model.SpaceSys{
					CreatedSys: model.CreatedSys{
						BaseSys: model.BaseSys{
							ID: d.Get("entry_id").(string),
						},
					},
				},
			},
		},
	}

	err = client.WithSpaceId(d.Get("space_id").(string)).WithEnvironment(d.Get("environment").(string)).Entries().Upsert(context.Background(), d.Get("contenttype_id").(string), entry)
	if err != nil {
		return err
	}

	if err := setEntryProperties(d, entry); err != nil {
		return err
	}

	d.SetId(entry.Sys.ID)

	if err := setEntryState(d, m); err != nil {
		return err
	}

	return err
}

func resourceUpdateEntry(d *schema.ResourceData, m interface{}) (err error) {
	client := m.(utils.ProviderData).CMAClient
	spaceID := d.Get("space_id").(string)
	entryID := d.Id()

	entryClient := client.WithSpaceId(spaceID).WithEnvironment(d.Get("environment").(string)).Entries()

	entry, err := entryClient.Get(context.Background(), entryID)
	if err != nil {
		return err
	}

	fieldProperties := map[string]interface{}{}
	rawField := d.Get("field").([]interface{})
	for i := 0; i < len(rawField); i++ {
		field := rawField[i].(map[string]interface{})
		fieldProperties[field["id"].(string)] = map[string]interface{}{}
		fieldProperties[field["id"].(string)].(map[string]interface{})[field["locale"].(string)] = field["content"].(string)
	}

	entry.Fields = fieldProperties

	err = entryClient.Upsert(context.Background(), d.Get("contenttype_id").(string), entry)
	if err != nil {
		return err
	}

	d.SetId(entry.Sys.ID)

	if err := setEntryProperties(d, entry); err != nil {
		return err
	}

	if err := setEntryState(d, m); err != nil {
		return err
	}

	return err
}

func setEntryState(d *schema.ResourceData, m interface{}) (err error) {
	client := m.(utils.ProviderData).CMAClient
	spaceID := d.Get("space_id").(string)
	entryID := d.Id()

	entryClient := client.WithSpaceId(spaceID).WithEnvironment(d.Get("environment").(string)).Entries()

	entry, _ := entryClient.Get(context.Background(), entryID)

	if d.Get("published").(bool) && entry.Sys.PublishedAt == "" {
		err = entryClient.Publish(context.Background(), entry)
	} else if !d.Get("published").(bool) && entry.Sys.PublishedAt != "" {
		err = entryClient.Unpublish(context.Background(), entry)
	}

	if d.Get("archived").(bool) && entry.Sys.ArchivedAt == "" {
		err = entryClient.Archive(context.Background(), entry)
	} else if !d.Get("archived").(bool) && entry.Sys.ArchivedAt != "" {
		err = entryClient.Unarchive(context.Background(), entry)
	}

	return err
}

func resourceReadEntry(d *schema.ResourceData, m interface{}) (err error) {
	client := m.(utils.ProviderData).CMAClient
	spaceID := d.Get("space_id").(string)
	entryID := d.Id()

	entry, err := client.WithSpaceId(spaceID).WithEnvironment(d.Get("environment").(string)).Entries().Get(context.Background(), entryID)
	if _, ok := err.(common.NotFoundError); ok {
		d.SetId("")
		return nil
	}

	return setEntryProperties(d, entry)
}

func resourceDeleteEntry(d *schema.ResourceData, m interface{}) (err error) {
	client := m.(utils.ProviderData).CMAClient
	spaceID := d.Get("space_id").(string)
	entryID := d.Id()

	entryClient := client.WithSpaceId(spaceID).WithEnvironment(d.Get("environment").(string)).Entries()

	entry, err := entryClient.Get(context.Background(), entryID)
	if err != nil {
		return err
	}

	return entryClient.Delete(context.Background(), entry)
}

func setEntryProperties(d *schema.ResourceData, entry *model.Entry) (err error) {
	if err = d.Set("space_id", entry.Sys.Space.Sys.ID); err != nil {
		return err
	}

	if err = d.Set("version", entry.Sys.Version); err != nil {
		return err
	}

	if err = d.Set("contenttype_id", entry.Sys.ContentType.Sys.ID); err != nil {
		return err
	}

	return err
}

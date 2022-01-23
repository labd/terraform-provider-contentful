package contentful

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	contentful "github.com/labd/contentful-go"
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
			"contenttype_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"locale": {
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
	client := m.(*contentful.Client)

	fieldProperties := map[string]interface{}{}
	rawField := d.Get("field").([]interface{})
	for i := 0; i < len(rawField); i++ {
		field := rawField[i].(map[string]interface{})
		fieldProperties[field["id"].(string)] = map[string]interface{}{}
		fieldProperties[field["id"].(string)].(map[string]interface{})[field["locale"].(string)] = field["content"].(string)
	}

	entry := &contentful.Entry{
		Locale: d.Get("locale").(string),
		Fields: fieldProperties,
		Sys: &contentful.Sys{
			ID: d.Get("entry_id").(string),
		},
	}

	err = client.Entries.Upsert(d.Get("space_id").(string), d.Get("contenttype_id").(string), entry)
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
	client := m.(*contentful.Client)
	spaceID := d.Get("space_id").(string)
	entryID := d.Id()

	entry, err := client.Entries.Get(spaceID, entryID)
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
	entry.Locale = d.Get("locale").(string)

	err = client.Entries.Upsert(d.Get("space_id").(string), d.Get("contenttype_id").(string), entry)
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
	client := m.(*contentful.Client)
	spaceID := d.Get("space_id").(string)
	entryID := d.Id()

	entry, _ := client.Entries.Get(spaceID, entryID)

	if d.Get("published").(bool) && entry.Sys.PublishedAt == "" {
		err = client.Entries.Publish(spaceID, entry)
	} else if !d.Get("published").(bool) && entry.Sys.PublishedAt != "" {
		err = client.Entries.Unpublish(spaceID, entry)
	}

	if d.Get("archived").(bool) && entry.Sys.ArchivedAt == "" {
		err = client.Entries.Archive(spaceID, entry)
	} else if !d.Get("archived").(bool) && entry.Sys.ArchivedAt != "" {
		err = client.Entries.Unarchive(spaceID, entry)
	}

	return err
}

func resourceReadEntry(d *schema.ResourceData, m interface{}) (err error) {
	client := m.(*contentful.Client)
	spaceID := d.Get("space_id").(string)
	entryID := d.Id()

	entry, err := client.Entries.Get(spaceID, entryID)
	if _, ok := err.(contentful.NotFoundError); ok {
		d.SetId("")
		return nil
	}

	return setEntryProperties(d, entry)
}

func resourceDeleteEntry(d *schema.ResourceData, m interface{}) (err error) {
	client := m.(*contentful.Client)
	spaceID := d.Get("space_id").(string)
	entryID := d.Id()

	_, err = client.Entries.Get(spaceID, entryID)
	if err != nil {
		return err
	}

	return client.Entries.Delete(spaceID, entryID)
}

func setEntryProperties(d *schema.ResourceData, entry *contentful.Entry) (err error) {
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

package contentful

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/labd/contentful-go"
)

func resourceContentfulEntry() *schema.Resource {
	return &schema.Resource{
		Description: "A Contentful Entry represents a piece of content in a space.",

		CreateContext: resourceCreateEntry,
		ReadContext:   resourceReadEntry,
		UpdateContext: resourceUpdateEntry,
		DeleteContext: resourceDeleteEntry,

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
							Type:        schema.TypeString,
							Required:    true,
							Description: "The content of the field. If the field type is Richtext the content can be passed as stringified JSON (see example).",
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

func resourceCreateEntry(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	time.Sleep(consistencyBuffer)
	client := m.(*contentful.Client)

	fieldProperties := map[string]interface{}{}
	rawField := d.Get("field").([]interface{})
	for i := 0; i < len(rawField); i++ {
		field := rawField[i].(map[string]interface{})
		fieldProperties[field["id"].(string)] = map[string]interface{}{}
		fieldProperties[field["id"].(string)].(map[string]interface{})[field["locale"].(string)] = parseContentValue(field["content"].(string))
	}

	entry := &contentful.Entry{
		Locale: d.Get("locale").(string),
		Fields: fieldProperties,
		Sys: &contentful.Sys{
			ID: d.Get("entry_id").(string),
		},
	}

	err := client.Entries.Upsert(d.Get("space_id").(string), d.Get("contenttype_id").(string), entry)
	if err != nil {
		return parseError(err)
	}

	if err := setEntryProperties(d, entry); err != nil {
		return parseError(err)
	}

	d.SetId(entry.Sys.ID)

	if err := setEntryState(d, m); err != nil {
		return parseError(err)
	}

	return parseError(err)
}

func resourceUpdateEntry(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	time.Sleep(consistencyBuffer)
	client := m.(*contentful.Client)
	spaceID := d.Get("space_id").(string)
	entryID := d.Id()

	entry, err := client.Entries.Get(spaceID, entryID)
	if err != nil {
		return parseError(err)
	}

	fieldProperties := map[string]interface{}{}
	rawField := d.Get("field").([]interface{})
	for i := 0; i < len(rawField); i++ {
		field := rawField[i].(map[string]interface{})
		fieldProperties[field["id"].(string)] = map[string]interface{}{}
		fieldProperties[field["id"].(string)].(map[string]interface{})[field["locale"].(string)] = parseContentValue(field["content"].(string))
	}

	entry.Fields = fieldProperties
	entry.Locale = d.Get("locale").(string)

	err = client.Entries.Upsert(d.Get("space_id").(string), d.Get("contenttype_id").(string), entry)
	if err != nil {
		return parseError(err)
	}

	d.SetId(entry.Sys.ID)

	if err := setEntryProperties(d, entry); err != nil {
		return parseError(err)
	}

	if err := setEntryState(d, m); err != nil {
		return parseError(err)
	}

	return nil
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

func resourceReadEntry(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	time.Sleep(consistencyBuffer)
	client := m.(*contentful.Client)
	spaceID := d.Get("space_id").(string)
	entryID := d.Id()

	entry, err := client.Entries.Get(spaceID, entryID)
	var notFoundError contentful.NotFoundError
	if errors.As(err, &notFoundError) {
		d.SetId("")
		return nil
	}

	err = setEntryProperties(d, entry)
	if err != nil {
		return parseError(err)
	}

	return nil
}

func resourceDeleteEntry(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	time.Sleep(consistencyBuffer)
	client := m.(*contentful.Client)
	spaceID := d.Get("space_id").(string)
	entryID := d.Id()

	_, err := client.Entries.Get(spaceID, entryID)
	if err != nil {
		return parseError(err)
	}

	err = client.Entries.Delete(spaceID, entryID)
	if err != nil {
		return parseError(err)
	}

	return nil
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

func parseContentValue(value interface{}) interface{} {
	var content interface{}
	err := json.Unmarshal([]byte(value.(string)), &content)
	if err != nil {
		content = value
	}

	return content
}

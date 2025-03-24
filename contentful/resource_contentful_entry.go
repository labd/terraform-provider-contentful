package contentful

import (
	"context"
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/labd/contentful-go/pkgs/common"
	"github.com/labd/contentful-go/pkgs/model"

	"github.com/labd/terraform-provider-contentful/internal/utils"
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
	client := m.(utils.ProviderData).CMAClient

	fieldProperties := map[string]interface{}{}
	rawField := d.Get("field").([]interface{})
	for i := 0; i < len(rawField); i++ {
		field := rawField[i].(map[string]interface{})
		fieldProperties[field["id"].(string)] = map[string]interface{}{}
		fieldProperties[field["id"].(string)].(map[string]interface{})[field["locale"].(string)] = parseContentValue(field["content"].(string))
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

	err := client.WithSpaceId(d.Get("space_id").(string)).WithEnvironment(d.Get("environment").(string)).Entries().Upsert(context.Background(), d.Get("contenttype_id").(string), entry)
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
	client := m.(utils.ProviderData).CMAClient
	spaceID := d.Get("space_id").(string)
	entryID := d.Id()

	entryClient := client.WithSpaceId(spaceID).WithEnvironment(d.Get("environment").(string)).Entries()

	entry, err := entryClient.Get(context.Background(), entryID)
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

	err = entryClient.Upsert(context.Background(), d.Get("contenttype_id").(string), entry)
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

func resourceReadEntry(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(utils.ProviderData).CMAClient
	spaceID := d.Get("space_id").(string)
	entryID := d.Id()

	entry, err := client.WithSpaceId(spaceID).WithEnvironment(d.Get("environment").(string)).Entries().Get(context.Background(), entryID)
	if _, ok := err.(common.NotFoundError); ok {
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
	client := m.(utils.ProviderData).CMAClient
	spaceID := d.Get("space_id").(string)
	entryID := d.Id()

	entryClient := client.WithSpaceId(spaceID).WithEnvironment(d.Get("environment").(string)).Entries()

	entry, err := entryClient.Get(context.Background(), entryID)
	if err != nil {
		return parseError(err)
	}

	err = entryClient.Delete(context.Background(), entry)
	if err != nil {
		return parseError(err)
	}

	return nil
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

func parseContentValue(value interface{}) interface{} {
	var content interface{}
	err := json.Unmarshal([]byte(value.(string)), &content)
	if err != nil {
		content = value
	}

	return content
}

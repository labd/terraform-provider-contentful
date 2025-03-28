package contentful

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/labd/terraform-provider-contentful/internal/sdk"
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

func resourceCreateEntry(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(utils.ProviderData).Client
	spaceID := d.Get("space_id").(string)
	environmentID := d.Get("environment").(string)

	fieldProperties := map[string]interface{}{}
	rawField := d.Get("field").([]interface{})
	for i := 0; i < len(rawField); i++ {
		field := rawField[i].(map[string]interface{})
		fieldProperties[field["id"].(string)] = map[string]interface{}{}
		fieldProperties[field["id"].(string)].(map[string]interface{})[field["locale"].(string)] = parseContentValue(field["content"].(string))
	}

	body := sdk.EntryCreate{
		Fields: &fieldProperties,
	}

	resp, err := client.CreateEntryWithResponse(ctx, spaceID, environmentID, nil, body)
	if err != nil {
		return parseError(err)
	}

	if resp.StatusCode() != 201 {
		return diag.Errorf("Failed to create entry")
	}

	entry := resp.JSON201
	if err := setEntryProperties(d, entry); err != nil {
		return parseError(err)
	}

	d.SetId(*entry.Sys.Id)

	if err := setEntryState(ctx, d, m); err != nil {
		return parseError(err)
	}

	return parseError(err)
}

func resourceUpdateEntry(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(utils.ProviderData).Client
	spaceID := d.Get("space_id").(string)
	environmentID := d.Get("environment").(string)
	entryID := d.Id()

	fieldProperties := map[string]any{}
	rawField := d.Get("field").([]any)
	for i := 0; i < len(rawField); i++ {
		field := rawField[i].(map[string]any)
		fieldProperties[field["id"].(string)] = map[string]any{}
		fieldProperties[field["id"].(string)].(map[string]any)[field["locale"].(string)] = parseContentValue(field["content"].(string))
	}

	body := sdk.EntryUpdate{
		Fields: &fieldProperties,
	}

	params := &sdk.UpdateEntryParams{
		XContentfulVersion: int64(d.Get("version").(int)),
	}

	resp, err := client.UpdateEntryWithResponse(ctx, spaceID, environmentID, entryID, params, body)
	if err != nil {
		return parseError(err)
	}

	if resp.StatusCode() != 200 {
		return diag.Errorf("Failed to update entry")
	}

	entry := resp.JSON200
	d.SetId(*entry.Sys.Id)

	if err := setEntryProperties(d, entry); err != nil {
		return parseError(err)
	}

	if err := setEntryState(ctx, d, m); err != nil {
		return parseError(err)
	}

	return nil
}

func setEntryState(ctx context.Context, d *schema.ResourceData, m interface{}) (err error) {
	client := m.(utils.ProviderData).Client
	spaceID := d.Get("space_id").(string)
	entryID := d.Id()
	environmentID := d.Get("environment").(string)

	published := d.Get("published").(bool)
	archived := d.Get("archived").(bool)

	resp, err := client.GetEntryWithResponse(ctx, spaceID, environmentID, entryID)
	if err != nil {
		return err
	}

	if resp.StatusCode() != 200 {
		return fmt.Errorf("Failed to get entry")
	}

	entry := resp.JSON200

	if published && entry.Sys.PublishedAt == nil {
		resp, err := client.PublishEntryWithResponse(ctx, spaceID, environmentID, entryID)
		if err != nil {
			return err
		}
		if resp.StatusCode() != 200 {
			return fmt.Errorf("Failed to publish entry")
		}
	} else if !published && entry.Sys.PublishedAt != nil {
		resp, err := client.UnpublishEntryWithResponse(ctx, spaceID, environmentID, entryID)
		if err != nil {
			return err
		}
		if resp.StatusCode() != 200 {
			return fmt.Errorf("Failed to unpublish entry")
		}
	}

	if archived && entry.Sys.ArchivedAt == nil {
		resp, err := client.ArchiveEntryWithResponse(ctx, spaceID, environmentID, entryID)
		if err != nil {
			return err
		}
		if resp.StatusCode() != 200 {
			return fmt.Errorf("Failed to archive entry")
		}

	} else if !archived && entry.Sys.ArchivedAt != nil {
		resp, err := client.UnarchiveEntryWithResponse(ctx, spaceID, environmentID, entryID)
		if err != nil {
			return err
		}
		if resp.StatusCode() != 200 {
			return fmt.Errorf("Failed to unarchive entry")
		}
	}

	return err
}

func resourceReadEntry(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(utils.ProviderData).Client
	spaceID := d.Get("space_id").(string)
	entryID := d.Id()
	environmentID := d.Get("environment").(string)

	resp, err := client.GetEntryWithResponse(ctx, spaceID, environmentID, entryID)
	if err != nil {
		return parseError(err)
	}

	if resp.StatusCode() != 200 {
		d.SetId("")
		return nil
	}

	entry := resp.JSON200
	err = setEntryProperties(d, entry)
	if err != nil {
		return parseError(err)
	}

	return nil
}
func resourceDeleteEntry(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(utils.ProviderData).Client
	spaceID := d.Get("space_id").(string)
	environmentID := d.Get("environment").(string)
	entryID := d.Id()

	resp, err := client.GetEntryWithResponse(ctx, spaceID, environmentID, entryID)
	if err != nil {
		return parseError(err)
	}

	if resp.StatusCode() != 200 {
		return diag.Errorf("Failed to get entry")
	}

	params := &sdk.DeleteEntryParams{
		XContentfulVersion: *resp.JSON200.Sys.Version,
	}

	respDelete, err := client.DeleteEntryWithResponse(ctx, spaceID, environmentID, entryID, params)
	if err != nil {
		return parseError(err)
	}

	if respDelete.StatusCode() != 204 {
		return diag.Errorf("Failed to delete entry")
	}

	return nil
}

func setEntryProperties(d *schema.ResourceData, entry *sdk.Entry) (err error) {
	if err = d.Set("space_id", entry.Sys.Space.Sys.Id); err != nil {
		return err
	}

	if err = d.Set("version", entry.Sys.Version); err != nil {
		return err
	}

	if err = d.Set("contenttype_id", entry.Sys.ContentType.Sys.Id); err != nil {
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

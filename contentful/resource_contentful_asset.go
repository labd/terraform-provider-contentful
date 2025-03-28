package contentful

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/labd/terraform-provider-contentful/internal/sdk"
	"github.com/labd/terraform-provider-contentful/internal/utils"
)

func resourceContentfulAsset() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceCreateAsset,
		ReadContext:   resourceReadAsset,
		UpdateContext: resourceUpdateAsset,
		DeleteContext: resourceDeleteAsset,

		Schema: map[string]*schema.Schema{
			"asset_id": {
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
			"fields": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"title": {
							Type:     schema.TypeList,
							Required: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
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
						"description": {
							Type:     schema.TypeList,
							Required: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
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
						"file": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"locale": {
										Type:     schema.TypeString,
										Required: true,
									},
									"url": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"upload": {
										Type:     schema.TypeString,
										Required: true,
									},
									"details": {
										Type:     schema.TypeSet,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"size": {
													Type:     schema.TypeInt,
													Required: true,
												},
												"image": {
													Type:     schema.TypeSet,
													Required: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"width": {
																Type:     schema.TypeInt,
																Required: true,
															},
															"height": {
																Type:     schema.TypeInt,
																Required: true,
															},
														},
													},
												},
											},
										},
									},
									"upload_from": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"file_name": {
										Type:     schema.TypeString,
										Required: true,
									},
									"content_type": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
					},
				},
			},
			"published": {
				Type:     schema.TypeBool,
				Required: true,
			},
			"environment": {
				Type:     schema.TypeString,
				Required: true,
			},
			"archived": {
				Type:     schema.TypeBool,
				Required: true,
			},
		},
	}
}

func resourceCreateAsset(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(utils.ProviderData).Client
	spaceID := d.Get("space_id").(string)
	environment := d.Get("environment").(string)

	draft, err := buildAsset(d)

	if err != nil {
		return parseError(err)
	}

	resp, err := client.CreateAssetWithResponse(ctx, spaceID, environment, *draft)
	if err != nil {
		return parseError(err)
	}
	if resp.StatusCode() != 201 {
		return diag.Errorf("Failed to create asset")
	}

	asset := resp.JSON201

	for locale := range asset.Fields.File {
		processResp, err := client.ProcessAssetWithResponse(ctx, spaceID, environment, *asset.Sys.Id, locale)
		if err != nil {
			return parseError(err)
		}

		if processResp.StatusCode() != 204 {
			return diag.Errorf("Failed to process asset")
		}
	}

	if err := setAssetProperties(d, asset); err != nil {
		return parseError(err)
	}

	time.Sleep(1 * time.Second) // avoid race conditions with version mismatches

	if err = setAssetState(ctx, d, m); err != nil {
		return parseError(err)
	}

	return nil
}

func resourceUpdateAsset(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(utils.ProviderData).Client
	spaceID := d.Get("space_id").(string)
	environmentID := d.Get("environment").(string)
	assetID := d.Id()

	draft, err := buildAsset(d)
	if err != nil {
		return parseError(err)
	}

	resp, err := client.GetAssetWithResponse(ctx, spaceID, environmentID, assetID)
	if err != nil {
		return parseError(err)
	}

	if resp.StatusCode() != 200 {
		return diag.Errorf("Failed to get asset")
	}
	asset := resp.JSON200

	params := &sdk.UpdateAssetParams{
		XContentfulVersion: int64(d.Get("version").(int)),
	}

	respUpdate, err := client.UpdateAssetWithResponse(ctx, spaceID, environmentID, *asset.Sys.Id, params, *draft)
	if err != nil {
		return parseError(err)
	}
	if respUpdate.StatusCode() != 200 {
		return diag.Errorf("Failed to create asset")
	}

	asset = resp.JSON200

	for locale := range asset.Fields.File {
		processResp, err := client.ProcessAssetWithResponse(ctx, spaceID, environmentID, *asset.Sys.Id, locale)
		if err != nil {
			return parseError(err)
		}

		if processResp.StatusCode() != 204 {
			return diag.Errorf("Failed to process asset")
		}
	}

	if err := setAssetProperties(d, asset); err != nil {
		return parseError(err)
	}

	if err = setAssetState(ctx, d, m); err != nil {
		return parseError(err)
	}

	return nil
}

func buildAsset(d *schema.ResourceData) (*sdk.AssetCreate, error) {
	fields := d.Get("fields").([]interface{})[0].(map[string]interface{})

	localizedTitle := map[string]string{}
	rawTitle := fields["title"].([]interface{})
	for i := 0; i < len(rawTitle); i++ {
		field := rawTitle[i].(map[string]interface{})
		localizedTitle[field["locale"].(string)] = field["content"].(string)
	}

	localizedDescription := map[string]string{}
	rawDescription := fields["description"].([]interface{})
	for i := 0; i < len(rawDescription); i++ {
		field := rawDescription[i].(map[string]interface{})
		localizedDescription[field["locale"].(string)] = field["content"].(string)
	}

	files := fields["file"].([]interface{})

	if len(files) == 0 {
		return nil, fmt.Errorf("file block not defined in asset")
	}

	fileData := map[string]sdk.AssetFile{}
	for _, file := range files {
		fileLocale := file.(map[string]any)
		key := fileLocale["locale"].(string)

		fileData[key] = sdk.AssetFile{
			Upload:      fileLocale["upload"].(string),
			FileName:    fileLocale["file_name"].(string),
			ContentType: fileLocale["content_type"].(string),
		}
	}

	return &sdk.AssetCreate{
		Fields: &sdk.AssetField{
			Title:       localizedTitle,
			Description: localizedDescription,
			File:        fileData,
		},
	}, nil
}

func setAssetState(ctx context.Context, d *schema.ResourceData, m interface{}) (err error) {
	client := m.(utils.ProviderData).Client
	spaceID := d.Get("space_id").(string)
	environmentID := d.Get("environment").(string)
	assetID := d.Id()

	resp, err := client.GetAssetWithResponse(ctx, spaceID, environmentID, assetID)
	if err != nil {
		return err
	}

	if resp.StatusCode() != 200 {
		return fmt.Errorf("Failed to get asset")
	}
	asset := resp.JSON200
	setAssetProperties(d, asset)

	if d.Get("published").(bool) && asset.Sys.PublishedAt == nil {
		params := &sdk.PublishAssetParams{
			XContentfulVersion: int64(d.Get("version").(int)),
		}
		resp, err := client.PublishAssetWithResponse(ctx, spaceID, environmentID, assetID, params)
		if err != nil {
			return err
		}
		if resp.StatusCode() != 200 {
			return fmt.Errorf("Failed to publish asset")
		}
		asset = resp.JSON200
	} else if !d.Get("published").(bool) && asset.Sys.PublishedAt != nil {
		params := &sdk.UnpublishAssetParams{
			XContentfulVersion: int64(d.Get("version").(int)),
		}
		resp, err := client.UnpublishAssetWithResponse(ctx, spaceID, environmentID, assetID, params)
		if err != nil {
			return err
		}
		if resp.StatusCode() != 200 {
			return fmt.Errorf("Failed to unpublish asset")
		}
		asset = resp.JSON200
	}
	setAssetProperties(d, asset)

	if d.Get("archived").(bool) && asset.Sys.ArchivedAt == nil {
		params := &sdk.ArchiveAssetParams{
			XContentfulVersion: int64(d.Get("version").(int)),
		}
		resp, err := client.ArchiveAssetWithResponse(ctx, spaceID, environmentID, assetID, params)
		if err != nil {
			return err
		}
		if resp.StatusCode() != 200 {
			return fmt.Errorf("Failed to publish asset")
		}
		asset = resp.JSON200
	} else if !d.Get("archived").(bool) && asset.Sys.ArchivedAt != nil {
		params := &sdk.UnarchiveAssetParams{
			XContentfulVersion: int64(d.Get("version").(int)),
		}
		resp, err := client.UnarchiveAssetWithResponse(ctx, spaceID, environmentID, assetID, params)
		if err != nil {
			return err
		}
		if resp.StatusCode() != 200 {
			return fmt.Errorf("Failed to unpublish asset")
		}
		asset = resp.JSON200
	}

	err = setAssetProperties(d, asset)
	return err
}

func resourceReadAsset(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(utils.ProviderData).Client
	spaceID := d.Get("space_id").(string)
	environmentID := d.Get("environment").(string)
	assetID := d.Id()

	resp, err := client.GetAssetWithResponse(ctx, spaceID, environmentID, assetID)
	if err != nil {
		return parseError(err)
	}

	if resp.StatusCode() == 404 {
		d.SetId("")
		return nil
	}

	if resp.StatusCode() != 200 {
		return parseError(err)
	}

	asset := resp.JSON200

	if err := setAssetProperties(d, asset); err != nil {
		return parseError(err)
	}
	return nil
}

func resourceDeleteAsset(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(utils.ProviderData).Client
	spaceID := d.Get("space_id").(string)
	environmentID := d.Get("environment").(string)
	assetID := d.Id()

	params := &sdk.DeleteAssetParams{
		XContentfulVersion: int64(d.Get("version").(int)),
	}
	resp, err := client.DeleteAssetWithResponse(ctx, spaceID, environmentID, assetID, params)
	if err != nil {
		return parseError(err)
	}

	if resp.StatusCode() != 204 {
		return diag.Errorf("Failed to delete asset")
	}

	return nil
}

func setAssetProperties(d *schema.ResourceData, asset *sdk.Asset) (err error) {
	d.SetId(*asset.Sys.Id)

	if err = d.Set("space_id", asset.Sys.Space.Sys.Id); err != nil {
		return err
	}

	if err = d.Set("version", asset.Sys.Version); err != nil {
		return err
	}

	return err
}

package contentful

import (
	"context"
	"fmt"
	"github.com/flaconi/contentful-go/pkgs/common"
	"github.com/flaconi/contentful-go/pkgs/model"
	"github.com/flaconi/terraform-provider-contentful/internal/utils"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceContentfulAsset() *schema.Resource {
	return &schema.Resource{
		Create: resourceCreateAsset,
		Read:   resourceReadAsset,
		Update: resourceUpdateAsset,
		Delete: resourceDeleteAsset,

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

func resourceCreateAsset(d *schema.ResourceData, m interface{}) (err error) {
	client := m.(utils.ProviderData).CMAClient
	spaceID := d.Get("space_id").(string)

	assetClient := client.WithSpaceId(spaceID).WithEnvironment(d.Get("environment").(string)).Assets()

	asset, err := buildAsset(d)

	if err != nil {
		return err
	}

	if err = assetClient.Upsert(context.Background(), asset); err != nil {
		return err
	}

	if err = assetClient.Process(context.Background(), asset); err != nil {
		return err
	}

	d.SetId(asset.Sys.ID)

	if err := setAssetProperties(d, asset); err != nil {
		return err
	}

	time.Sleep(1 * time.Second) // avoid race conditions with version mismatches

	if err = setAssetState(d, m); err != nil {
		return err
	}

	return err
}

func resourceUpdateAsset(d *schema.ResourceData, m interface{}) (err error) {
	client := m.(utils.ProviderData).CMAClient
	spaceID := d.Get("space_id").(string)
	assetID := d.Id()

	assetClient := client.WithSpaceId(spaceID).WithEnvironment(d.Get("environment").(string)).Assets()

	_, err = assetClient.Get(context.Background(), assetID)
	if err != nil {
		return err
	}

	asset, err := buildAsset(d)

	if err != nil {
		return err
	}

	if err := assetClient.Upsert(context.Background(), asset); err != nil {
		return err
	}

	if err = assetClient.Process(context.Background(), asset); err != nil {
		return err
	}

	d.SetId(asset.Sys.ID)

	if err := setAssetProperties(d, asset); err != nil {
		return err
	}

	if err = setAssetState(d, m); err != nil {
		return err
	}

	return err
}

func buildAsset(d *schema.ResourceData) (*model.Asset, error) {
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

	fileData := map[string]*model.File{}
	for _, file := range files {
		fileLocale := file.(map[string]any)

		fileData[fileLocale["locale"].(string)] = &model.File{
			URL:         "",
			UploadURL:   "",
			UploadFrom:  nil,
			Details:     nil,
			FileName:    fileLocale["file_name"].(string),
			ContentType: fileLocale["content_type"].(string),
		}

		if url, ok := fileLocale["url"].(string); ok && url != "" {
			fileData[fileLocale["locale"].(string)].URL = url
		}

		if upload, ok := fileLocale["upload"].(string); ok && upload != "" {
			fileData[fileLocale["locale"].(string)].UploadURL = upload
		}

		if details, ok := fileLocale["file_details"].(*model.FileDetails); ok {
			fileData[fileLocale["locale"].(string)].Details = details
		}

		if uploadFrom, ok := fileLocale["upload_from"].(string); ok && uploadFrom != "" {
			fileData[fileLocale["locale"].(string)].UploadFrom = &model.UploadFrom{
				Sys: &model.BaseSys{
					ID: uploadFrom,
				},
			}
		}

	}

	return &model.Asset{
		Sys: &model.PublishSys{
			EnvironmentSys: model.EnvironmentSys{
				SpaceSys: model.SpaceSys{
					CreatedSys: model.CreatedSys{
						BaseSys: model.BaseSys{
							ID:      d.Get("asset_id").(string),
							Version: d.Get("version").(int),
						},
					},
				},
			},
		},
		Fields: &model.AssetFields{
			Title:       localizedTitle,
			Description: localizedDescription,
			File:        fileData,
		},
	}, nil
}

func setAssetState(d *schema.ResourceData, m interface{}) (err error) {
	client := m.(utils.ProviderData).CMAClient
	spaceID := d.Get("space_id").(string)
	assetID := d.Id()

	assetClient := client.WithSpaceId(spaceID).WithEnvironment(d.Get("environment").(string)).Assets()

	ctx := context.Background()

	asset, _ := assetClient.Get(ctx, assetID)

	if d.Get("published").(bool) && asset.Sys.PublishedAt == "" {
		if err = assetClient.Publish(ctx, asset); err != nil {
			return err
		}
	} else if !d.Get("published").(bool) && asset.Sys.PublishedAt != "" {
		if err = assetClient.Unpublish(ctx, asset); err != nil {
			return err
		}
	}

	if d.Get("archived").(bool) && asset.Sys.ArchivedAt == "" {
		if err = assetClient.Archive(ctx, asset); err != nil {
			return err
		}
	} else if !d.Get("archived").(bool) && asset.Sys.ArchivedAt != "" {
		if err = assetClient.Unarchive(ctx, asset); err != nil {
			return err
		}
	}

	err = setAssetProperties(d, asset)
	return err
}

func resourceReadAsset(d *schema.ResourceData, m interface{}) (err error) {
	client := m.(utils.ProviderData).CMAClient
	spaceID := d.Get("space_id").(string)
	assetID := d.Id()

	assetClient := client.WithSpaceId(spaceID).WithEnvironment(d.Get("environment").(string)).Assets()

	asset, err := assetClient.Get(context.Background(), assetID)
	if _, ok := err.(common.NotFoundError); ok {
		d.SetId("")
		return nil
	}

	return setAssetProperties(d, asset)
}

func resourceDeleteAsset(d *schema.ResourceData, m interface{}) (err error) {
	client := m.(utils.ProviderData).CMAClient
	spaceID := d.Get("space_id").(string)
	assetID := d.Id()

	assetClient := client.WithSpaceId(spaceID).WithEnvironment(d.Get("environment").(string)).Assets()

	asset, err := assetClient.Get(context.Background(), assetID)
	if err != nil {
		return err
	}

	return assetClient.Delete(context.Background(), asset)
}

func setAssetProperties(d *schema.ResourceData, asset *model.Asset) (err error) {
	if err = d.Set("space_id", asset.Sys.Space.Sys.ID); err != nil {
		return err
	}

	if err = d.Set("version", asset.Sys.Version); err != nil {
		return err
	}

	return err
}

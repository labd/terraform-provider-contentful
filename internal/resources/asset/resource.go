package asset

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/labd/terraform-provider-contentful/internal/sdk"
	"github.com/labd/terraform-provider-contentful/internal/utils"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &assetResource{}
	_ resource.ResourceWithConfigure   = &assetResource{}
	_ resource.ResourceWithImportState = &assetResource{}
)

func NewAssetResource() resource.Resource {
	return &assetResource{}
}

// assetResource is the resource implementation.
type assetResource struct {
	client *sdk.ClientWithResponses
}

func (e *assetResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_asset"
}

func (e *assetResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Contentful Asset represents a media file in Contentful.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Asset ID",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"asset_id": schema.StringAttribute{
				Required:    true,
				Description: "Asset identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"version": schema.Int64Attribute{
				Computed:    true,
				Description: "The current version of the asset",
			},
			"space_id": schema.StringAttribute{
				Required:    true,
				Description: "Space ID",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"environment": schema.StringAttribute{
				Required:    true,
				Description: "Environment ID",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"published": schema.BoolAttribute{
				Required:    true,
				Description: "Whether the asset is published",
			},
			"archived": schema.BoolAttribute{
				Required:    true,
				Description: "Whether the asset is archived",
			},
		},
		Blocks: map[string]schema.Block{
			"fields": schema.SingleNestedBlock{
				Description: "Asset fields",
				Blocks: map[string]schema.Block{
					"title": schema.ListNestedBlock{
						Description: "Asset title in different locales",
						NestedObject: schema.NestedBlockObject{
							Attributes: map[string]schema.Attribute{
								"content": schema.StringAttribute{
									Required:    true,
									Description: "The title content",
								},
								"locale": schema.StringAttribute{
									Required:    true,
									Description: "The locale code",
								},
							},
						},
					},
					"description": schema.ListNestedBlock{
						Description: "Asset description in different locales",
						NestedObject: schema.NestedBlockObject{
							Attributes: map[string]schema.Attribute{
								"content": schema.StringAttribute{
									Required:    true,
									Description: "The description content",
								},
								"locale": schema.StringAttribute{
									Required:    true,
									Description: "The locale code",
								},
							},
						},
					},
					"file": schema.ListNestedBlock{
						Description: "Asset file information",
						NestedObject: schema.NestedBlockObject{
							Attributes: map[string]schema.Attribute{
								"locale": schema.StringAttribute{
									Required:    true,
									Description: "The locale code",
								},
								"upload": schema.StringAttribute{
									Required:    true,
									Description: "Upload URL or ID",
								},
								"url": schema.StringAttribute{
									Computed:    true,
									Description: "URL of the uploaded file",
								},
								"file_name": schema.StringAttribute{
									Required:    true,
									Description: "File name",
								},
								"content_type": schema.StringAttribute{
									Optional:    true,
									Computed:    true,
									Description: "Content type of the file",
								},
								"filesize": schema.Int64Attribute{
									Computed:    true,
									Description: "File size in bytes",
								},
								"image_width": schema.Int64Attribute{
									Computed:    true,
									Description: "Image width in pixels",
								},
								"image_height": schema.Int64Attribute{
									Computed:    true,
									Description: "Image height in pixels",
								},
							},
						},
					},
				},
			},
		},
	}
}

func (e *assetResource) Configure(_ context.Context, request resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	data := request.ProviderData.(utils.ProviderData)
	e.client = data.Client
}

func (e *assetResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	// Get plan values
	var plan Asset
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, fmt.Sprintf("Asset plan:\n %v", spew.Sdump(plan)))

	// Create the asset
	draft := plan.DraftForCreate()
	resp, err := e.client.UpdateAssetWithResponse(
		ctx,
		plan.SpaceID.ValueString(),
		plan.Environment.ValueString(),
		plan.AssetID.ValueString(),
		nil,
		*draft,
	)
	if err := utils.CheckClientResponse(resp, err, http.StatusCreated); err != nil {
		response.Diagnostics.AddError(
			"Error creating asset",
			"Could not create asset with id: "+err.Error(),
		)
		return
	}
	asset := resp.JSON201

	state := plan
	state.Import(asset)

	tflog.Debug(ctx, fmt.Sprintf("Asset created with ID: %s\n %v", state.ID.ValueString(), spew.Sdump(state)))

	if diag := e.processAsset(ctx, &state); diag != nil {
		response.Diagnostics.Append(diag)
		return
	}

	// Handle publishing/archiving
	if err := e.setAssetState(ctx, &state, &plan); err != nil {
		response.Diagnostics.AddError(
			"Error setting asset state",
			err.Error(),
		)
		return
	}

	// Set state
	response.Diagnostics.Append(response.State.Set(ctx, state)...)
}

func (e *assetResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	// Get current state
	var state Asset
	diags := request.State.Get(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	e.doRead(ctx, &state, &response.State, &response.Diagnostics)
}

func (e *assetResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	// Get plan values
	var plan Asset
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Get current state
	var state Asset
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Create update parameters with version
	params := &sdk.UpdateAssetParams{
		XContentfulVersion: state.Version.ValueInt64(),
	}

	// Update the asset
	draft := plan.DraftForCreate() // Reuse the create draft for updates
	resp, err := e.client.UpdateAssetWithResponse(
		ctx,
		plan.SpaceID.ValueString(),
		plan.Environment.ValueString(),
		plan.AssetID.ValueString(),
		params,
		*draft,
	)
	if err := utils.CheckClientResponse(resp, err, http.StatusOK); err != nil {
		response.Diagnostics.AddError(
			"Error updating asset",
			"Could not update asset: "+err.Error(),
		)
		return
	}

	state.Import(resp.JSON200)

	if diag := e.processAsset(ctx, &state); diag != nil {
		response.Diagnostics.Append(diag)
		return
	}

	// Handle publishing/archiving
	if err := e.setAssetState(ctx, &state, &plan); err != nil {
		response.Diagnostics.AddError(
			"Error setting asset state",
			err.Error(),
		)
		return
	}

	// Set state
	response.Diagnostics.Append(response.State.Set(ctx, state)...)
}

func (e *assetResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	// Get current state
	var state Asset
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Create delete parameters with version
	params := &sdk.DeleteAssetParams{
		XContentfulVersion: state.Version.ValueInt64(),
	}

	// Delete the asset
	resp, err := e.client.DeleteAssetWithResponse(
		ctx,
		state.SpaceID.ValueString(),
		state.Environment.ValueString(),
		state.ID.ValueString(),
		params,
	)
	if err := utils.CheckClientResponse(resp, err, http.StatusNoContent); err != nil {
		if resp.StatusCode() == 404 {
			// Asset not found, no action needed
			tflog.Warn(ctx, fmt.Sprintf("Asset %s not found, no action needed", state.ID.ValueString()))
			return
		}

		response.Diagnostics.AddError(
			"Error deleting asset",
			"Could not delete asset: "+err.Error(),
		)
		return
	}
}

func (e *assetResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	// Extract the asset ID, space ID, and environment ID from the import ID
	idParts, err := utils.ParseThreePartID(request.ID)
	if err != nil {
		response.Diagnostics.AddError(
			"Error importing asset",
			fmt.Sprintf("Expected import format: asset_id:space_id:environment, got: %s", request.ID),
		)
		return
	}

	assetID := idParts[0]
	spaceID := idParts[1]
	environment := idParts[2]

	resp, err := e.client.GetAssetWithResponse(
		ctx,
		spaceID,
		environment,
		assetID,
	)
	if err := utils.CheckClientResponse(resp, err, http.StatusOK); err != nil {
		response.Diagnostics.AddError(
			"Error importing asset",
			fmt.Sprintf("Could not import asset with ID %s: %s", assetID, err.Error()),
		)
		return
	}

	state := &Asset{}
	state.Import(resp.JSON200)
	response.Diagnostics.Append(response.State.Set(ctx, state)...)
}

func (e *assetResource) doRead(ctx context.Context, asset *Asset, state *tfsdk.State, d *diag.Diagnostics) {
	oldState := *asset

	resp, err := e.client.GetAssetWithResponse(
		ctx,
		asset.SpaceID.ValueString(),
		asset.Environment.ValueString(),
		asset.ID.ValueString(),
	)
	if err := utils.CheckClientResponse(resp, err, http.StatusOK); err != nil {
		if resp.StatusCode() == 404 {
			tflog.Warn(ctx, fmt.Sprintf("Asset %s not found", asset.ID.ValueString()))
			return
		}
		d.AddError(
			"Error reading asset",
			"Could not read asset: "+err.Error(),
		)
		return
	}

	// Map response to state
	asset.Import(resp.JSON200)
	asset.CopyInputValues(&oldState)

	// Set state
	d.Append(state.Set(ctx, asset)...)
}

/**
 * processAsset handles the processing of the asset after creation or update.
 * Contentful will inspect the asset and generate a URL and set various other
 * attributes (filesize, image width, image height, etc.) based on the uploaded file.
 *
 * Note that this can also cause conflicts, if the content/type is for example
 * resolved differently by contentful, so we do some copying of the values
 * to avoid provider errors
 */
func (e *assetResource) processAsset(ctx context.Context, state *Asset) diag.Diagnostic {
	oldState := *state

	// Process asset for each locale
	for _, locale := range state.Fields.File {
		resp, err := e.client.ProcessAssetWithResponse(
			ctx,
			state.SpaceID.ValueString(),
			state.Environment.ValueString(),
			state.ID.ValueString(),
			locale.Locale.ValueString(),
		)
		if err := utils.CheckClientResponse(resp, err, http.StatusNoContent); err != nil {
			return diag.NewErrorDiagnostic(
				"Error processing asset",
				fmt.Sprintf("Could not process asset for locale %s: %s", locale.Locale.ValueString(), err.Error()),
			)
		}
	}

	// We should just poll the asset to see if it is done processing
	time.Sleep(2 * time.Second)

	resp, err := e.client.GetAssetWithResponse(
		ctx,
		state.SpaceID.ValueString(),
		state.Environment.ValueString(),
		state.ID.ValueString(),
	)
	if err := utils.CheckClientResponse(resp, err, http.StatusOK); err != nil {
		return diag.NewErrorDiagnostic(
			"Error reading asset",
			"Could not read asset: "+err.Error(),
		)
	}

	state.Import(resp.JSON200)
	state.CopyInputValues(&oldState)

	return nil
}

// setAssetState handles publishing and archiving based on the desired state
func (e *assetResource) setAssetState(ctx context.Context, state *Asset, plan *Asset) error {
	oldState := *state

	// Handle publishing state
	isCurrentlyPublished := state.Published.ValueBool()
	shouldBePublished := plan.Published.ValueBool()

	if shouldBePublished && !isCurrentlyPublished {
		publishParams := &sdk.PublishAssetParams{
			XContentfulVersion: state.Version.ValueInt64(),
		}

		resp, err := e.client.PublishAssetWithResponse(
			ctx,
			state.SpaceID.ValueString(),
			state.Environment.ValueString(),
			state.ID.ValueString(),
			publishParams,
		)
		if err := utils.CheckClientResponse(resp, err, http.StatusOK); err != nil {
			return fmt.Errorf("failed to publish asset: %v", err)
		}
		state.Import(resp.JSON200)

	} else if !shouldBePublished && isCurrentlyPublished {
		unpublishParams := &sdk.UnpublishAssetParams{
			XContentfulVersion: state.Version.ValueInt64(),
		}

		resp, err := e.client.UnpublishAssetWithResponse(
			ctx,
			state.SpaceID.ValueString(),
			state.Environment.ValueString(),
			state.ID.ValueString(),
			unpublishParams,
		)
		if err := utils.CheckClientResponse(resp, err, http.StatusOK); err != nil {
			return fmt.Errorf("failed to unpublish asset: %v", err)
		}
		state.Import(resp.JSON200)
	}

	// Handle archiving state
	isCurrentlyArchived := state.Archived.ValueBool()
	shouldBeArchived := plan.Archived.ValueBool()

	if shouldBeArchived && !isCurrentlyArchived {
		archiveParams := &sdk.ArchiveAssetParams{
			XContentfulVersion: state.Version.ValueInt64(),
		}

		resp, err := e.client.ArchiveAssetWithResponse(
			ctx,
			state.SpaceID.ValueString(),
			state.Environment.ValueString(),
			state.ID.ValueString(),
			archiveParams,
		)
		if err := utils.CheckClientResponse(resp, err, http.StatusOK); err != nil {
			return fmt.Errorf("failed to archive asset: %v", err)
		}
		state.Import(resp.JSON200)
	} else if !shouldBeArchived && isCurrentlyArchived {
		unarchiveParams := &sdk.UnarchiveAssetParams{
			XContentfulVersion: state.Version.ValueInt64(),
		}

		resp, err := e.client.UnarchiveAssetWithResponse(
			ctx,
			state.SpaceID.ValueString(),
			state.Environment.ValueString(),
			state.ID.ValueString(),
			unarchiveParams,
		)
		if err := utils.CheckClientResponse(resp, err, http.StatusOK); err != nil {
			return fmt.Errorf("failed to unarchive asset: %v", err)
		}
		state.Import(resp.JSON200)
	}

	state.CopyInputValues(&oldState)

	return nil
}

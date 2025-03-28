package entry

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/labd/terraform-provider-contentful/internal/sdk"
	"github.com/labd/terraform-provider-contentful/internal/utils"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &entryResource{}
	_ resource.ResourceWithConfigure   = &entryResource{}
	_ resource.ResourceWithImportState = &entryResource{}
)

func NewEntryResource() resource.Resource {
	return &entryResource{}
}

// entryResource is the resource implementation.
type entryResource struct {
	client *sdk.ClientWithResponses
}

func (e *entryResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_entry"
}

func (e *entryResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "A Contentful Entry represents a piece of content in a space.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Entry ID",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"entry_id": schema.StringAttribute{
				Required:    true,
				Description: "Entry identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"version": schema.Int64Attribute{
				Computed:    true,
				Description: "The current version of the entry",
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
			"contenttype_id": schema.StringAttribute{
				Required:    true,
				Description: "Content Type ID",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"published": schema.BoolAttribute{
				Required:    true,
				Description: "Whether the entry is published",
			},
			"archived": schema.BoolAttribute{
				Required:    true,
				Description: "Whether the entry is archived",
			},
		},
		Blocks: map[string]schema.Block{
			"field": schema.ListNestedBlock{
				Description: "Content fields",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Required:    true,
							Description: "Field ID",
						},
						"content": schema.StringAttribute{
							Required:    true,
							Description: "Field content. If the field type is Richtext the content can be passed as stringified JSON.",
						},
						"locale": schema.StringAttribute{
							Required:    true,
							Description: "Locale code",
						},
					},
				},
			},
		},
	}
}

func (e *entryResource) Configure(_ context.Context, request resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	data := request.ProviderData.(utils.ProviderData)
	e.client = data.Client
}

func (e *entryResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	// Get plan values
	var plan Entry
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Create the entry
	draft := plan.Draft()

	var entry *sdk.Entry
	if plan.EntryID.IsUnknown() || plan.EntryID.IsNull() {
		params := &sdk.CreateEntryParams{
			XContentfulContentType: plan.ContentTypeID.ValueString(),
		}
		resp, err := e.client.CreateEntryWithResponse(ctx, plan.SpaceID.ValueString(), plan.Environment.ValueString(), params, draft)
		if err := utils.CheckClientResponse(resp, err, http.StatusCreated); err != nil {
			response.Diagnostics.AddError(
				"Error creating entry",
				"Could not create entry: "+err.Error(),
			)
			return
		}

		entry = resp.JSON201
	} else {
		params := &sdk.UpdateEntryParams{
			XContentfulContentType: plan.ContentTypeID.ValueString(),
		}
		resp, err := e.client.UpdateEntryWithResponse(ctx, plan.SpaceID.ValueString(), plan.Environment.ValueString(), plan.EntryID.ValueString(), params, draft)
		if err := utils.CheckClientResponse(resp, err, http.StatusCreated); err != nil {
			response.Diagnostics.AddError(
				"Error creating entry",
				"Could not create entry: "+err.Error(),
			)
			return
		}

		entry = resp.JSON201
	}

	// Map response to state
	state := Entry{}
	state.Import(entry)

	// Set entry state (published/archived)
	if err := e.setEntryState(ctx, &state, &plan); err != nil {
		response.Diagnostics.AddError(
			"Error setting entry state",
			err.Error(),
		)
		return
	}

	// Set state
	response.Diagnostics.Append(response.State.Set(ctx, state)...)
}

func (e *entryResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	// Get current state
	var state Entry
	diags := request.State.Get(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	e.doRead(ctx, &state, &response.State, &response.Diagnostics)
}

func (e *entryResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	// Get plan values
	var plan Entry
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Get current state
	var state Entry
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Create update parameters with version
	params := &sdk.UpdateEntryParams{
		XContentfulVersion: state.Version.ValueInt64(),
	}

	// Update the entry
	draft := plan.Draft()
	resp, err := e.client.UpdateEntryWithResponse(
		ctx,
		plan.SpaceID.ValueString(),
		plan.Environment.ValueString(),
		plan.ID.ValueString(),
		params,
		draft,
	)

	if err := utils.CheckClientResponse(resp, err, http.StatusOK); err != nil {
		response.Diagnostics.AddError(
			"Error updating entry",
			"Could not update entry: "+err.Error(),
		)
		return
	}

	state.Import(resp.JSON200)

	// Set entry state (published/archived)
	if err := e.setEntryState(ctx, &state, &plan); err != nil {
		response.Diagnostics.AddError(
			"Error setting entry state",
			err.Error(),
		)
		return
	}

	// Set state
	response.Diagnostics.Append(response.State.Set(ctx, state)...)
}

func (e *entryResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	// Get current state
	var state Entry
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Get latest version first to avoid conflicts
	resp, err := e.client.GetEntryWithResponse(
		ctx,
		state.SpaceID.ValueString(),
		state.Environment.ValueString(),
		state.ID.ValueString(),
	)
	if err := utils.CheckClientResponse(resp, err, http.StatusOK); err != nil {
		if resp.StatusCode() == 404 {
			// Entry already deleted, nothing to do
			return
		}

		response.Diagnostics.AddError(
			"Error deleting entry",
			"Could not get latest entry version: "+err.Error(),
		)
		return
	}

	state.Import(resp.JSON200)

	if state.Published.ValueBool() {
		resp, err := e.client.UnpublishEntryWithResponse(
			ctx,
			state.SpaceID.ValueString(),
			state.Environment.ValueString(),
			state.ID.ValueString(),
			&sdk.UnpublishEntryParams{
				XContentfulVersion: state.Version.ValueInt64(),
			},
		)
		if err := utils.CheckClientResponse(resp, err, http.StatusOK); err != nil {
			response.Diagnostics.AddError(
				"Error deleting entry",
				"Could not unpublish entry before deletion: "+err.Error(),
			)
			return
		}
		state.Import(resp.JSON200)
	}

	// Create delete parameters with latest version
	params := &sdk.DeleteEntryParams{
		XContentfulVersion: int64(state.Version.ValueInt64()),
	}

	// Delete the entry
	deleteResp, err := e.client.DeleteEntryWithResponse(
		ctx,
		state.SpaceID.ValueString(),
		state.Environment.ValueString(),
		state.ID.ValueString(),
		params,
	)

	if err != nil {
		response.Diagnostics.AddError(
			"Error deleting entry",
			"Could not delete entry: "+err.Error(),
		)
		return
	}

	if deleteResp.StatusCode() != 204 && deleteResp.StatusCode() != 404 {
		response.Diagnostics.AddError(
			"Error deleting entry",
			fmt.Sprintf("Received unexpected status code: %d", deleteResp.StatusCode()),
		)
		return
	}
}

func (e *entryResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	// Extract the entry ID, space ID, and environment ID from the import ID
	idParts := strings.Split(request.ID, ":")
	if len(idParts) != 3 || idParts[0] == "" || idParts[1] == "" || idParts[2] == "" {
		response.Diagnostics.AddError(
			"Error importing entry",
			fmt.Sprintf("Expected import format: entry_id:space_id:environment, got: %s", request.ID),
		)
		return
	}

	entryID := idParts[0]
	spaceID := idParts[1]
	environment := idParts[2]

	entry := Entry{
		ID:          types.StringValue(entryID),
		EntryID:     types.StringValue(entryID),
		SpaceID:     types.StringValue(spaceID),
		Environment: types.StringValue(environment),
	}

	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("id"), entryID)...)
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("entry_id"), entryID)...)
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("space_id"), spaceID)...)
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("environment"), environment)...)

	e.doRead(ctx, &entry, &response.State, &response.Diagnostics)
}

func (e *entryResource) doRead(ctx context.Context, entry *Entry, state *tfsdk.State, d *diag.Diagnostics) {
	resp, err := e.client.GetEntryWithResponse(
		ctx,
		entry.SpaceID.ValueString(),
		entry.Environment.ValueString(),
		entry.ID.ValueString(),
	)

	if err != nil {
		d.AddError(
			"Error reading entry",
			"Could not read entry: "+err.Error(),
		)
		return
	}

	// Handle 404 Not Found
	if resp.StatusCode() == 404 {
		d.AddWarning(
			"Entry not found",
			fmt.Sprintf("Entry %s was not found, removing from state",
				entry.ID.ValueString()),
		)
		return
	}

	if resp.StatusCode() != 200 {
		d.AddError(
			"Error reading entry",
			fmt.Sprintf("Received unexpected status code: %d", resp.StatusCode()),
		)
		return
	}

	// Map response to state
	entry.Import(resp.JSON200)

	// Set state
	d.Append(state.Set(ctx, entry)...)
}

// setEntryState handles publishing and archiving based on the desired state
func (e *entryResource) setEntryState(ctx context.Context, state *Entry, plan *Entry) error {

	// Handle publishing state
	isCurrentlyPublished := state.Published.ValueBool()
	shouldBePublished := plan.Published.ValueBool()

	// Handle archiving state
	isCurrentlyArchived := state.Archived.ValueBool()
	shouldBeArchived := plan.Archived.ValueBool()

	if shouldBePublished && !isCurrentlyPublished {
		resp, err := e.client.PublishEntryWithResponse(
			ctx,
			state.SpaceID.ValueString(),
			state.Environment.ValueString(),
			state.ID.ValueString(),
			&sdk.PublishEntryParams{
				XContentfulVersion: state.Version.ValueInt64(),
			},
		)
		if err := utils.CheckClientResponse(resp, err, http.StatusOK); err != nil {
			return err
		}
		state.Import(resp.JSON200)
	} else if !shouldBePublished && isCurrentlyPublished {
		resp, err := e.client.UnpublishEntryWithResponse(
			ctx,
			state.SpaceID.ValueString(),
			state.Environment.ValueString(),
			state.ID.ValueString(),
			&sdk.UnpublishEntryParams{
				XContentfulVersion: state.Version.ValueInt64(),
			},
		)
		if err := utils.CheckClientResponse(resp, err, http.StatusOK); err != nil {
			return err
		}
		state.Import(resp.JSON200)
	}

	if shouldBeArchived && !isCurrentlyArchived {
		resp, err := e.client.ArchiveEntryWithResponse(
			ctx,
			state.SpaceID.ValueString(),
			state.Environment.ValueString(),
			state.ID.ValueString(),
			&sdk.ArchiveEntryParams{
				XContentfulVersion: state.Version.ValueInt64(),
			},
		)
		if err := utils.CheckClientResponse(resp, err, http.StatusOK); err != nil {
			return err
		}
		state.Import(resp.JSON200)
	} else if !shouldBeArchived && isCurrentlyArchived {
		resp, err := e.client.UnarchiveEntryWithResponse(
			ctx,
			state.SpaceID.ValueString(),
			state.Environment.ValueString(),
			state.ID.ValueString(),
			&sdk.UnarchiveEntryParams{
				XContentfulVersion: state.Version.ValueInt64(),
			},
		)
		if err := utils.CheckClientResponse(resp, err, http.StatusOK); err != nil {
			return err
		}
		state.Import(resp.JSON200)
	}

	return nil
}

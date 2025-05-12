package preview_environment

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/labd/terraform-provider-contentful/internal/sdk"
	"github.com/labd/terraform-provider-contentful/internal/utils"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &previewEnvironmentResource{}
	_ resource.ResourceWithConfigure   = &previewEnvironmentResource{}
	_ resource.ResourceWithImportState = &previewEnvironmentResource{}
)

func NewPreviewEnvironmentResource() resource.Resource {
	return &previewEnvironmentResource{}
}

// previewEnvironmentResource is the resource implementation.
type previewEnvironmentResource struct {
	client *sdk.ClientWithResponses
}

func (e *previewEnvironmentResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_preview_environment"
}

func (e *previewEnvironmentResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Manages preview environments in Contentful which control how content previews are displayed.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Preview environment ID",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"version": schema.Int64Attribute{
				Computed:    true,
				Description: "The current version of the preview environment",
			},
			"space_id": schema.StringAttribute{
				Required:    true,
				Description: "Space ID",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the preview environment",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "Description of the preview environment",
			},
			"configuration": schema.ListNestedAttribute{
				Required:    true,
				Description: "Configuration for content type previews",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"content_type": schema.StringAttribute{
							Required:    true,
							Description: "Content type ID",
						},
						"enabled": schema.BoolAttribute{
							Required:    true,
							Description: "Whether preview is enabled for this content type",
						},
						"url": schema.StringAttribute{
							Required:    true,
							Description: "URL template for previewing entries of this content type",
						},
					},
				},
			},
		},
	}
}

func (e *previewEnvironmentResource) Configure(_ context.Context, request resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	data := request.ProviderData.(utils.ProviderData)
	e.client = data.Client
}

func (e *previewEnvironmentResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	// Get plan values
	var plan PreviewEnvironment
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Create draft
	draft := plan.Draft()

	// Create the preview environment
	resp, err := e.client.CreatePreviewEnvironmentWithResponse(
		ctx,
		plan.SpaceID.ValueString(),
		*draft,
	)

	if err := utils.CheckClientResponse(resp, err, http.StatusCreated); err != nil {
		response.Diagnostics.AddError(
			"Error creating preview environment",
			"Could not create preview environment, unexpected error: "+err.Error(),
		)
		return
	}

	// Import the response into our model
	plan.Import(resp.JSON201)

	// Set state to fully populated data
	response.Diagnostics.Append(response.State.Set(ctx, plan)...)
}

func (e *previewEnvironmentResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	// Get current state
	var state PreviewEnvironment
	diags := request.State.Get(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	e.doRead(ctx, &state, &response.State, &response.Diagnostics)
}

func (e *previewEnvironmentResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	// Get plan values
	var plan PreviewEnvironment
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Get current state
	var state PreviewEnvironment
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Create update parameters with version
	params := &sdk.UpdatePreviewEnvironmentParams{
		XContentfulVersion: state.Version.ValueInt64(),
	}

	// Create draft
	draft := plan.Draft()

	// Update the preview environment
	resp, err := e.client.UpdatePreviewEnvironmentWithResponse(
		ctx,
		plan.SpaceID.ValueString(),
		plan.ID.ValueString(),
		params,
		*draft,
	)

	if err := utils.CheckClientResponse(resp, err, http.StatusOK); err != nil {
		response.Diagnostics.AddError(
			"Error updating preview environment",
			"Could not update preview environment, unexpected error: "+err.Error(),
		)
		return
	}

	// Import the response into our model
	plan.Import(resp.JSON200)

	// Set state to fully populated data
	response.Diagnostics.Append(response.State.Set(ctx, plan)...)
}

func (e *previewEnvironmentResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	// Get current state
	var state PreviewEnvironment
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Create delete parameters with version
	params := &sdk.DeletePreviewEnvironmentParams{
		XContentfulVersion: state.Version.ValueInt64(),
	}

	// Delete the preview environment
	resp, err := e.client.DeletePreviewEnvironmentWithResponse(
		ctx,
		state.SpaceID.ValueString(),
		state.ID.ValueString(),
		params,
	)

	if err := utils.CheckClientResponse(resp, err, http.StatusNoContent); err != nil {
		response.Diagnostics.AddError(
			"Error deleting preview environment",
			"Could not delete preview environment, unexpected error: "+err.Error(),
		)
		return
	}
}

func (e *previewEnvironmentResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	// Extract the preview environment ID and space ID from the import ID
	idParts := strings.SplitN(request.ID, ":", 2)

	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		response.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: previewEnvironmentId:spaceId. Got: %q", request.ID),
		)
		return
	}

	previewEnvironmentID := idParts[0]
	spaceID := idParts[1]

	// Set the main attributes
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("id"), previewEnvironmentID)...)
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("space_id"), spaceID)...)

	// Create a minimal state to pass to doRead
	previewEnv := &PreviewEnvironment{
		ID:      types.StringValue(previewEnvironmentID),
		SpaceID: types.StringValue(spaceID),
	}

	// Use doRead to populate the rest of the state
	e.doRead(ctx, previewEnv, &response.State, &response.Diagnostics)
}

func (e *previewEnvironmentResource) doRead(ctx context.Context, previewEnv *PreviewEnvironment, state *tfsdk.State, d *diag.Diagnostics) {
	resp, err := e.client.GetPreviewEnvironmentWithResponse(
		ctx,
		previewEnv.SpaceID.ValueString(),
		previewEnv.ID.ValueString(),
	)

	if err != nil {
		d.AddError(
			"Error reading preview environment",
			"Could not read preview environment: "+err.Error(),
		)
		return
	}

	// Handle 404 Not Found
	if resp.StatusCode() == 404 {
		d.AddWarning(
			"Preview environment not found",
			fmt.Sprintf("Preview environment %s was not found, removing from state",
				previewEnv.ID.ValueString()),
		)
		return
	}

	if err := utils.CheckClientResponse(resp, err, http.StatusOK); err != nil {
		d.AddError(
			"Error reading preview environment",
			"Could not read preview environment, unexpected error: "+err.Error(),
		)
		return
	}

	// Import the response into our model
	previewEnv.Import(resp.JSON200)

	// Set state to fully populated data
	d.Append(state.Set(ctx, previewEnv)...)
}

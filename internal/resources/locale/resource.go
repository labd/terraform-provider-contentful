package locale

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"

	"github.com/labd/terraform-provider-contentful/internal/sdk"
	"github.com/labd/terraform-provider-contentful/internal/utils"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &localeResource{}
	_ resource.ResourceWithConfigure   = &localeResource{}
	_ resource.ResourceWithImportState = &localeResource{}
)

func NewLocaleResource() resource.Resource {
	return &localeResource{}
}

// localeResource is the resource implementation.
type localeResource struct {
	client *sdk.ClientWithResponses
}

func (e *localeResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_locale"
}

func (e *localeResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "A Contentful Locale represents a language and region combination.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Locale ID",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"version": schema.Int64Attribute{
				Computed:    true,
				Description: "The current version of the locale",
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
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the locale",
			},
			"code": schema.StringAttribute{
				Required:    true,
				Description: "Locale code (e.g., en-US, de-DE)",
			},
			"fallback_code": schema.StringAttribute{
				Optional:    true,
				Description: "Code of the fallback locale",
			},
			"optional": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether this locale is optional for content",
				Default:     booldefault.StaticBool(false),
			},
			"cda": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether this locale is available in the content delivery API",
				Default:     booldefault.StaticBool(true),
			},
			"cma": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether this locale is available in the content management API",
				Default:     booldefault.StaticBool(false),
			},
		},
	}
}

func (e *localeResource) Configure(_ context.Context, request resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	data := request.ProviderData.(utils.ProviderData)
	e.client = data.Client
}

func (e *localeResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan Locale
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Create the locale
	draft := plan.DraftForCreate()

	resp, err := e.client.CreateLocaleWithResponse(ctx, plan.SpaceID.ValueString(), plan.Environment.ValueString(), draft)
	if err := utils.CheckClientResponse(resp, err, 201); err != nil {
		response.Diagnostics.AddError(
			"Error creating locale",
			"Could not create locale: "+err.Error(),
		)
		return
	}

	state := &Locale{}
	state.Import(resp.JSON201)
	response.Diagnostics.Append(response.State.Set(ctx, state)...)
}

func (e *localeResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state Locale
	diags := request.State.Get(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, err := e.client.GetLocaleWithResponse(
		ctx,
		state.SpaceID.ValueString(),
		state.Environment.ValueString(),
		state.ID.ValueString(),
	)

	if err := utils.CheckClientResponse(resp, err, 200); err != nil {
		if resp.StatusCode() == 404 {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.AddError(
			"Error reading locale",
			"Could not read locale: "+err.Error(),
		)
		return
	}

	state.Import(resp.JSON200)
	response.Diagnostics.Append(response.State.Set(ctx, state)...)
}

func (e *localeResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	// Get plan values
	var plan Locale
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Get current state
	var state Locale
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Create update parameters with version
	params := &sdk.UpdateLocaleParams{
		XContentfulVersion: state.Version.ValueInt64(),
	}

	// Update the locale
	draft := plan.DraftForUpdate()
	resp, err := e.client.UpdateLocaleWithResponse(
		ctx,
		plan.SpaceID.ValueString(),
		plan.Environment.ValueString(),
		state.ID.ValueString(),
		params,
		draft,
	)
	if err := utils.CheckClientResponse(resp, err, 200); err != nil {
		response.Diagnostics.AddError(
			"Error updating locale",
			"Could not update locale: "+err.Error(),
		)
		return
	}

	state.Import(resp.JSON200)
	response.Diagnostics.Append(response.State.Set(ctx, state)...)
}

func (e *localeResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state Locale
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, err := e.client.DeleteLocaleWithResponse(
		ctx,
		state.SpaceID.ValueString(),
		state.Environment.ValueString(),
		state.ID.ValueString(),
	)
	if err := utils.CheckClientResponse(resp, err, 204); err != nil {
		response.Diagnostics.AddError(
			"Error deleting locale",
			"Could not delete locale: "+err.Error(),
		)
		return
	}
}

func (e *localeResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	idParts := strings.Split(request.ID, ":")
	if len(idParts) != 3 || idParts[0] == "" || idParts[1] == "" || idParts[2] == "" {
		response.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: localeId:environment:spaceId. Got: %q", request.ID),
		)
		return
	}

	localeID, environment, spaceID := idParts[0], idParts[1], idParts[2]

	resp, err := e.client.GetLocaleWithResponse(
		ctx,
		spaceID,
		environment,
		localeID,
	)
	if err := utils.CheckClientResponse(resp, err, 200); err != nil {
		response.Diagnostics.AddError(
			"Error importing locale",
			"Could not import locale: "+err.Error(),
		)
		return
	}

	state := &Locale{}
	state.Import(resp.JSON200)
	response.Diagnostics.Append(response.State.Set(ctx, state)...)
}

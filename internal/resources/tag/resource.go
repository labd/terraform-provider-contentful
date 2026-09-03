package tag

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	"github.com/labd/terraform-provider-contentful/internal/sdk"
	"github.com/labd/terraform-provider-contentful/internal/utils"
)

var (
	_ resource.Resource                = &tagResource{}
	_ resource.ResourceWithConfigure   = &tagResource{}
	_ resource.ResourceWithImportState = &tagResource{}
)

func NewTagResource() resource.Resource {
	return &tagResource{}
}

type tagResource struct {
	client *sdk.ClientWithResponses
}

func (e *tagResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_tag"
}

func (e *tagResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "A Contentful tag groups and organizes entries and assets within an environment.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:    true,
				Description: "The unique ID of the tag",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"version": schema.Int64Attribute{
				Computed:    true,
				Description: "The current version of the tag",
			},
			"space_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the space",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"environment": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the environment",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The human-readable unique name of the tag",
			},
			"visibility": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The tag visibility: private tags are CMA-only; public tags are available through all Contentful APIs",
				Validators: []validator.String{
					stringvalidator.OneOf("private", "public"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (e *tagResource) Configure(_ context.Context, request resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	data := request.ProviderData.(utils.ProviderData)
	e.client = data.Client
}

func (e *tagResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan Tag
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	params := &sdk.UpsertTagParams{
		XContentfulTagVisibility: utils.Pointer(plan.VisibilityForCreate()),
	}
	resp, err := e.client.UpsertTagWithResponse(
		ctx,
		plan.SpaceID.ValueString(),
		plan.Environment.ValueString(),
		plan.ID.ValueString(),
		params,
		plan.Draft(),
	)
	if err := utils.CheckClientResponse(resp, err, http.StatusCreated); err != nil {
		response.Diagnostics.AddError(
			"Error creating tag",
			"Could not create tag: "+err.Error(),
		)
		return
	}

	plan.Import(resp.JSON201)
	response.Diagnostics.Append(response.State.Set(ctx, &plan)...)
}

func (e *tagResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state Tag
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, err := e.client.GetTagWithResponse(
		ctx,
		state.SpaceID.ValueString(),
		state.Environment.ValueString(),
		state.ID.ValueString(),
	)
	if err := utils.CheckClientResponse(resp, err, http.StatusOK); err != nil {
		if resp != nil && resp.StatusCode() == http.StatusNotFound {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.AddError(
			"Error reading tag",
			"Could not read tag: "+err.Error(),
		)
		return
	}

	state.Import(resp.JSON200)
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (e *tagResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan Tag
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	var state Tag
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	version := state.Version.ValueInt64()
	params := &sdk.UpsertTagParams{
		XContentfulVersion: &version,
	}
	resp, err := e.client.UpsertTagWithResponse(
		ctx,
		plan.SpaceID.ValueString(),
		plan.Environment.ValueString(),
		plan.ID.ValueString(),
		params,
		plan.Draft(),
	)
	if err != nil {
		response.Diagnostics.AddError(
			"Error updating tag",
			"Could not update tag: "+err.Error(),
		)
		return
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		plan.Import(resp.JSON200)
	case http.StatusCreated:
		plan.Import(resp.JSON201)
	default:
		err := utils.CheckClientResponse(resp, nil, http.StatusOK)
		response.Diagnostics.AddError(
			"Error updating tag",
			"Could not update tag: "+err.Error(),
		)
		return
	}
	response.Diagnostics.Append(response.State.Set(ctx, &plan)...)
}

func (e *tagResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state Tag
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	params := &sdk.DeleteTagParams{
		XContentfulVersion: state.Version.ValueInt64(),
	}
	resp, err := e.client.DeleteTagWithResponse(
		ctx,
		state.SpaceID.ValueString(),
		state.Environment.ValueString(),
		state.ID.ValueString(),
		params,
	)
	if resp != nil && resp.StatusCode() == http.StatusNotFound {
		return
	}
	if err := utils.CheckClientResponse(resp, err, http.StatusNoContent); err != nil {
		response.Diagnostics.AddError(
			"Error deleting tag",
			"Could not delete tag: "+err.Error(),
		)
	}
}

func (e *tagResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	idParts := strings.Split(request.ID, ":")
	if len(idParts) != 3 || idParts[0] == "" || idParts[1] == "" || idParts[2] == "" {
		response.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: tagId:environment:spaceId. Got: %q", request.ID),
		)
		return
	}

	resp, err := e.client.GetTagWithResponse(ctx, idParts[2], idParts[1], idParts[0])
	if err := utils.CheckClientResponse(resp, err, http.StatusOK); err != nil {
		response.Diagnostics.AddError(
			"Error importing tag",
			"Could not import tag: "+err.Error(),
		)
		return
	}

	state := &Tag{}
	state.Import(resp.JSON200)
	response.Diagnostics.Append(response.State.Set(ctx, state)...)
}

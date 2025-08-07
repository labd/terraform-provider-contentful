package taxonomy

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/labd/terraform-provider-contentful/internal/sdk"
	"github.com/labd/terraform-provider-contentful/internal/utils"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &taxonomyConceptResource{}
	_ resource.ResourceWithConfigure   = &taxonomyConceptResource{}
	_ resource.ResourceWithImportState = &taxonomyConceptResource{}
)

func NewTaxonomyConceptResource() resource.Resource {
	return &taxonomyConceptResource{}
}

// taxonomyConceptResource is the resource implementation.
type taxonomyConceptResource struct {
	client *sdk.ClientWithResponses
}

func (r *taxonomyConceptResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_taxonomy_concept"
}

func (r *taxonomyConceptResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "A Contentful taxonomy concept represents a categorization term that can be used to organize and classify content.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Taxonomy concept ID",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"version": schema.Int64Attribute{
				Computed:    true,
				Description: "The current version of the taxonomy concept",
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
			"concept_scheme_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the concept scheme this concept belongs to",
			},
			"pref_label": schema.MapAttribute{
				ElementType: types.StringType,
				Required:    true,
				Description: "Preferred label for the concept in different locales",
			},
			"alt_label": schema.MapAttribute{
				ElementType: types.ListType{ElemType: types.StringType},
				Optional:    true,
				Description: "Alternative labels for the concept in different locales",
			},
			"definition": schema.MapAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "Definition of the concept in different locales",
			},
			"note": schema.MapAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "Notes about the concept in different locales",
			},
			"notations": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "Notation codes for the concept",
			},
			"broader": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "IDs of broader concepts in the taxonomy hierarchy",
			},
			"narrower": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "IDs of narrower concepts in the taxonomy hierarchy",
			},
			"related": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "IDs of related concepts",
			},
		},
	}
}

func (r *taxonomyConceptResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	providerData, ok := request.ProviderData.(utils.ProviderData)
	if !ok {
		response.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected utils.ProviderData, got: %T. Please report this issue to the provider developers.", request.ProviderData),
		)
		return
	}

	r.client = providerData.Client
}

func (r *taxonomyConceptResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data TaxonomyConcept

	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	draft := data.DraftForCreate()

	result, err := r.client.CreateTaxonomyConceptWithResponse(
		ctx,
		data.SpaceID.ValueString(),
		data.Environment.ValueString(),
		&sdk.CreateTaxonomyConceptParams{},
		draft,
	)

	if err != nil {
		response.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create taxonomy concept, got error: %s", err))
		return
	}

	if result.StatusCode() != 201 {
		response.Diagnostics.AddError("API Error", fmt.Sprintf("Unable to create taxonomy concept, status: %d, body: %s", result.StatusCode(), string(result.Body)))
		return
	}

	data.Import(result.JSON201)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *taxonomyConceptResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data TaxonomyConcept

	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	result, err := r.client.GetTaxonomyConceptWithResponse(
		ctx,
		data.SpaceID.ValueString(),
		data.Environment.ValueString(),
		data.ID.ValueString(),
	)

	if err != nil {
		response.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read taxonomy concept, got error: %s", err))
		return
	}

	if result.StatusCode() == 404 {
		response.State.RemoveResource(ctx)
		return
	}

	if result.StatusCode() != 200 {
		response.Diagnostics.AddError("API Error", fmt.Sprintf("Unable to read taxonomy concept, status: %d, body: %s", result.StatusCode(), string(result.Body)))
		return
	}

	data.Import(result.JSON200)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *taxonomyConceptResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var data TaxonomyConcept

	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	draft := data.DraftForUpdate()

	result, err := r.client.UpdateTaxonomyConceptWithResponse(
		ctx,
		data.SpaceID.ValueString(),
		data.Environment.ValueString(),
		data.ID.ValueString(),
		&sdk.UpdateTaxonomyConceptParams{
			XContentfulVersion: data.Version.ValueInt64(),
		},
		draft,
	)

	if err != nil {
		response.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update taxonomy concept, got error: %s", err))
		return
	}

	if result.StatusCode() != 200 {
		response.Diagnostics.AddError("API Error", fmt.Sprintf("Unable to update taxonomy concept, status: %d, body: %s", result.StatusCode(), string(result.Body)))
		return
	}

	data.Import(result.JSON200)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *taxonomyConceptResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data TaxonomyConcept

	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	result, err := r.client.DeleteTaxonomyConceptWithResponse(
		ctx,
		data.SpaceID.ValueString(),
		data.Environment.ValueString(),
		data.ID.ValueString(),
		&sdk.DeleteTaxonomyConceptParams{
			XContentfulVersion: data.Version.ValueInt64(),
		},
	)

	if err != nil {
		response.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete taxonomy concept, got error: %s", err))
		return
	}

	if result.StatusCode() != 204 && result.StatusCode() != 404 {
		response.Diagnostics.AddError("API Error", fmt.Sprintf("Unable to delete taxonomy concept, status: %d, body: %s", result.StatusCode(), string(result.Body)))
		return
	}
}

func (r *taxonomyConceptResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	idParts := strings.Split(request.ID, ":")
	if len(idParts) != 3 || idParts[0] == "" || idParts[1] == "" || idParts[2] == "" {
		response.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: space_id:environment:concept_id. Got: %q", request.ID),
		)
		return
	}

	spaceID, environment, conceptID := idParts[0], idParts[1], idParts[2]

	resp, err := r.client.GetTaxonomyConceptWithResponse(
		ctx,
		spaceID,
		environment,
		conceptID,
	)
	if err := utils.CheckClientResponse(resp, err, 200); err != nil {
		response.Diagnostics.AddError(
			"Error importing taxonomy concept",
			"Could not import taxonomy concept: "+err.Error(),
		)
		return
	}

	state := &TaxonomyConcept{}
	state.Import(resp.JSON200)
	response.Diagnostics.Append(response.State.Set(ctx, state)...)
}
package app_installation

import (
	"context"
	_ "embed"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
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
	_ resource.Resource                = &appInstallationResource{}
	_ resource.ResourceWithConfigure   = &appInstallationResource{}
	_ resource.ResourceWithImportState = &appInstallationResource{}
)

func NewAppInstallationResource() resource.Resource {
	return &appInstallationResource{}
}

// appInstallationResource is the resource implementation.
type appInstallationResource struct {
	client         *sdk.ClientWithResponses
	organizationId string
}

func (e *appInstallationResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_app_installation"
}

func (e *appInstallationResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Todo for explaining app installation",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"app_definition_id": schema.StringAttribute{
				Required:    true,
				Description: "app definition id",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"space_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description: "space id",
			},
			"environment": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"parameters": schema.StringAttribute{
				CustomType:  jsontypes.NormalizedType{},
				Description: "Parameters needed for the installation of the app in the given space, like credentials or other configuration parameters",
				Required:    true,
			},
			"accepted_terms": schema.ListAttribute{
				Optional:    true,
				Description: "List of needed terms to accept to install the app",
				ElementType: types.StringType,
			},
		},
	}
}

func (e *appInstallationResource) Configure(_ context.Context, request resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	data := request.ProviderData.(utils.ProviderData)
	e.client = data.Client

}

func (e *appInstallationResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan *AppInstallation
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	draft := plan.Draft()

	params := &sdk.UpsertAppInstallationParams{
		XContentfulMarketplace: plan.AcceptedTermsHeader(),
	}
	resp, err := e.client.UpsertAppInstallationWithResponse(
		ctx, plan.SpaceId.ValueString(), plan.Environment.ValueString(), plan.AppDefinitionID.ValueString(), params, *draft)

	if err != nil {
		response.Diagnostics.AddError("Error creating app_installation", err.Error())
		return
	}

	if resp.StatusCode() != 200 {
		response.Diagnostics.AddError("Error creating app_installation", fmt.Sprintf("Unexpected status code: %d", resp.StatusCode()))
		return
	}

	plan.ID = plan.AppDefinitionID

	// Set state to fully populated data
	response.Diagnostics.Append(response.State.Set(ctx, plan)...)
	if response.Diagnostics.HasError() {
		return
	}
}

func (e *appInstallationResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	// Get current state
	var state *AppInstallation
	diags := request.State.Get(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	e.doRead(ctx, state, &response.State, &response.Diagnostics)
}

func (e *appInstallationResource) doRead(ctx context.Context, app *AppInstallation, state *tfsdk.State, d *diag.Diagnostics) {

	appDefinition, err := e.getAppInstallation(ctx, app)
	if err != nil {
		d.AddError(
			"Error reading app installation",
			"Could not retrieve app installation, unexpected error: "+err.Error(),
		)
		return
	}

	app.Import(appDefinition)

	// Set refreshed state
	d.Append(state.Set(ctx, &app)...)
	if d.HasError() {
		return
	}
}

func (e *appInstallationResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan *AppInstallation
	diags := request.Plan.Get(ctx, &plan)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	// Get current state
	var state *AppInstallation
	diags = request.State.Get(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	appDefinition, err := e.getAppInstallation(ctx, plan)
	if err != nil {
		response.Diagnostics.AddError(
			"Error reading app installation",
			"Could not retrieve app installation, unexpected error: "+err.Error(),
		)
		return
	}

	if !plan.Equal(appDefinition) {

		draft := plan.Draft()

		params := &sdk.UpsertAppInstallationParams{
			XContentfulMarketplace: plan.AcceptedTermsHeader(),
		}
		resp, err := e.client.UpsertAppInstallationWithResponse(
			ctx, plan.SpaceId.ValueString(), plan.Environment.ValueString(), plan.AppDefinitionID.ValueString(), params, *draft)

		if err != nil {
			response.Diagnostics.AddError("Error updating app_installation", err.Error())
			return
		}

		if resp.StatusCode() != 201 {
			response.Diagnostics.AddError("Error updating app_installation", fmt.Sprintf("Unexpected status code: %d", resp.StatusCode()))
			return
		}
	}

	e.doRead(ctx, plan, &response.State, &response.Diagnostics)
}

func (e *appInstallationResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	// Get current state
	var state *AppInstallation
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)

	resp, err := e.client.DeleteAppInstallationWithResponse(
		ctx, state.SpaceId.ValueString(), state.Environment.ValueString(), state.ID.ValueString(), nil)

	if err != nil {
		response.Diagnostics.AddError(
			"Error deleting app_installation",
			"Could not delete app_installation, unexpected error: "+err.Error(),
		)
		return
	}

	if resp.StatusCode() != 204 {
		response.Diagnostics.AddError(
			"Error deleting app_installation",
			fmt.Sprintf("Unexpected status code: %d", resp.StatusCode()),
		)
		return
	}
}

func (e *appInstallationResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	idParts := strings.SplitN(request.ID, ":", 3)

	if len(idParts) != 3 || idParts[0] == "" || idParts[1] == "" || idParts[2] == "" {
		response.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: appDefinitionID:env:spaceId. Got: %q", request.ID),
		)
		return
	}

	futureState := &AppInstallation{
		AppDefinitionID: types.StringValue(idParts[0]),
		SpaceId:         types.StringValue(idParts[2]),
		Environment:     types.StringValue(idParts[1]),
	}

	e.doRead(ctx, futureState, &response.State, &response.Diagnostics)
}

func (e *appInstallationResource) getAppInstallation(ctx context.Context, app *AppInstallation) (*sdk.AppInstallation, error) {

	resp, err := e.client.GetAppInstallationWithResponse(
		ctx, app.SpaceId.ValueString(), app.Environment.ValueString(), app.ID.ValueString())
	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("Unexpected status code: %d", resp.StatusCode())
	}

	return resp.JSON200, nil
}

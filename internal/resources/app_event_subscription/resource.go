package app_event_subscription

import (
	"context"
	_ "embed"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"net/http"

	"github.com/labd/terraform-provider-contentful/internal/sdk"
	"github.com/labd/terraform-provider-contentful/internal/utils"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &appEventSubscriptionResource{}
	_ resource.ResourceWithConfigure   = &appEventSubscriptionResource{}
	_ resource.ResourceWithImportState = &appEventSubscriptionResource{}
)

func NewAppEventSubscriptionResource() resource.Resource {
	return &appEventSubscriptionResource{}
}

// appEventSubscriptionResource is the resource implementation.
type appEventSubscriptionResource struct {
	client         *sdk.ClientWithResponses
	clientUpload   *sdk.ClientWithResponses
	organizationId string
}

func createID(organizationId, appDefinitionId string) string {
	return organizationId + ":" + appDefinitionId
}

func (e *appEventSubscriptionResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_app_event_subscription"
}

func (e *appEventSubscriptionResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Use app events to be notified about changes in the environments your app is installed in.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "app definition id",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"app_definition_id": schema.StringAttribute{
				Description: "The ID of the app definition this event subscription belongs to.",
				Required:    true,
			},
			"target_url": schema.StringAttribute{
				Description: "The URL to which the events will be sent.",
				Required:    true,
			},
			"topics": schema.ListAttribute{
				Description: "The list of topics for which the app will receive events. See [the docs](https://www.contentful.com/developers/docs/references/content-management-api/#/reference/app-event-subscriptions/app-event-subscription) for which topics are available",
				Required:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (e *appEventSubscriptionResource) Configure(_ context.Context, request resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	data := request.ProviderData.(utils.ProviderData)
	e.client = data.Client
	e.organizationId = data.OrganizationId
}

func (e *appEventSubscriptionResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan AppEventSubscription
	var state AppEventSubscription
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	draft := plan.Draft()

	state = AppEventSubscription{
		AppDefinitionID: plan.AppDefinitionID,
	}

	gResp, err := e.client.GetAppEventSubscriptionWithResponse(ctx, e.organizationId, plan.AppDefinitionID.ValueString())
	if err != nil {
		response.Diagnostics.AddError("Error reading app event subscription", err.Error())
		return
	}
	if gResp.StatusCode() == http.StatusOK {
		response.Diagnostics.AddError(
			"App event subscription already exists",
			fmt.Sprintf("An app event subscription with ID '%s'. Please import it instead", createID(e.organizationId, plan.AppDefinitionID.ValueString())),
		)
		return
	}

	resp, err := e.client.UpdateAppEventSubscriptionWithResponse(ctx, e.organizationId, plan.AppDefinitionID.ValueString(), *draft)
	if err := utils.CheckClientResponse(resp, err, http.StatusCreated); err != nil {
		response.Diagnostics.AddError("Error creating app_event_subscription", err.Error())
		return
	}
	Import(&state, resp.JSON201)

	state.ID = types.StringValue(createID(e.organizationId, state.AppDefinitionID.ValueString()))

	// Set state to fully populated data
	response.Diagnostics.Append(response.State.Set(ctx, state)...)
	if response.Diagnostics.HasError() {
		return
	}
}

func (e *appEventSubscriptionResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	// Get current state
	var state *AppEventSubscription
	diags := request.State.Get(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	e.doRead(ctx, state, &response.State, &response.Diagnostics)
}

func (e *appEventSubscriptionResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan *AppEventSubscription
	diags := request.Plan.Get(ctx, &plan)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	// Get current state
	var state *AppEventSubscription
	diags = request.State.Get(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	if !plan.Equal(state) {
		draft := plan.Draft()
		resp, err := e.client.UpdateAppEventSubscriptionWithResponse(ctx, e.organizationId, plan.AppDefinitionID.ValueString(), *draft)
		if err := utils.CheckClientResponse(resp, err, http.StatusOK); err != nil {
			response.Diagnostics.AddError("Error updating app_event_subscription", err.Error())
			return
		}

		plan.ID = types.StringValue(createID(e.organizationId, state.AppDefinitionID.ValueString()))
	}

	e.doRead(ctx, plan, &response.State, &response.Diagnostics)
}

func (e *appEventSubscriptionResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state *AppEventSubscription
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)

	resp, err := e.client.DeleteAppEventSubscriptionWithResponse(ctx, e.organizationId, state.AppDefinitionID.ValueString())
	if err := utils.CheckClientResponse(resp, err, http.StatusNoContent); err != nil {
		response.Diagnostics.AddError(
			"Error deleting app_event_subscription",
			"Could not delete app_event_subscription, unexpected error: "+err.Error(),
		)
		return
	}
}

func (e *appEventSubscriptionResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	futureState := &AppEventSubscription{
		ID: types.StringValue(request.ID),
	}

	e.doRead(ctx, futureState, &response.State, &response.Diagnostics)
}

func (e *appEventSubscriptionResource) doRead(ctx context.Context, app *AppEventSubscription, state *tfsdk.State, d *diag.Diagnostics) {
	resp, err := e.client.GetAppEventSubscriptionWithResponse(ctx, e.organizationId, app.AppDefinitionID.ValueString())
	if err := utils.CheckClientResponse(resp, err, http.StatusOK); err != nil {
		d.AddError(
			"Error reading app event subscription",
			fmt.Sprintf("Could not retrieve app event subscription, unexpected error: %s", err.Error()),
		)
		return
	}

	Import(app, resp.JSON200)
	app.AppDefinitionID = types.StringValue(app.AppDefinitionID.ValueString())
	app.ID = types.StringValue(createID(e.organizationId, app.AppDefinitionID.ValueString()))

	// Set refreshed state
	d.Append(state.Set(ctx, &app)...)
	if d.HasError() {
		return
	}
}

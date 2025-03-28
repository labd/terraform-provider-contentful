package webhook

import (
	"context"
	"fmt"
	"strings"

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
	_ resource.Resource                = &webhookResource{}
	_ resource.ResourceWithConfigure   = &webhookResource{}
	_ resource.ResourceWithImportState = &webhookResource{}
)

func NewWebhookResource() resource.Resource {
	return &webhookResource{}
}

// webhookResource is the resource implementation.
type webhookResource struct {
	client *sdk.ClientWithResponses
}

func (e *webhookResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_webhook"
}

func (e *webhookResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "A Contentful Webhook represents a webhook that can be used to notify external services of changes in a space.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Webhook ID",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"version": schema.Int64Attribute{
				Computed:    true,
				Description: "The current version of the webhook",
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
				Description: "Name of the webhook",
			},
			"url": schema.StringAttribute{
				Required:    true,
				Description: "URL to notify",
			},
			"http_basic_auth_username": schema.StringAttribute{
				Optional:    true,
				Description: "HTTP basic auth username",
			},
			"http_basic_auth_password": schema.StringAttribute{
				Optional:    true,
				Description: "HTTP basic auth password",
				Sensitive:   true,
			},
			"headers": schema.MapAttribute{
				Optional:    true,
				Description: "HTTP headers to send with the webhook request",
				ElementType: types.StringType,
			},
			"topics": schema.ListAttribute{
				Required:    true,
				Description: "List of topics this webhook should be triggered for",
				ElementType: types.StringType,
			},
		},
	}
}

func (e *webhookResource) Configure(_ context.Context, request resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	data := request.ProviderData.(utils.ProviderData)
	e.client = data.Client
}

func (e *webhookResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	// Get plan values
	var plan Webhook
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Create the webhook
	draft := plan.DraftForCreate()

	resp, err := e.client.CreateWebhookWithResponse(ctx, plan.SpaceId.ValueString(), draft)
	if err != nil {
		response.Diagnostics.AddError(
			"Error creating webhook",
			"Could not create webhook: "+err.Error(),
		)
		return
	}

	if resp.StatusCode() != 201 {
		response.Diagnostics.AddError(
			"Error creating webhook",
			fmt.Sprintf("Received unexpected status code: %d", resp.StatusCode()),
		)
		return
	}

	// Map response to state
	plan.Import(resp.JSON201)

	// Set state
	response.Diagnostics.Append(response.State.Set(ctx, plan)...)
}

func (e *webhookResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	// Get current state
	var state Webhook
	diags := request.State.Get(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	e.doRead(ctx, &state, &response.State, &response.Diagnostics)
}

func (e *webhookResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	// Get plan values
	var plan Webhook
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Get current state
	var state Webhook
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Create update parameters with version
	params := &sdk.UpdateWebhookParams{
		XContentfulVersion: state.Version.ValueInt64(),
	}

	// Update the webhook
	draft := plan.DraftForUpdate()
	resp, err := e.client.UpdateWebhookWithResponse(
		ctx,
		plan.SpaceId.ValueString(),
		state.ID.ValueString(),
		params,
		draft,
	)

	if err != nil {
		response.Diagnostics.AddError(
			"Error updating webhook",
			"Could not update webhook: "+err.Error(),
		)
		return
	}

	if resp.StatusCode() != 200 {
		response.Diagnostics.AddError(
			"Error updating webhook",
			fmt.Sprintf("Received unexpected status code: %d", resp.StatusCode()),
		)
		return
	}

	// Map response to state
	plan.Import(resp.JSON200)

	// Set state
	response.Diagnostics.Append(response.State.Set(ctx, plan)...)
}

func (e *webhookResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	// Get current state
	var state Webhook
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Create delete parameters with version
	params := &sdk.DeleteWebhookParams{
		XContentfulVersion: state.Version.ValueInt64(),
	}

	// Delete the webhook
	resp, err := e.client.DeleteWebhookWithResponse(
		ctx,
		state.SpaceId.ValueString(),
		state.ID.ValueString(),
		params,
	)

	if err != nil {
		response.Diagnostics.AddError(
			"Error deleting webhook",
			"Could not delete webhook: "+err.Error(),
		)
		return
	}

	if resp.StatusCode() != 204 {
		response.Diagnostics.AddError(
			"Error deleting webhook",
			fmt.Sprintf("Received unexpected status code: %d", resp.StatusCode()),
		)
		return
	}
}

func (e *webhookResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	idParts := strings.SplitN(request.ID, ":", 2)

	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		response.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: webhookId:spaceId. Got: %q", request.ID),
		)
		return
	}

	futureState := &Webhook{
		ID:      types.StringValue(idParts[0]),
		SpaceId: types.StringValue(idParts[1]),
	}

	e.doRead(ctx, futureState, &response.State, &response.Diagnostics)
}

func (e *webhookResource) doRead(ctx context.Context, webhook *Webhook, state *tfsdk.State, d *diag.Diagnostics) {
	resp, err := e.client.GetWebhookWithResponse(ctx, webhook.SpaceId.ValueString(), webhook.ID.ValueString())
	if err != nil {
		d.AddError(
			"Error reading webhook",
			"Could not read webhook: "+err.Error(),
		)
		return
	}

	// Handle 404 Not Found
	if resp.StatusCode() == 404 {
		d.AddWarning(
			"Webhook not found",
			fmt.Sprintf("Webhook %s in space %s was not found, removing from state",
				webhook.ID.ValueString(), webhook.SpaceId.ValueString()),
		)
		return
	}

	if resp.StatusCode() != 200 {
		d.AddError(
			"Error reading webhook",
			fmt.Sprintf("Received unexpected status code: %d", resp.StatusCode()),
		)
		return
	}

	// Map response to state
	webhook.Import(resp.JSON200)

	// Set state
	d.Append(state.Set(ctx, webhook)...)
}

package webhook

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
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
				Optional:           true,
				Description:        "HTTP basic auth username",
				DeprecationMessage: "Setting Basic Auth credentials is not supported anymore. Please use the headers object instead.",
			},
			"http_basic_auth_password": schema.StringAttribute{
				Optional:           true,
				Description:        "HTTP basic auth password",
				Sensitive:          true,
				DeprecationMessage: "Setting Basic Auth credentials is not supported anymore. Please use the headers object instead.",
			},
			"headers": schema.MapAttribute{
				Optional:           true,
				Description:        "HTTP headers to send with the webhook request",
				DeprecationMessage: "Use headers_v2 instead. headers_v2 supports marking individual header values as secret.",
				ElementType:        types.StringType,
				Validators: []validator.Map{
					mapvalidator.ConflictsWith(path.MatchRoot("headers_v2")),
				},
			},
			"headers_v2": schema.SetNestedAttribute{
				Optional:    true,
				Description: "HTTP headers to send with the webhook request. Supports marking individual header values as secret.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"key": schema.StringAttribute{
							Required:    true,
							Description: "Header name",
						},
						"value": schema.StringAttribute{
							Required:    true,
							Description: "Header value",
							Sensitive:   true,
						},
						"secret": schema.BoolAttribute{
							Optional:    true,
							Computed:    true,
							Description: "Whether the header value is secret. Secret header values are not returned by the API.",
							Default:     booldefault.StaticBool(false),
						},
					},
				},
				Validators: []validator.Set{
					setvalidator.ConflictsWith(path.MatchRoot("headers")),
				},
			},
			"topics": schema.ListAttribute{
				Required:    true,
				Description: "List of topics this webhook should be triggered for",
				ElementType: types.StringType,
			},
			"active": schema.BoolAttribute{
				Computed:    true,
				Optional:    true,
				Description: "Whether the webhook is active or not",
				Default:     booldefault.StaticBool(true),
			},
			"filters": schema.StringAttribute{
				Computed: true,
				Optional: true,
				Description: "List of filters this webhook should match for before triggering. The filters should be " +
					"provided as a JSON string. For example: {\"sys\":{\"type\":\"Entry\"}}",
				Default: stringdefault.StaticString("[]"),
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
	draft, err := plan.DraftForCreate()
	if err != nil {
		response.Diagnostics.AddError(
			"Error creating webhook",
			"Could not create webhook: "+err.Error(),
		)
		return
	}

	if draft.HttpBasicPassword != nil || draft.HttpBasicUsername != nil {
		response.Diagnostics.AddError(
			"Error creating webhook",
			"Setting Basic Auth credentials is not supported anymore. Please use the headers object instead.",
		)
		return
	}

	resp, err := e.client.CreateWebhookWithResponse(ctx, plan.SpaceId.ValueString(), draft)
	if err := utils.CheckClientResponse(resp, err, http.StatusOK); err != nil {
		response.Diagnostics.AddError(
			"Error creating webhook",
			"Could not create webhook: "+err.Error(),
		)
		return
	}

	// Map response to state
	state := &Webhook{}
	err = state.MapFromSDK(resp.JSON200)
	if err != nil {
		response.Diagnostics.AddError(
			"Error mapping webhook",
			"Could not import webhook: "+err.Error(),
		)
		return
	}

	// If the plan used headers_v2, repopulate HeadersV2 from the API response and preserve
	// secret header values (the API does not return secret values).
	if plan.HeadersV2 != nil {
		state.Headers = nil
		if len(resp.JSON200.Headers) > 0 {
			state.HeadersV2 = populateV2HeadersFromSDK(resp.JSON200.Headers)
		} else {
			state.HeadersV2 = make([]WebhookHeader, 0)
		}
		state.PreserveSecretHeaderValues(plan.HeadersV2)
	} else {
		state.HeadersV2 = nil
	}

	if !plan.HttpBasicAuthPassword.IsNull() {
		state.HttpBasicAuthPassword = plan.HttpBasicAuthPassword
	} else {
		state.HttpBasicAuthPassword = types.StringNull()
	}
	if !plan.HttpBasicAuthUsername.IsNull() {
		state.HttpBasicAuthUsername = plan.HttpBasicAuthUsername
	} else {
		state.HttpBasicAuthUsername = types.StringNull()
	}

	// Set state
	response.Diagnostics.Append(response.State.Set(ctx, state)...)
}

func (e *webhookResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	// Get current state
	var state Webhook
	diags := request.State.Get(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, err := e.client.GetWebhookWithResponse(ctx, state.SpaceId.ValueString(), state.ID.ValueString())
	if err := utils.CheckClientResponse(resp, err, http.StatusOK); err != nil {
		if resp.StatusCode() == 404 {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.AddError("Error reading webhook", err.Error())
	}

	// Capture whether headers_v2 was in use and copy the prior values for secret preservation.
	usingV2 := state.HeadersV2 != nil
	priorHeadersV2 := make([]WebhookHeader, len(state.HeadersV2))
	copy(priorHeadersV2, state.HeadersV2)

	err = state.MapFromSDK(resp.JSON200)
	if err != nil {
		response.Diagnostics.AddError(
			"Error mapping webhook",
			"Could not import webhook: "+err.Error(),
		)
		return
	}

	// Restore headers_v2 from the API response and preserve secret values.
	if usingV2 {
		state.Headers = nil
		if len(resp.JSON200.Headers) > 0 {
			state.HeadersV2 = populateV2HeadersFromSDK(resp.JSON200.Headers)
		} else {
			state.HeadersV2 = make([]WebhookHeader, 0)
		}
		state.PreserveSecretHeaderValues(priorHeadersV2)
	} else {
		state.HeadersV2 = nil
	}

	response.Diagnostics.Append(response.State.Set(ctx, state)...)
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
	draft, err := plan.DraftForUpdate()
	if err != nil {
		response.Diagnostics.AddError(
			"Error updating webhook",
			"Could not update webhook: "+err.Error(),
		)
		return
	}

	if draft.HttpBasicPassword != nil || draft.HttpBasicUsername != nil {
		response.Diagnostics.AddError(
			"Error updating webhook",
			"Setting Basic Auth credentials is not supported anymore. Please use the headers object instead.",
		)
		return
	}

	resp, err := e.client.UpdateWebhookWithResponse(
		ctx,
		plan.SpaceId.ValueString(),
		state.ID.ValueString(),
		params,
		draft,
	)
	if err := utils.CheckClientResponse(resp, err, http.StatusOK); err != nil {
		response.Diagnostics.AddError(
			"Error updating webhook",
			"Could not update webhook: "+err.Error(),
		)
		return
	}

	err = state.MapFromSDK(resp.JSON200)
	if err != nil {
		response.Diagnostics.AddError(
			"Error mapping webhook",
			"Could not import webhook: "+err.Error(),
		)
		return
	}

	// If the plan used headers_v2, repopulate HeadersV2 from the API response and preserve
	// secret header values (the API does not return secret values).
	if plan.HeadersV2 != nil {
		state.Headers = nil
		if len(resp.JSON200.Headers) > 0 {
			state.HeadersV2 = populateV2HeadersFromSDK(resp.JSON200.Headers)
		} else {
			state.HeadersV2 = make([]WebhookHeader, 0)
		}
		state.PreserveSecretHeaderValues(plan.HeadersV2)
	} else {
		state.HeadersV2 = nil
	}

	if !plan.HttpBasicAuthPassword.IsNull() {
		state.HttpBasicAuthPassword = plan.HttpBasicAuthPassword
	} else {
		state.HttpBasicAuthPassword = types.StringNull()
	}
	if !plan.HttpBasicAuthUsername.IsNull() {
		state.HttpBasicAuthUsername = plan.HttpBasicAuthUsername
	} else {
		state.HttpBasicAuthUsername = types.StringNull()
	}

	response.Diagnostics.Append(response.State.Set(ctx, state)...)
}

func (e *webhookResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
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
	if err := utils.CheckClientResponse(resp, err, http.StatusNoContent); err != nil {
		response.Diagnostics.AddError(
			"Error deleting webhook",
			"Could not delete webhook: "+err.Error(),
		)
		return
	}
}

func (e *webhookResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	idParts := strings.SplitN(request.ID, ":", 2)

	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		response.Diagnostics.AddError(
			"Unexpected MapFromSDK Identifier",
			fmt.Sprintf("Expected import identifier with format: webhookId:spaceId. Got: %q", request.ID),
		)
		return
	}

	id := idParts[0]
	spaceId := idParts[1]

	resp, err := e.client.GetWebhookWithResponse(ctx, spaceId, id)
	if err := utils.CheckClientResponse(resp, err, http.StatusOK); err != nil {
		response.Diagnostics.AddError(
			"Error reading webhook",
			"Could not read webhook: "+err.Error(),
		)
		return
	}

	state := &Webhook{}
	err = state.MapFromSDK(resp.JSON200)
	if err != nil {
		response.Diagnostics.AddError(
			"Error mapping webhook",
			"Could not import webhook: "+err.Error(),
		)
		return
	}
	// HeadersV2 is intentionally left nil on import; users must configure it explicitly.
	state.HeadersV2 = nil
	response.Diagnostics.Append(response.State.Set(ctx, state)...)
}

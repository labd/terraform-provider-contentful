package oauth_application

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/labd/terraform-provider-contentful/internal/custommodifier"
	"github.com/labd/terraform-provider-contentful/internal/sdk"
	"github.com/labd/terraform-provider-contentful/internal/utils"
)

var (
	_ resource.Resource                = &oauthApplicationResource{}
	_ resource.ResourceWithConfigure   = &oauthApplicationResource{}
	_ resource.ResourceWithImportState = &oauthApplicationResource{}
)

func NewOAuthApplicationResource() resource.Resource {
	return &oauthApplicationResource{}
}

type oauthApplicationResource struct {
	client *sdk.ClientWithResponses
}

func (e *oauthApplicationResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_oauth_application"
}

func (e *oauthApplicationResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Manages a Contentful OAuth application. OAuth applications are owned by the " +
			"user whose management token is used by the provider (Contentful's API only exposes " +
			"OAuth apps on the `/users/me` endpoint, even though the UI groups them under an " +
			"organization). The `client_secret` is only returned by Contentful at creation time " +
			"and cannot be retrieved later. The provider stores it in Terraform state; if the " +
			"state is lost or the resource is imported, the secret cannot be recovered and the " +
			"application must be replaced.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "OAuth application ID (sys.id).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Human-readable name of the OAuth application.",
			},
			"description": schema.StringAttribute{
				Required:    true,
				Description: "Description of the OAuth application.",
			},
			"scopes": schema.ListAttribute{
				Required:    true,
				ElementType: types.StringType,
				Description: "Scopes granted to access tokens issued for this application. " +
					"Valid values: `content_management_manage`, `content_management_read`.",
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
					listvalidator.ValueStringsAre(
						stringvalidator.OneOf("content_management_manage", "content_management_read"),
					),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"redirect_uri": schema.StringAttribute{
				Required:    true,
				Description: "Redirect URI used in the OAuth authorization code flow.",
			},
			"confidential": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Description: "Whether the client is confidential (server-side, can keep the " +
					"`client_secret`) or public (e.g. SPA/mobile/CLI). Defaults to `true`.",
				PlanModifiers: []planmodifier.Bool{
					custommodifier.BoolDefault(true),
				},
			},
			"client_id": schema.StringAttribute{
				Computed:    true,
				Description: "Public client identifier issued by Contentful.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"client_secret": schema.StringAttribute{
				Computed:    true,
				Sensitive:   true,
				Description: "Client secret. Returned by Contentful only at creation time; cannot be re-read.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (e *oauthApplicationResource) Configure(_ context.Context, request resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	data := request.ProviderData.(utils.ProviderData)
	e.client = data.Client
}

func (e *oauthApplicationResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan OAuthApplication
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	draft := plan.Draft()

	resp, err := e.client.CreateOAuthApplicationWithResponse(ctx, *draft)
	if err := utils.CheckClientResponse(resp, err, http.StatusCreated); err != nil {
		response.Diagnostics.AddError(
			"Error creating OAuth application",
			"Could not create OAuth application: "+err.Error(),
		)
		return
	}

	plan.Import(resp.JSON201)

	// client_secret is only returned here — capture it now. If Contentful ever
	// stops returning it on create, fail loudly rather than silently storing "".
	if resp.JSON201.ClientSecret == nil {
		response.Diagnostics.AddError(
			"Missing client_secret in create response",
			"Contentful did not return a client_secret on OAuth application creation; cannot continue.",
		)
		return
	}
	plan.ClientSecret = types.StringValue(*resp.JSON201.ClientSecret)

	response.Diagnostics.Append(response.State.Set(ctx, plan)...)
}

func (e *oauthApplicationResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state OAuthApplication
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	e.doRead(ctx, &state, &response.State, &response.Diagnostics)
}

func (e *oauthApplicationResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan OAuthApplication
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	var state OAuthApplication
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	draft := plan.Draft()

	resp, err := e.client.UpdateOAuthApplicationWithResponse(ctx, state.ID.ValueString(), *draft)
	if err := utils.CheckClientResponse(resp, err, http.StatusOK); err != nil {
		response.Diagnostics.AddError(
			"Error updating OAuth application",
			"Could not update OAuth application: "+err.Error(),
		)
		return
	}

	plan.Import(resp.JSON200)
	// Preserve client_secret from prior state — the API never returns it again.
	plan.ClientSecret = state.ClientSecret

	response.Diagnostics.Append(response.State.Set(ctx, plan)...)
}

func (e *oauthApplicationResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state OAuthApplication
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, err := e.client.DeleteOAuthApplicationWithResponse(ctx, state.ID.ValueString())
	if err != nil {
		response.Diagnostics.AddError(
			"Error deleting OAuth application",
			"Could not delete OAuth application: "+err.Error(),
		)
		return
	}

	// Treat 204 and 404 as success (idempotent delete).
	if resp.StatusCode() != http.StatusNoContent && resp.StatusCode() != http.StatusNotFound {
		response.Diagnostics.AddError(
			"Error deleting OAuth application",
			fmt.Sprintf("Unexpected response from Contentful API: %d", resp.StatusCode()),
		)
		return
	}
}

func (e *oauthApplicationResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	app := &OAuthApplication{
		ID: types.StringValue(request.ID),
		// client_secret cannot be retrieved from the API; surface it as null.
		// Subsequent applies will keep it null unless the resource is recreated.
		ClientSecret: types.StringNull(),
	}

	e.doRead(ctx, app, &response.State, &response.Diagnostics)
}

func (e *oauthApplicationResource) doRead(ctx context.Context, app *OAuthApplication, state *tfsdk.State, d *diag.Diagnostics) {
	resp, err := e.client.GetOAuthApplicationWithResponse(ctx, app.ID.ValueString())
	if err != nil {
		d.AddError(
			"Error reading OAuth application",
			"Could not read OAuth application: "+err.Error(),
		)
		return
	}

	if resp.StatusCode() == http.StatusNotFound {
		d.AddWarning(
			"OAuth application not found",
			fmt.Sprintf("OAuth application %s was not found, removing from state", app.ID.ValueString()),
		)
		return
	}

	if err := utils.CheckClientResponse(resp, err, http.StatusOK); err != nil {
		d.AddError(
			"Error reading OAuth application",
			"Could not read OAuth application: "+err.Error(),
		)
		return
	}

	// Hold on to client_secret across reads — the API doesn't return it.
	previousSecret := app.ClientSecret
	app.Import(resp.JSON200)
	app.ClientSecret = previousSecret

	d.Append(state.Set(ctx, app)...)
}

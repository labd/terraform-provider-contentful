package app_definition

import (
	"context"
	_ "embed"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/labd/contentful-go"
	"github.com/labd/terraform-provider-contentful/internal/custommodifier"
	"github.com/labd/terraform-provider-contentful/internal/customvalidator"
	"github.com/labd/terraform-provider-contentful/internal/utils"
)

//go:embed bundle.zip
var defaultDummyBundle []byte

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &appDefinitionResource{}
	_ resource.ResourceWithConfigure   = &appDefinitionResource{}
	_ resource.ResourceWithImportState = &appDefinitionResource{}
)

func NewAppDefinitionResource() resource.Resource {
	return &appDefinitionResource{}
}

// appDefinitionResource is the resource implementation.
type appDefinitionResource struct {
	client         *contentful.Client
	organizationId string
}

func (e *appDefinitionResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_app_definition"
}

func (e *appDefinitionResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Todo for explaining app definition",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "app definition id",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"src": schema.StringAttribute{
				Optional: true,
			},
			"use_bundle": schema.BoolAttribute{
				Required: true,
				PlanModifiers: []planmodifier.Bool{
					custommodifier.BoolDefault(false),
				},
			},
			"bundle_id": schema.StringAttribute{
				Computed: true,
			},
			"locations": schema.ListNestedAttribute{
				Validators: []validator.List{
					customvalidator.WhenOtherValueExistListValidator(path.MatchRelative().AtParent().AtName("src"), listvalidator.SizeAtLeast(1)),
				},
				Required: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"location": schema.StringAttribute{
							Required: true,
							Validators: []validator.String{
								stringvalidator.OneOf("entry-field", "entry-sidebar", "entry-editor", "dialog", "app-config", "page", "home"),
								customvalidator.AttributeNeedsToBeSetValidator(path.MatchRelative().AtParent().AtName("field_types"), "entry-field"),
							},
						},
						"navigation_item": schema.SingleNestedAttribute{
							Optional: true,
							// only be valid when type = page
							Attributes: map[string]schema.Attribute{
								"name": schema.StringAttribute{
									Required: true,
								},
								"path": schema.StringAttribute{
									Required: true,
								},
							},
						},
						// needs to be set when location is entry-field
						"field_types": schema.ListNestedAttribute{
							Optional: true,
							Validators: []validator.List{
								listvalidator.SizeAtLeast(1),
							},
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"type": schema.StringAttribute{
										Required: true,
										Validators: []validator.String{
											stringvalidator.OneOf(utils.GetAppFieldTypes()...),
											customvalidator.AttributeNeedsToBeSetValidator(path.MatchRelative().AtParent().AtName("link_type"), "Link"),
											customvalidator.AttributeNeedsToBeSetValidator(path.MatchRelative().AtParent().AtName("items"), "Array"),
										},
									},
									"link_type": schema.StringAttribute{
										Optional: true,
										// needs to be set when type is link
										Validators: []validator.String{
											customvalidator.StringAllowedWhenSetValidator(path.MatchRelative().AtParent().AtName("type"), "Link"),
											stringvalidator.OneOf(utils.GetLinkTypes()...),
										},
									},
									"items": schema.SingleNestedAttribute{
										Optional: true,
										//needs to be set when type is Array
										Attributes: map[string]schema.Attribute{
											"type": schema.StringAttribute{
												Required: true,
												Validators: []validator.String{
													stringvalidator.OneOf("Symbol", "Link"),
													customvalidator.AttributeNeedsToBeSetValidator(path.MatchRelative().AtParent().AtName("link_type"), "Link"),
												},
											},
											"link_type": schema.StringAttribute{
												Optional: true,
												// needs to be set when type is link
												Validators: []validator.String{
													stringvalidator.OneOf(utils.GetLinkTypes()...),
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func (e *appDefinitionResource) Configure(_ context.Context, request resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	data := request.ProviderData.(utils.ProviderData)
	e.client = data.Client
	e.organizationId = data.OrganizationId

}

func (e *appDefinitionResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan AppDefinition
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	if e.setDefaultBundle(&plan, response.Diagnostics) {
		return
	}

	draft := plan.Draft()

	if err := e.client.AppDefinitions.Upsert(e.organizationId, draft); err != nil {
		response.Diagnostics.AddError("Error creating app_definition definition", err.Error())
		return
	}

	plan.ID = types.StringValue(draft.Sys.ID)

	// Set state to fully populated data
	response.Diagnostics.Append(response.State.Set(ctx, plan)...)
	if response.Diagnostics.HasError() {
		return
	}
}

func (e *appDefinitionResource) setDefaultBundle(plan *AppDefinition, diagnostics diag.Diagnostics) bool {
	if plan.UseBundle.ValueBool() && (plan.BundleId.IsNull() || plan.BundleId.IsUnknown()) {

		draft := plan.Draft()

		locations := draft.Locations

		draft.Locations = []contentful.Locations{}

		if err := e.client.AppDefinitions.Upsert(e.organizationId, draft); err != nil {
			diagnostics.AddError("Error upsert app_definition definition", err.Error())
			return true
		}

		upload, err := e.client.AppUpload.Create(e.organizationId, defaultDummyBundle)

		if err != nil {
			diagnostics.AddError("Error uploading default bundle", err.Error())
			return true
		}

		draft.Locations = locations

		bundle, err := e.client.AppBundle.Create(e.organizationId, draft.Sys.ID, "Default Terraform BundleId", upload.Sys.ID)
		if err != nil {
			diagnostics.AddError("Error creating app_bundle for app definition", err.Error())
			return true
		}

		plan.ID = types.StringValue(draft.Sys.ID)
		plan.BundleId = types.StringValue(bundle.Sys.ID)
	}
	return false
}

func (e *appDefinitionResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	// Get current state
	var state *AppDefinition
	diags := request.State.Get(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	e.doRead(ctx, state, &response.State, &response.Diagnostics)
}

func (e *appDefinitionResource) doRead(ctx context.Context, app *AppDefinition, state *tfsdk.State, d *diag.Diagnostics) {

	appDefinition, err := e.getApp(app)
	if err != nil {
		d.AddError(
			"Error reading app definition",
			"Could not retrieve app definition, unexpected error: "+err.Error(),
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

func (e *appDefinitionResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan *AppDefinition
	diags := request.Plan.Get(ctx, &plan)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	// Get current state
	var state *AppDefinition
	diags = request.State.Get(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	appDefinition, err := e.getApp(plan)
	if err != nil {
		response.Diagnostics.AddError(
			"Error reading app definition",
			"Could not retrieve app definition, unexpected error: "+err.Error(),
		)
		return
	}

	if !plan.Equal(appDefinition) {

		if e.setDefaultBundle(plan, response.Diagnostics) {
			return
		}

		draft := plan.Draft()

		if err = e.client.AppDefinitions.Upsert(e.organizationId, draft); err != nil {
			response.Diagnostics.AddError(
				"Error updating app definition",
				"Could not update app definition, unexpected error: "+err.Error(),
			)
			return
		}
	}

	e.doRead(ctx, plan, &response.State, &response.Diagnostics)
}

func (e *appDefinitionResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	// Get current state
	var state *AppDefinition
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)

	if err := e.client.AppDefinitions.Delete(e.organizationId, state.ID.ValueString()); err != nil {
		response.Diagnostics.AddError(
			"Error deleting app_definition definition",
			"Could not delete app_definition definition, unexpected error: "+err.Error(),
		)
		return
	}

}

func (e *appDefinitionResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	futureState := &AppDefinition{
		ID: types.StringValue(request.ID),
	}

	e.doRead(ctx, futureState, &response.State, &response.Diagnostics)
}

func (e *appDefinitionResource) getApp(app *AppDefinition) (*contentful.AppDefinition, error) {
	return e.client.AppDefinitions.Get(e.organizationId, app.ID.ValueString())
}

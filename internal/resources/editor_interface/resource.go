package editor_interface

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/labd/terraform-provider-contentful/internal/custommodifier"
	"github.com/labd/terraform-provider-contentful/internal/customvalidator"
	"github.com/labd/terraform-provider-contentful/internal/sdk"
	"github.com/labd/terraform-provider-contentful/internal/utils"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &editorInterfaceResource{}
	_ resource.ResourceWithConfigure   = &editorInterfaceResource{}
	_ resource.ResourceWithImportState = &editorInterfaceResource{}
)

func NewEditorInterfaceResource() resource.Resource {
	return &editorInterfaceResource{}
}

// editorInterfaceResource is the resource implementation.
type editorInterfaceResource struct {
	client *sdk.ClientWithResponses
}

func (e *editorInterfaceResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_editor_interface"
}

func (e *editorInterfaceResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	widgetIdPath := path.MatchRelative().AtParent().AtParent().AtName("widget_id")

	response.Schema = schema.Schema{
		Description: "Contentful Editor Interface customizes the appearance and behavior of field editing controls in the Contentful web app.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Editor Interface ID (combination of content type, space, and environment)",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
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
			"content_type": schema.StringAttribute{
				Required:    true,
				Description: "Content Type ID that this editor interface applies to",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"version": schema.Int64Attribute{
				Computed:    true,
				Description: "The current version of the editor interface",
			},
			"controls": schema.ListNestedAttribute{
				Required:    true,
				Description: "The controls for the editor interface",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"field_id": schema.StringAttribute{
							Required:    true,
							Description: "The ID of the field",
						},
						"widget_id": schema.StringAttribute{
							Required:    true,
							Description: "The ID of the widget to use",
						},
						"widget_namespace": schema.StringAttribute{
							Optional:    true,
							Description: "Namespace of the widget",
							Computed:    true,
						},
						"settings": schema.SingleNestedAttribute{
							Attributes: map[string]schema.Attribute{
								"help_text": schema.StringAttribute{
									Optional: true,
								},
								"true_label": schema.StringAttribute{
									Optional: true,
									Validators: []validator.String{
										customvalidator.StringAllowedWhenSetValidator(widgetIdPath, "boolean"),
									},
								},
								"false_label": schema.StringAttribute{
									Optional: true,
									Validators: []validator.String{
										customvalidator.StringAllowedWhenSetValidator(widgetIdPath, "boolean"),
									},
								},
								"stars": schema.Int64Attribute{
									Optional: true,
									Validators: []validator.Int64{
										customvalidator.Int64AllowedWhenSetValidator(widgetIdPath, "rating"),
									},
								},
								"format": schema.StringAttribute{
									Optional: true,
									Validators: []validator.String{
										stringvalidator.OneOf("dateonly", "time", "timeZ"),
										customvalidator.StringAllowedWhenSetValidator(widgetIdPath, "datePicker"),
									},
								},
								"ampm": schema.StringAttribute{
									Optional: true,
									Validators: []validator.String{
										stringvalidator.OneOf("12", "24"),
										customvalidator.StringAllowedWhenSetValidator(widgetIdPath, "datePicker"),
									},
								},
								/** (only for References, many) Select whether to enable Bulk Editing mode */
								"bulk_editing": schema.BoolAttribute{
									Optional: true,
								},
								"tracking_field_id": schema.StringAttribute{
									Optional: true,
									Validators: []validator.String{
										customvalidator.StringAllowedWhenSetValidator(widgetIdPath, "slugEditor"),
									},
								},
							},
							Optional: true,
							PlanModifiers: []planmodifier.Object{
								objectplanmodifier.UseStateForUnknown(),
							},
						},
					},
				},
			},
			"sidebar": schema.ListNestedAttribute{
				Optional: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"disabled": schema.BoolAttribute{
							Optional: true,
							Computed: true,
							Default:  booldefault.StaticBool(false),
						},
						"widget_id": schema.StringAttribute{
							Required: true,
						},
						"widget_namespace": schema.StringAttribute{
							Required: true,
						},
						"settings": schema.StringAttribute{
							CustomType: jsontypes.NormalizedType{},
							Optional:   true,
							Computed:   true,
							PlanModifiers: []planmodifier.String{
								custommodifier.StringDefault("{}"),
							},
						},
					},
				},
			},
			"editors": schema.ListNestedAttribute{
				Optional: true,
				Description: "You can add or replace the default entry editor with a custom editor (App or UI " +
					"Extension) by configuring the optional editors property, which allows passing instance " +
					"parameters and disabling the default editor if desired.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"disabled": schema.BoolAttribute{
							Optional: true,
							Computed: true,
							Default:  booldefault.StaticBool(false),
						},
						"widget_id": schema.StringAttribute{
							Required: true,
						},
						"widget_namespace": schema.StringAttribute{
							Required: true,
						},
						"settings": schema.StringAttribute{
							CustomType: jsontypes.NormalizedType{},
							Optional:   true,
							Computed:   true,
							PlanModifiers: []planmodifier.String{
								custommodifier.StringDefault("{}"),
							},
						},
					},
				},
				Validators: []validator.List{
					listvalidator.SizeBetween(1, 10),
				},
			},
		},
	}
}

func (e *editorInterfaceResource) Configure(_ context.Context, request resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	data := request.ProviderData.(utils.ProviderData)
	e.client = data.Client
}

func (e *editorInterfaceResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	// Get plan values
	var plan EditorInterface
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Check if the editor interface exists by fetching it
	resp, err := e.client.GetEditorInterfaceWithResponse(
		ctx,
		plan.SpaceID.ValueString(),
		plan.Environment.ValueString(),
		plan.ContentType.ValueString(),
	)
	if err := utils.CheckClientResponse(resp, err, http.StatusOK); err != nil {
		response.Diagnostics.AddError(
			"Error fetching editor interface",
			"Could not fetch editor interface: "+err.Error(),
		)
		return
	}

	// Prepare update data
	updateBody := plan.ToUpdateBody()

	// Update the editor interface
	updateParams := &sdk.UpdateEditorInterfaceParams{
		XContentfulVersion: int64(resp.JSON200.Sys.Version),
	}

	updateResp, err := e.client.UpdateEditorInterfaceWithResponse(
		ctx,
		plan.SpaceID.ValueString(),
		plan.Environment.ValueString(),
		plan.ContentType.ValueString(),
		updateParams,
		updateBody,
	)
	if err := utils.CheckClientResponse(updateResp, err, http.StatusOK); err != nil {
		response.Diagnostics.AddError(
			"Error updating editor interface",
			"Could not update editor interface: "+err.Error(),
		)
		return
	}

	state := EditorInterface{}
	state.Import(updateResp.JSON200)

	response.Diagnostics.Append(response.State.Set(ctx, state)...)
}

func (e *editorInterfaceResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	// Get current state
	var state EditorInterface
	diags := request.State.Get(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, err := e.client.GetEditorInterfaceWithResponse(
		ctx,
		state.SpaceID.ValueString(),
		state.Environment.ValueString(),
		state.ContentType.ValueString(),
	)
	if err := utils.CheckClientResponse(resp, err, http.StatusOK); err != nil {
		if resp.StatusCode() == 404 {
			request.State.RemoveResource(ctx)
			return
		}

		response.Diagnostics.AddError(
			"Error importing editor interface",
			fmt.Sprintf("Could not import editor interface: %s", err.Error()),
		)
		return
	}

	state.Import(resp.JSON200)
	response.Diagnostics.Append(response.State.Set(ctx, state)...)
}

func (e *editorInterfaceResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	// Get plan values
	var plan EditorInterface
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Get current state
	var state EditorInterface
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Retrieve latest version of the editor interface. It seems that it might
	// have been updated if the content type was updated in the meantime.
	version, err := e.GetCurrentVersion(ctx, plan, response)
	if err != nil {
		response.Diagnostics.AddError(
			"Error fetching editor interface version",
			"Could not fetch editor interface version: "+err.Error(),
		)
		return
	}
	state.Version = types.Int64Value(version)

	// Prepare update data
	updateBody := plan.ToUpdateBody()

	// Update the editor interface
	updateParams := &sdk.UpdateEditorInterfaceParams{
		XContentfulVersion: state.Version.ValueInt64(),
	}

	updateResp, err := e.client.UpdateEditorInterfaceWithResponse(
		ctx,
		plan.SpaceID.ValueString(),
		plan.Environment.ValueString(),
		plan.ContentType.ValueString(),
		updateParams,
		updateBody,
	)
	if err := utils.CheckClientResponse(updateResp, err, http.StatusOK); err != nil {
		response.Diagnostics.AddError(
			"Error updating editor interface",
			"Could not update editor interface: "+err.Error(),
		)
		return
	}

	state.Import(updateResp.JSON200)
	response.Diagnostics.Append(response.State.Set(ctx, state)...)
}

func (e *editorInterfaceResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	// Editor interfaces cannot be truly deleted, only reset to defaults
	// For now, we'll just let it be removed from state without any API call
	tflog.Info(ctx, "Editor interface removed from state (but not deleted in Contentful)")
}

func (e *editorInterfaceResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	// Extract the space ID, environment ID, and content type ID from the import ID
	idParts := strings.Split(request.ID, ":")
	if len(idParts) != 3 || idParts[0] == "" || idParts[1] == "" || idParts[2] == "" {
		response.Diagnostics.AddError(
			"Error importing editor interface",
			fmt.Sprintf("Expected import format: space_id:environment:content_type_id, got: %s", request.ID),
		)
		return
	}

	spaceID := idParts[0]
	environment := idParts[1]
	contentTypeID := idParts[2]

	resp, err := e.client.GetEditorInterfaceWithResponse(
		ctx, spaceID, environment, contentTypeID)

	if err := utils.CheckClientResponse(resp, err, http.StatusOK); err != nil {
		response.Diagnostics.AddError(
			"Error importing editor interface",
			fmt.Sprintf("Could not import editor interface: %s", err.Error()),
		)
		return
	}

	state := EditorInterface{}
	state.Import(resp.JSON200)
	response.Diagnostics.Append(response.State.Set(ctx, state)...)
}

func (e *editorInterfaceResource) GetCurrentVersion(ctx context.Context, plan EditorInterface, response *resource.UpdateResponse) (int64, error) {
	getResp, err := e.client.GetEditorInterfaceWithResponse(
		ctx,
		plan.SpaceID.ValueString(),
		plan.Environment.ValueString(),
		plan.ContentType.ValueString(),
	)
	if err := utils.CheckClientResponse(getResp, err, http.StatusOK); err != nil {
		return 0, fmt.Errorf("Error fetching editor interface: %s", err.Error())
	}
	return getResp.JSON200.Sys.Version, nil
}

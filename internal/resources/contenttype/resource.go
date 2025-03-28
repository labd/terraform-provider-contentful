package contenttype

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/davecgh/go-spew/spew"
	"github.com/elliotchance/pie/v2"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/labd/terraform-provider-contentful/internal/custommodifier"
	"github.com/labd/terraform-provider-contentful/internal/customvalidator"
	"github.com/labd/terraform-provider-contentful/internal/sdk"
	"github.com/labd/terraform-provider-contentful/internal/utils"
)

type key int

const (
	OnlyControlVersion key = iota
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &contentTypeResource{}
	_ resource.ResourceWithConfigure   = &contentTypeResource{}
	_ resource.ResourceWithImportState = &contentTypeResource{}
)

func NewContentTypeResource() resource.Resource {
	return &contentTypeResource{}
}

// contentTypeResource is the resource implementation.
type contentTypeResource struct {
	client *sdk.ClientWithResponses
}

func (e *contentTypeResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_contenttype"
}

var resourceLinkTypes = []string{"Contentful:Entry"}
var arrayItemTypes = []string{"Symbol", "Link", "ResourceLink"}

//https://www.contentful.com/developers/docs/extensibility/app-framework/editor-interfaces/

func (e *contentTypeResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {

	sizeSchema := schema.SingleNestedAttribute{
		Optional: true,
		Attributes: map[string]schema.Attribute{
			"min": schema.Float64Attribute{
				Optional: true,
			},
			"max": schema.Float64Attribute{
				Optional: true,
			},
		},
	}

	validationsSchema := schema.ListNestedAttribute{
		Optional: true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"range": sizeSchema,
				"size":  sizeSchema,
				"unique": schema.BoolAttribute{
					Optional: true,
				},
				"asset_file_size": sizeSchema,
				"regexp": schema.SingleNestedAttribute{
					Optional: true,
					Attributes: map[string]schema.Attribute{
						"pattern": schema.StringAttribute{
							Optional: true,
						},
					},
				},
				"link_mimetype_group": schema.ListAttribute{
					Optional:    true,
					ElementType: types.StringType,
				},
				"link_content_type": schema.ListAttribute{
					Optional:    true,
					ElementType: types.StringType,
				},
				"in": schema.ListAttribute{
					Optional:    true,
					ElementType: types.StringType,
				},
				"enabled_marks": schema.ListAttribute{
					Optional:    true,
					ElementType: types.StringType,
				},
				"enabled_node_types": schema.ListAttribute{
					Optional:    true,
					ElementType: types.StringType,
				},
				"message": schema.StringAttribute{
					Optional: true,
				},
			},
			Validators: []validator.Object{
				objectvalidator.AtLeastOneOf(path.MatchRelative().AtName("range")),
			},
		},
	}

	widgetIdPath := path.MatchRelative().AtParent().AtParent().AtName("widget_id")

	response.Schema = schema.Schema{
		Description: "Todo for explaining contenttype",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "content type id",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"version": schema.Int64Attribute{
				Computed: true,
			},
			"version_controls": schema.Int64Attribute{
				Computed: true,
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
			"name": schema.StringAttribute{
				Required: true,
			},
			"display_field": schema.StringAttribute{
				Required: true,
			},
			"description": schema.StringAttribute{
				Optional: true,
			},
			"manage_field_controls": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					custommodifier.BoolDefault(false),
				},
			},
			"fields": schema.ListNestedAttribute{
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
					listvalidator.UniqueValues(),
				},
				Required: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Required: true,
						},
						"name": schema.StringAttribute{
							Required: true,
						},
						"type": schema.StringAttribute{
							Required: true,
							Validators: []validator.String{
								stringvalidator.OneOf(utils.GetContentTypes()...),
								customvalidator.AttributeNeedsToBeSetValidator(path.MatchRelative().AtParent().AtName("link_type"), "Link"),
								customvalidator.AttributeNeedsToBeSetValidator(path.MatchRelative().AtParent().AtName("items"), "Array"),
							},
						},
						"link_type": schema.StringAttribute{
							Optional: true,
							Validators: []validator.String{
								stringvalidator.OneOf(utils.GetLinkTypes()...),
							},
						},
						"required": schema.BoolAttribute{
							Optional: true,
							Computed: true,
							PlanModifiers: []planmodifier.Bool{
								custommodifier.BoolDefault(false),
							},
						},
						"localized": schema.BoolAttribute{
							Optional: true,
							Computed: true,
							PlanModifiers: []planmodifier.Bool{
								custommodifier.BoolDefault(false),
							},
						},
						"disabled": schema.BoolAttribute{
							Optional: true,
							Computed: true,
							PlanModifiers: []planmodifier.Bool{
								custommodifier.BoolDefault(false),
							},
						},
						"omitted": schema.BoolAttribute{
							Optional: true,
							Computed: true,
							PlanModifiers: []planmodifier.Bool{
								custommodifier.BoolDefault(false),
							},
						},
						"validations": validationsSchema,
						"items": schema.SingleNestedAttribute{
							Optional: true,
							PlanModifiers: []planmodifier.Object{
								objectplanmodifier.UseStateForUnknown(),
							},
							Attributes: map[string]schema.Attribute{
								"type": schema.StringAttribute{
									Required: true,
									Validators: []validator.String{
										stringvalidator.OneOf(arrayItemTypes...),
										customvalidator.AttributeNeedsToBeSetValidator(path.MatchRelative().AtParent().AtName("link_type"), "Link"),
									},
								},
								"link_type": schema.StringAttribute{
									Optional: true,
									Validators: []validator.String{
										stringvalidator.OneOf(utils.GetLinkTypes()...),
									},
								},
								"validations": validationsSchema,
							},
						},
						"default_value": schema.SingleNestedAttribute{
							Optional: true,
							Attributes: map[string]schema.Attribute{
								"bool": schema.MapAttribute{
									ElementType: types.BoolType,
									Optional:    true,
								},
								"string": schema.MapAttribute{
									ElementType: types.StringType,
									Optional:    true,
								},
							},
						},
						"control": schema.SingleNestedAttribute{
							Optional: true,
							PlanModifiers: []planmodifier.Object{
								objectplanmodifier.UseStateForUnknown(),
							},

							Attributes: map[string]schema.Attribute{
								"widget_id": schema.StringAttribute{
									Required: true,
								},
								"widget_namespace": schema.StringAttribute{
									Required: true,
									Validators: []validator.String{
										stringvalidator.OneOf("builtin", "extension", "app"),
									},
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
												customvalidator.StringAllowedWhenSetValidator(widgetIdPath, "datepicker"),
											},
										},
										"ampm": schema.StringAttribute{
											Optional: true,
											Validators: []validator.String{
												stringvalidator.OneOf("12", "24"),
												customvalidator.StringAllowedWhenSetValidator(widgetIdPath, "datepicker"),
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
					PlanModifiers: []planmodifier.Object{
						custommodifier.FieldTypeChangeProhibited(),
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
							PlanModifiers: []planmodifier.Bool{
								custommodifier.BoolDefault(false),
							},
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
		},
	}
}

func (e *contentTypeResource) Configure(_ context.Context, request resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	data := request.ProviderData.(utils.ProviderData)
	e.client = data.Client

}

func (e *contentTypeResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan ContentType
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	spaceId := plan.SpaceId.ValueString()
	environment := plan.Environment.ValueString()

	var contentType *sdk.ContentType

	if !plan.ID.IsUnknown() && !plan.ID.IsNull() {
		draft, err := plan.Update()
		if err != nil {
			response.Diagnostics.AddError("Error creating contenttype", err.Error())
			return
		}
		resp, err := e.client.UpdateContentTypeWithResponse(ctx, spaceId, environment, plan.ID.ValueString(), nil, *draft)
		if err != nil {
			response.Diagnostics.AddError("Error creating contenttype", "Could not create contenttype with id "+plan.ID.ValueString()+", unexpected error: "+err.Error())
			return
		}

		if resp.StatusCode() != 201 {
			response.Diagnostics.AddError("Error creating contenttype", "Could not create contenttype with id "+plan.ID.ValueString()+", unexpected status code: "+resp.Status())
			return
		}
		contentType = resp.JSON201
	} else {
		draft, err := plan.Update()
		if err != nil {
			response.Diagnostics.AddError("Error creating contenttype", err.Error())
			return
		}
		resp, err := e.client.UpdateContentTypeWithResponse(ctx, spaceId, environment, plan.Name.ValueString(), nil, *draft)
		if err != nil {
			response.Diagnostics.AddError("Error creating contenttype", "Could not create contenttype, unexpected error: "+err.Error())
			return
		}
		if resp.StatusCode() != 201 {
			response.Diagnostics.AddError("Error creating contenttype", "Could not create contenttype, unexpected status code: "+resp.Status())
			return
		}
		contentType = resp.JSON201
	}

	contentType, err := e.activateContentType(ctx, spaceId, environment, contentType.Sys.Id, contentType.Sys.Version)
	if err != nil {
		response.Diagnostics.AddError("Error creating contenttype", "Could not activate contenttype, unexpected error: "+err.Error())
		return
	}

	tflog.Debug(ctx, spew.Sdump(contentType))

	plan.ID = types.StringValue(contentType.Sys.Id)
	plan.Version = types.Int64Value(contentType.Sys.Version)
	plan.VersionControls = types.Int64Value(0)

	// Set state to fully populated data
	response.Diagnostics.Append(response.State.Set(ctx, plan)...)
	if response.Diagnostics.HasError() {
		return
	}
}

func (e *contentTypeResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	// Get current state
	var state *ContentType
	diags := request.State.Get(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	e.doRead(ctx, state, &response.State, &response.Diagnostics)
}

func (e *contentTypeResource) doRead(ctx context.Context, contentType *ContentType, state *tfsdk.State, d *diag.Diagnostics) {

	contentfulContentType, err := e.getContentType(ctx, contentType)
	if err != nil {
		d.AddError(
			"Error reading contenttype",
			"Could not retrieve contenttype, unexpected error: "+err.Error(),
		)
		return
	}

	var editorInterface *sdk.EditorInterface

	if contentType.ManageFieldControls.ValueBool() {
		editorInterface, err = e.getEditorInterface(ctx, contentType)
		if err != nil {
			d.AddError(
				"Error reading contenttype",
				"Could not retrieve contenttype, unexpected error: "+err.Error(),
			)
			return
		}

		if u, ok := ctx.Value(OnlyControlVersion).(bool); ok && u {
			contentType.VersionControls = types.Int64Value(*editorInterface.Sys.Version)

			ei := &sdk.EditorInterface{}
			contentType.DraftEditorInterface(ei)

			// remove all controls which are not in the plan for an easier import
			editorInterface.Controls = pie.Filter(editorInterface.Controls, func(c sdk.EditorInterfaceControl) bool {
				return pie.Any(ei.Controls, func(value sdk.EditorInterfaceControl) bool {
					return value.FieldId == c.FieldId && value.WidgetId != nil && reflect.DeepEqual(value.WidgetId, c.WidgetId)
				})
			})
		}
	}

	err = contentType.Import(contentfulContentType, editorInterface)
	tflog.Debug(ctx, spew.Sdump(editorInterface))

	if err != nil {
		d.AddError(
			"Error importing contenttype to state",
			"Could not import contenttype to state, unexpected error: "+err.Error(),
		)
		return
	}

	// Set refreshed state
	d.Append(state.Set(ctx, &contentType)...)
	if d.HasError() {
		return
	}
}

func (e *contentTypeResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan *ContentType
	diags := request.Plan.Get(ctx, &plan)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	// Get current state
	var state *ContentType
	diags = request.State.Get(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	contentfulContentType, err := e.getContentType(ctx, plan)
	if err != nil {
		response.Diagnostics.AddError(
			"Error reading contenttype",
			"Could not retrieve contenttype, unexpected error: "+err.Error(),
		)
		return
	}
	plan.Version = types.Int64Value(contentfulContentType.Sys.Version)

	// Mark the fields as omitted that are no longer in the plan
	deletedFields := pie.Of(pie.FilterNot(contentfulContentType.Fields, func(cf sdk.Field) bool {
		return pie.FindFirstUsing(plan.Fields, func(f Field) bool {
			return cf.Id == f.Id.ValueString()
		}) != -1

	})).Map(func(f sdk.Field) sdk.Field {
		f.Omitted = utils.Pointer(true)

		return f
	}).Result

	draft, err := plan.Update()

	if err != nil {
		response.Diagnostics.AddError("Error updating contenttype", err.Error())
		return
	}

	if len(deletedFields) > 0 {
		draft.Fields = append(draft.Fields, deletedFields...)
	}

	// To remove a field from a content type 4 API calls need to be made.
	// Omit the removed fields and publish the new version of the content type,
	// followed by the field removal and final publish.

	if !plan.Equal(contentfulContentType) {
		contentType, err := e.doUpdate(ctx, plan, draft)
		if err != nil {
			response.Diagnostics.AddError(
				"Error updating contenttype",
				"Could not update contenttype, unexpected error: "+err.Error(),
			)
			return
		}
		plan.Version = types.Int64Value(contentType.Sys.Version)

		// Now generate a new plan, to remove the fields that we previously marked
		// as omitted
		if len(deletedFields) > 0 {
			draft, err = plan.Update()

			if err != nil {
				response.Diagnostics.AddError("Error updating contenttype", err.Error())
				return
			}

			contentType, err = e.doUpdate(ctx, plan, draft)
			if err != nil {
				response.Diagnostics.AddError(
					"Error updating contenttype",
					"Could not update contenttype, unexpected error: "+err.Error(),
				)
				return
			}
			plan.Version = types.Int64Value(contentType.Sys.Version)
		}
	}

	ctxControls := e.updateEditorInterface(ctx, state, plan, &response.Diagnostics)

	e.doRead(ctxControls, plan, &response.State, &response.Diagnostics)
}

func (e *contentTypeResource) updateEditorInterface(ctx context.Context, state *ContentType, plan *ContentType, d *diag.Diagnostics) context.Context {
	if plan.ManageFieldControls.ValueBool() {
		// first import of editorInterface to the state just get the editorInterface version
		if state.VersionControls.IsNull() {
			return context.WithValue(ctx, OnlyControlVersion, true)
		}

		editorInterface, err := e.getEditorInterface(ctx, plan)
		if err != nil {
			d.AddError(
				"Error reading contenttype",
				"Could not retrieve contenttype editorInterface, unexpected error: "+err.Error(),
			)
			return ctx
		}

		if !plan.EqualEditorInterface(editorInterface) {

			plan.DraftEditorInterface(editorInterface)

			params := &sdk.UpdateEditorInterfaceParams{
				XContentfulVersion: *editorInterface.Sys.Version,
			}
			resp, err := e.client.UpdateEditorInterfaceWithResponse(
				ctx, plan.SpaceId.ValueString(), plan.Environment.ValueString(), plan.ID.ValueString(), params, *editorInterface)

			if err != nil {
				d.AddError(
					"Error updating contenttype editorInterface",
					"Could not update contenttype editorInterface, unexpected error: "+err.Error(),
				)

				return ctx
			}

			if resp.StatusCode() != 200 {
				d.AddError(
					"Error updating contenttype editorInterface",
					"Could not update contenttype editorInterface, unexpected status code: "+resp.Status(),
				)
				return ctx
			}
		}
	}

	return ctx
}

func (e *contentTypeResource) doUpdate(ctx context.Context, plan *ContentType, draft *sdk.ContentTypeUpdate) (*sdk.ContentType, error) {
	spaceId := plan.SpaceId.ValueString()
	environment := plan.Environment.ValueString()
	id := plan.ID.ValueString()

	params := &sdk.UpdateContentTypeParams{
		XContentfulVersion: plan.Version.ValueInt64(),
	}

	resp, err := e.client.UpdateContentTypeWithResponse(ctx, spaceId, environment, id, params, *draft)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("unexpected status code: %s", resp.Status())
	}

	contentType := resp.JSON200

	return e.activateContentType(ctx, spaceId, environment, id, contentType.Sys.Version)
}

func (e *contentTypeResource) activateContentType(ctx context.Context, spaceId, environment, id string, version int64) (*sdk.ContentType, error) {
	params := &sdk.ActivateContentTypeParams{
		XContentfulVersion: version,
	}
	resp, err := e.client.ActivateContentTypeWithResponse(ctx, spaceId, environment, id, params)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("unexpected status code: %s", resp.Status())
	}
	return resp.JSON200, nil
}

func (e *contentTypeResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	// Get current state
	var state *ContentType
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)

	spaceId := state.SpaceId.ValueString()
	environment := state.Environment.ValueString()
	id := state.ID.ValueString()

	resp, err := e.client.DeactivateContentTypeWithResponse(
		ctx,
		spaceId,
		environment,
		id,
		&sdk.DeactivateContentTypeParams{
			XContentfulVersion: state.Version.ValueInt64(),
		},
	)

	if err != nil {
		response.Diagnostics.AddError(
			"Error deleting contenttype",
			"Could not deactivate contenttype, unexpected error: "+err.Error(),
		)
	}

	if resp.StatusCode() != 200 {
		response.Diagnostics.AddError(
			"Error deleting contenttype",
			"Could not deactivate contenttype, unexpected status code: "+resp.Status(),
		)
	}

	version := resp.JSON200.Sys.Version

	respDelete, err := e.client.DeleteContentTypeWithResponse(
		ctx,
		spaceId,
		environment,
		id,
		&sdk.DeleteContentTypeParams{
			XContentfulVersion: version,
		},
	)

	if err != nil {
		response.Diagnostics.AddError(
			"Error deleting contenttype",
			"Could not delete contenttype, unexpected error: "+err.Error(),
		)
	}

	if respDelete.StatusCode() != 204 {
		response.Diagnostics.AddError(
			"Error deleting contenttype",
			"Could not delete contenttype, unexpected status code: "+resp.Status(),
		)
	}

}

func (e *contentTypeResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	idParts := strings.SplitN(request.ID, ":", 3)

	if len(idParts) != 3 || idParts[0] == "" || idParts[1] == "" || idParts[2] == "" {
		response.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: localeId:env:spaceId. Got: %q", request.ID),
		)
		return
	}

	futureState := &ContentType{
		ID:          types.StringValue(idParts[0]),
		SpaceId:     types.StringValue(idParts[2]),
		Environment: types.StringValue(idParts[1]),
	}

	e.doRead(ctx, futureState, &response.State, &response.Diagnostics)
}

func (e *contentTypeResource) getContentType(ctx context.Context, editor *ContentType) (*sdk.ContentType, error) {
	spaceId := editor.SpaceId.ValueString()
	environment := editor.Environment.ValueString()
	id := editor.ID.ValueString()

	tflog.Debug(ctx, fmt.Sprintf("spaceId: %s, environment: %s, id: %s", spaceId, environment, id))
	resp, err := e.client.GetContentTypeWithResponse(ctx, spaceId, environment, id)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("unexpected status code: %s", resp.Status())
	}

	return resp.JSON200, nil
}

func (e *contentTypeResource) getEditorInterface(ctx context.Context, editor *ContentType) (*sdk.EditorInterface, error) {
	spaceId := editor.SpaceId.ValueString()
	environment := editor.Environment.ValueString()

	resp, err := e.client.GetEditorInterfaceWithResponse(ctx, spaceId, environment, editor.ID.ValueString())
	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("unexpected status code: %s", resp.Status())
	}

	return resp.JSON200, nil
}

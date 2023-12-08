package contenttype

import (
	"context"
	"fmt"
	"github.com/elliotchance/pie/v2"
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
	"github.com/labd/contentful-go"
	"github.com/labd/terraform-provider-contentful/internal/custommodifier"
	"github.com/labd/terraform-provider-contentful/internal/customvalidator"
	"github.com/labd/terraform-provider-contentful/internal/utils"
	"reflect"
	"strings"
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
	client *contentful.Client
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
				Optional: true,
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

	draft, err := plan.Draft()

	if err != nil {
		response.Diagnostics.AddError("Error creating contenttype", err.Error())
		return
	}

	if plan.Environment.IsUnknown() || plan.Environment.IsNull() {
		if err = e.client.ContentTypes.Upsert(plan.SpaceId.ValueString(), draft); err != nil {
			response.Diagnostics.AddError("Error creating contenttype", err.Error())
			return
		}

		if err = e.client.ContentTypes.Activate(plan.SpaceId.ValueString(), draft); err != nil {
			response.Diagnostics.AddError("Error activating contenttype", err.Error())
			return
		}
	} else {

		env := &contentful.Environment{Sys: &contentful.Sys{
			ID: plan.Environment.ValueString(),
			Space: &contentful.Space{
				Sys: &contentful.Sys{ID: plan.SpaceId.ValueString()},
			},
		}}

		if err = e.client.ContentTypes.UpsertWithEnv(env, draft); err != nil {
			response.Diagnostics.AddError("Error creating contenttype", err.Error())
			return
		}

		if err = e.client.ContentTypes.ActivateWithEnv(env, draft); err != nil {
			response.Diagnostics.AddError("Error activating contenttype", err.Error())
			return
		}
	}

	plan.Version = types.Int64Value(int64(draft.Sys.Version))
	plan.VersionControls = types.Int64Value(0)
	plan.ID = types.StringValue(draft.Sys.ID)

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

	contentfulContentType, err := e.getContentType(contentType)
	if err != nil {
		d.AddError(
			"Error reading contenttype",
			"Could not retrieve contenttype, unexpected error: "+err.Error(),
		)
		return
	}

	var controls *contentful.EditorInterface

	if contentType.ManageFieldControls.ValueBool() {
		controls, err = e.getContentTypeControls(contentType)
		if err != nil {
			d.AddError(
				"Error reading contenttype",
				"Could not retrieve contenttype, unexpected error: "+err.Error(),
			)
			return
		}

		if u, ok := ctx.Value(OnlyControlVersion).(bool); ok && u {
			contentType.VersionControls = types.Int64Value(int64(controls.Sys.Version))

			draftControls := contentType.DraftControls()
			// remove all controls which are not in the plan for an easier import
			controls.Controls = pie.Filter(controls.Controls, func(c contentful.Controls) bool {
				return pie.Any(draftControls, func(value contentful.Controls) bool {
					return value.FieldID == c.FieldID && value.WidgetID != nil && reflect.DeepEqual(value.WidgetID, c.WidgetID)
				})
			})
		}
	}

	err = contentType.Import(contentfulContentType, controls)

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

	contentfulContentType, err := e.getContentType(plan)
	if err != nil {
		response.Diagnostics.AddError(
			"Error reading contenttype",
			"Could not retrieve contenttype, unexpected error: "+err.Error(),
		)
		return
	}

	deletedFields := pie.Of(pie.FilterNot(contentfulContentType.Fields, func(cf *contentful.Field) bool {
		return pie.FindFirstUsing(plan.Fields, func(f Field) bool {
			return cf.ID == f.Id.ValueString()
		}) != -1

	})).Map(func(f *contentful.Field) *contentful.Field {
		f.Omitted = true

		return f
	}).Result

	draft, err := plan.Draft()

	if err != nil {
		response.Diagnostics.AddError("Error updating contenttype", err.Error())
		return
	}

	draft.Sys = contentfulContentType.Sys

	if len(deletedFields) > 0 {
		draft.Fields = append(draft.Fields, deletedFields...)
	}

	// To remove a field from a content type 4 API calls need to be made.
	// Omit the removed fields and publish the new version of the content type,
	// followed by the field removal and final publish.

	if !plan.Equal(contentfulContentType) {
		err = e.doUpdate(plan, draft)
		if err != nil {
			response.Diagnostics.AddError(
				"Error updating subscription",
				"Could not update subscription, unexpected error: "+err.Error(),
			)
			return
		}

		if len(deletedFields) > 0 {
			sys := draft.Sys
			draft, err = plan.Draft()

			if err != nil {
				response.Diagnostics.AddError("Error updating contenttype", err.Error())
				return
			}

			draft.Sys = sys

			err = e.doUpdate(plan, draft)
			if err != nil {
				response.Diagnostics.AddError(
					"Error updating subscription",
					"Could not update subscription, unexpected error: "+err.Error(),
				)
				return
			}
		}
	}

	ctxControls := e.updateControls(ctx, state, plan, &response.Diagnostics)

	e.doRead(ctxControls, plan, &response.State, &response.Diagnostics)
}

func (e *contentTypeResource) updateControls(ctx context.Context, state *ContentType, plan *ContentType, d *diag.Diagnostics) context.Context {
	if plan.ManageFieldControls.ValueBool() {
		// first import of controls to the state just get the controls version
		if state.VersionControls.IsNull() {
			return context.WithValue(ctx, OnlyControlVersion, true)
		}

		controls, err := e.getContentTypeControls(plan)
		if err != nil {
			d.AddError(
				"Error reading contenttype",
				"Could not retrieve contenttype controls, unexpected error: "+err.Error(),
			)
			return ctx
		}

		if !plan.EqualControls(controls.Controls) {

			controls.Controls = plan.DraftControls()
			if !plan.Environment.IsUnknown() && !plan.Environment.IsNull() {
				e.client.SetEnvironment(plan.Environment.ValueString())
			}

			if err = e.client.EditorInterfaces.Update(plan.SpaceId.ValueString(), plan.ID.ValueString(), controls); err != nil {
				d.AddError(
					"Error updating contenttype controls",
					"Could not update contenttype controls, unexpected error: "+err.Error(),
				)

				return ctx
			}
		}

	}

	return ctx
}

func (e *contentTypeResource) doUpdate(plan *ContentType, draft *contentful.ContentType) error {
	if plan.Environment.IsUnknown() || plan.Environment.IsNull() {

		if err := e.client.ContentTypes.Upsert(plan.SpaceId.ValueString(), draft); err != nil {
			return err
		}

		if err := e.client.ContentTypes.Activate(plan.SpaceId.ValueString(), draft); err != nil {
			return err
		}
	} else {

		env := &contentful.Environment{Sys: &contentful.Sys{
			ID: plan.Environment.ValueString(),
			Space: &contentful.Space{
				Sys: &contentful.Sys{ID: plan.SpaceId.ValueString()},
			},
		}}

		if err := e.client.ContentTypes.UpsertWithEnv(env, draft); err != nil {
			return err
		}

		if err := e.client.ContentTypes.ActivateWithEnv(env, draft); err != nil {
			return err
		}
	}
	return nil
}

func (e *contentTypeResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	// Get current state
	var state *ContentType
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)

	contentfulContentType, err := e.getContentType(state)

	if err != nil {
		response.Diagnostics.AddError(
			"Error reading contenttype",
			"Could not retrieve contenttype, unexpected error: "+err.Error(),
		)
		return
	}

	err = e.doDelete(state, contentfulContentType)
	if err != nil {
		response.Diagnostics.AddError(
			"Error deleting contenttype",
			"Could not delete contenttype, unexpected error: "+err.Error(),
		)
		return
	}

}

func (e *contentTypeResource) doDelete(data *ContentType, draft *contentful.ContentType) error {
	if data.Environment.IsUnknown() || data.Environment.IsNull() {

		if err := e.client.ContentTypes.Deactivate(data.SpaceId.ValueString(), draft); err != nil {
			return err
		}

		if err := e.client.ContentTypes.Delete(data.SpaceId.ValueString(), draft); err != nil {
			return err
		}
	} else {

		env := &contentful.Environment{Sys: &contentful.Sys{
			ID: data.Environment.ValueString(),
			Space: &contentful.Space{
				Sys: &contentful.Sys{ID: data.SpaceId.ValueString()},
			},
		}}

		if err := e.client.ContentTypes.DeactivateWithEnv(env, draft); err != nil {
			return err
		}

		if err := e.client.ContentTypes.DeleteWithEnv(env, draft); err != nil {
			return err
		}

	}
	return nil
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

func (e *contentTypeResource) getContentType(editor *ContentType) (*contentful.ContentType, error) {
	if !editor.Environment.IsUnknown() && !editor.Environment.IsNull() {

		env := &contentful.Environment{Sys: &contentful.Sys{
			ID: editor.Environment.ValueString(),
			Space: &contentful.Space{
				Sys: &contentful.Sys{ID: editor.SpaceId.ValueString()},
			},
		}}

		return e.client.ContentTypes.GetWithEnv(env, editor.ID.ValueString())
	} else {
		return e.client.ContentTypes.Get(editor.SpaceId.ValueString(), editor.ID.ValueString())
	}
}

func (e *contentTypeResource) getContentTypeControls(editor *ContentType) (*contentful.EditorInterface, error) {
	if !editor.Environment.IsUnknown() && !editor.Environment.IsNull() {

		env := &contentful.Environment{Sys: &contentful.Sys{
			ID: editor.Environment.ValueString(),
			Space: &contentful.Space{
				Sys: &contentful.Sys{ID: editor.SpaceId.ValueString()},
			},
		}}

		return e.client.EditorInterfaces.GetWithEnv(env, editor.ID.ValueString())
	} else {
		return e.client.EditorInterfaces.Get(editor.SpaceId.ValueString(), editor.ID.ValueString())
	}
}

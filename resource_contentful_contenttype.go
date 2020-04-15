package main

import (
	"github.com/hashicorp/terraform/helper/schema"
	contentful "github.com/labd/contentful-go"
	"time"
)

func resourceContentfulContentType() *schema.Resource {
	return &schema.Resource{
		Create: resourceContentTypeCreate,
		Read:   resourceContentTypeRead,
		Update: resourceContentTypeUpdate,
		Delete: resourceContentTypeDelete,

		Schema: map[string]*schema.Schema{
			"space_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"version": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"display_field": {
				Type:     schema.TypeString,
				Required: true,
			},
			"field": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"type": {
							Type:     schema.TypeString,
							Required: true,
						},
						"link_type": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"items": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"type": {
										Type:     schema.TypeString,
										Required: true,
									},
									"link_type": {
										Type:     schema.TypeString,
										Required: true,
									},
									"validation": {
										Type:     schema.TypeList,
										Optional: true,
										Elem:     generateValidationSchema(),
									},
								},
							},
						},
						"required": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"localized": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"disabled": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"omitted": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"validation": {
							Type:     schema.TypeList,
							Optional: true,
							Elem:     generateValidationSchema(),
						},
					},
				},
			},
		},
	}
}

func resourceContentTypeCreate(d *schema.ResourceData, m interface{}) (err error) {
	client := m.(*contentful.Client)
	spaceID := d.Get("space_id").(string)

	ct := &contentful.ContentType{
		Name:         d.Get("name").(string),
		DisplayField: d.Get("display_field").(string),
		Fields:       []*contentful.Field{},
	}

	if description, ok := d.GetOk("description"); ok {
		ct.Description = description.(string)
	}

	rawField := d.Get("field").([]interface{})
	for i := 0; i < len(rawField); i++ {
		field := rawField[i].(map[string]interface{})

		contentfulField := &contentful.Field{
			ID:        field["id"].(string),
			Name:      field["name"].(string),
			Type:      field["type"].(string),
			Localized: field["localized"].(bool),
			Required:  field["required"].(bool),
			Disabled:  field["disabled"].(bool),
			Omitted:   field["omitted"].(bool),
		}

		if linkType, ok := field["link_type"].(string); ok {
			contentfulField.LinkType = linkType
		}

		if validations, ok := field["validations"].([]interface{}); ok {
			parsedValidations, err := contentful.ParseValidations(validations)
			if err != nil {
				return err
			}

			contentfulField.Validations = parsedValidations
		}

		if items := processItems(field["items"].([]interface{})); items != nil {
			contentfulField.Items = items
		}

		ct.Fields = append(ct.Fields, contentfulField)
	}

	if err = client.ContentTypes.Upsert(spaceID, ct); err != nil {
		return err
	}

	if err = client.ContentTypes.Activate(spaceID, ct); err != nil {
		return err
	}

	if err = setContentTypeProperties(d, ct); err != nil {
		return err
	}

	d.SetId(ct.Sys.ID)

	return nil
}

func resourceContentTypeRead(d *schema.ResourceData, m interface{}) (err error) {
	client := m.(*contentful.Client)
	spaceID := d.Get("space_id").(string)

	_, err = client.ContentTypes.Get(spaceID, d.Id())

	return err
}

func resourceContentTypeUpdate(d *schema.ResourceData, m interface{}) (err error) {
	var existingFields []*contentful.Field
	var deletedFields []*contentful.Field

	client := m.(*contentful.Client)
	spaceID := d.Get("space_id").(string)

	ct, err := client.ContentTypes.Get(spaceID, d.Id())
	if err != nil {
		return err
	}

	ct.Name = d.Get("name").(string)
	ct.DisplayField = d.Get("display_field").(string)

	if description, ok := d.GetOk("description"); ok {
		ct.Description = description.(string)
	}

	if d.HasChange("field") {
		old, nw := d.GetChange("field")

		existingFields, deletedFields = checkFieldChanges(old.([]interface{}), nw.([]interface{}))

		ct.Fields = existingFields

		if deletedFields != nil {
			ct.Fields = append(ct.Fields, deletedFields...)
		}
	}

	// To remove a field from a content type 4 API calls need to be made.
	// Ommit the removed fields and publish the new version of the content type,
	// followed by the field removal and final publish.
	if err = client.ContentTypes.Upsert(spaceID, ct); err != nil {
		return err
	}

	if err = client.ContentTypes.Activate(spaceID, ct); err != nil {
		return err
	}

	if deletedFields != nil {
		ct.Fields = existingFields

		if err = client.ContentTypes.Upsert(spaceID, ct); err != nil {
			return err
		}

		if err = client.ContentTypes.Activate(spaceID, ct); err != nil {
			return err
		}
	}

	return setContentTypeProperties(d, ct)
}

func resourceContentTypeDelete(d *schema.ResourceData, m interface{}) (err error) {
	client := m.(*contentful.Client)
	spaceID := d.Get("space_id").(string)

	ct, err := client.ContentTypes.Get(spaceID, d.Id())
	if err != nil {
		return err
	}

	err = client.ContentTypes.Deactivate(spaceID, ct)
	if err != nil {
		return err
	}

	if err = client.ContentTypes.Delete(spaceID, ct); err != nil {
		return err
	}

	return nil
}

func setContentTypeProperties(d *schema.ResourceData, ct *contentful.ContentType) (err error) {

	if err = d.Set("version", ct.Sys.Version); err != nil {
		return err
	}

	return nil
}

func checkFieldChanges(old, new []interface{}) ([]*contentful.Field, []*contentful.Field) {
	var contentfulField *contentful.Field
	var existingFields []*contentful.Field
	var deletedFields []*contentful.Field
	var fieldRemoved bool

	for i := 0; i < len(old); i++ {
		oldField := old[i].(map[string]interface{})

		fieldRemoved = true
		for j := 0; j < len(new); j++ {
			if oldField["id"].(string) == new[j].(map[string]interface{})["id"].(string) {
				fieldRemoved = false
				break
			}
		}

		if fieldRemoved {
			deletedFields = append(deletedFields,
				&contentful.Field{
					ID:        oldField["id"].(string),
					Name:      oldField["name"].(string),
					Type:      oldField["type"].(string),
					LinkType:  oldField["link_type"].(string),
					Localized: oldField["localized"].(bool),
					Required:  oldField["required"].(bool),
					Disabled:  oldField["disabled"].(bool),
					Omitted:   true,
				})
		}
	}

	for k := 0; k < len(new); k++ {
		newField := new[k].(map[string]interface{})

		contentfulField = &contentful.Field{
			ID:        newField["id"].(string),
			Name:      newField["name"].(string),
			Type:      newField["type"].(string),
			Localized: newField["localized"].(bool),
			Required:  newField["required"].(bool),
			Disabled:  newField["disabled"].(bool),
			Omitted:   newField["omitted"].(bool),
		}

		if linkType, ok := newField["link_type"].(string); ok {
			contentfulField.LinkType = linkType
		}

		if validations, ok := newField["validations"].([]interface{}); ok {
			parsedValidations, _ := contentful.ParseValidations(validations)

			contentfulField.Validations = parsedValidations
		}

		if items := processItems(newField["items"].([]interface{})); items != nil {
			contentfulField.Items = items
		}

		existingFields = append(existingFields, contentfulField)
	}

	return existingFields, deletedFields
}

func processItems(fieldItems []interface{}) *contentful.FieldTypeArrayItem {
	var items *contentful.FieldTypeArrayItem

	for i := 0; i < len(fieldItems); i++ {
		item := fieldItems[i].(map[string]interface{})

		var validations []contentful.FieldValidation

		if fieldValidations, ok := item["validations"].([]interface{}); ok {
			validations, _ = contentful.ParseValidations(fieldValidations)
		}

		items = &contentful.FieldTypeArrayItem{
			Type:        item["type"].(string),
			Validations: validations,
			LinkType:    item["link_type"].(string),
		}
	}
	return items
}

// Expanders
// noinspection GoUnusedFunction
func expandContentTypeField(in interface{}) *contentful.Field {
	field := &contentful.Field{}
	m := in.(map[string]interface{})
	if v, ok := m["id"].(string); ok {
		field.ID = v
	}
	if v, ok := m["name"].(string); ok {
		field.Name = v
	}
	if v, ok := m["type"].(string); ok {
		field.Type = v
	}
	if v, ok := m["link_type"].(string); ok {
		field.LinkType = v
	}
	if v, ok := m["required"].(bool); ok {
		field.Required = v
	}
	if v, ok := m["localized"].(bool); ok {
		field.Localized = v
	}
	if v, ok := m["disabled"].(bool); ok {
		field.Disabled = v
	}
	if v, ok := m["omitted"].(bool); ok {
		field.Omitted = v
	}
	if v, ok := m["validation"].([]interface{}); ok && len(v) > 0 {
		validations := make([]contentful.FieldValidation, len(v))
		for i, val := range v {
			validations[i] = expandContentTypeFieldValidation(val)
		}
		field.Validations = validations
	}
	return field
}

func expandContentTypeFieldValidation(in interface{}) contentful.FieldValidation {
	m := in.(map[string]interface{})
	if v, ok := m["link"].([]string); ok {
		return contentful.FieldValidationLink{
			LinkContentType: v,
		}
	}
	if v, ok := m["mime_type"].([]string); ok {
		return contentful.FieldValidationMimeType{
			MimeTypes: v,
		}
	}
	if v, ok := m["dimension"].([]interface{}); ok {
		return expandContentTypeFieldValidationDimension(v)
	}
	if v, ok := m["size"].([]interface{}); ok {
		return expandContentTypeFieldValidationSize(v)
	}
	if v, ok := m["file_size"].([]interface{}); ok {
		return expandContentTypeFieldValidationFileSize(v)
	}
	if v, ok := m["unique"].(bool); ok {
		return contentful.FieldValidationUnique{
			Unique: v,
		}
	}
	if v, ok := m["range"].([]interface{}); ok {
		return expandContentTypeFieldValidationRange(v)
	}
	if v, ok := m["date"].([]interface{}); ok {
		return expandContentTypeFieldValidationDate(v)
	}
	return nil
}

func expandContentTypeFieldValidationDimension(in []interface{}) contentful.FieldValidation {
	if len(in) == 0 || in[0] == nil {
		return contentful.FieldValidationDimension{}
	}

	validation := contentful.FieldValidationDimension{}
	m := in[0].(map[string]interface{})
	if v, ok := m["min_width"].(float64); ok {
		if validation.Width == nil {
			validation.Width = &contentful.MinMax{}
		}
		validation.Width.Min = v
	}
	if v, ok := m["max_width"].(float64); ok {
		if validation.Width == nil {
			validation.Width = &contentful.MinMax{}
		}
		validation.Width.Max = v
	}
	if v, ok := m["min_height"].(float64); ok {
		if validation.Height == nil {
			validation.Width = &contentful.MinMax{}
		}
		validation.Height.Min = v
	}
	if v, ok := m["max_height"].(float64); ok {
		if validation.Height == nil {
			validation.Width = &contentful.MinMax{}
		}
		validation.Height.Max = v
	}
	if v, ok := m["err_message"].(string); ok {
		validation.ErrorMessage = v
	}
	return validation
}

func expandContentTypeFieldValidationSize(in []interface{}) contentful.FieldValidation {
	if len(in) == 0 || in[0] == nil {
		return contentful.FieldValidationSize{}
	}

	validation := contentful.FieldValidationSize{}
	m := in[0].(map[string]interface{})
	if v, ok := m["min"].(float64); ok {
		if validation.Size == nil {
			validation.Size = &contentful.MinMax{}
		}
		validation.Size.Min = v
	}
	if v, ok := m["max"].(float64); ok {
		if validation.Size == nil {
			validation.Size = &contentful.MinMax{}
		}
		validation.Size.Max = v
	}
	if v, ok := m["err_message"].(string); ok {
		validation.ErrorMessage = v
	}
	return validation
}

func expandContentTypeFieldValidationFileSize(in []interface{}) contentful.FieldValidation {
	if len(in) == 0 || in[0] == nil {
		return contentful.FieldValidationFileSize{}
	}

	validation := contentful.FieldValidationFileSize{}
	m := in[0].(map[string]interface{})
	if v, ok := m["min"].(float64); ok {
		if validation.Size == nil {
			validation.Size = &contentful.MinMax{}
		}
		validation.Size.Min = v
	}
	if v, ok := m["max"].(float64); ok {
		if validation.Size == nil {
			validation.Size = &contentful.MinMax{}
		}
		validation.Size.Max = v
	}
	if v, ok := m["err_message"].(string); ok {
		validation.ErrorMessage = v
	}
	return validation
}

func expandContentTypeFieldValidationRange(in []interface{}) contentful.FieldValidation {
	if len(in) == 0 || in[0] == nil {
		return contentful.FieldValidationRange{}
	}

	validation := contentful.FieldValidationRange{}
	m := in[0].(map[string]interface{})
	if v, ok := m["min"].(float64); ok {
		if validation.Range == nil {
			validation.Range = &contentful.MinMax{}
		}
		validation.Range.Min = v
	}
	if v, ok := m["max"].(float64); ok {
		if validation.Range == nil {
			validation.Range = &contentful.MinMax{}
		}
		validation.Range.Max = v
	}
	if v, ok := m["err_message"].(string); ok {
		validation.ErrorMessage = v
	}
	return validation
}

func expandContentTypeFieldValidationDate(in []interface{}) contentful.FieldValidation {
	if len(in) == 0 || in[0] == nil {
		return contentful.FieldValidationDate{}
	}

	validation := contentful.FieldValidationDate{}
	m := in[0].(map[string]interface{})
	if v, ok := m["min"].(time.Time); ok {
		if validation.Range == nil {
			validation.Range = &contentful.DateMinMax{}
		}
		validation.Range.Min = v
	}
	if v, ok := m["max"].(time.Time); ok {
		if validation.Range == nil {
			validation.Range = &contentful.DateMinMax{}
		}
		validation.Range.Max = v
	}
	if v, ok := m["err_message"].(string); ok {
		validation.ErrorMessage = v
	}
	return validation
}

// noinspection GoUnusedFunction
func expandContentTypeFieldValidationRegex(in []interface{}) contentful.FieldValidation {
	if len(in) == 0 || in[0] == nil {
		return contentful.FieldValidationRegex{}
	}

	validation := contentful.FieldValidationRegex{}
	m := in[0].(map[string]interface{})
	if v, ok := m["pattern"].(string); ok {
		if validation.Regex == nil {
			validation.Regex = &contentful.Regex{}
		}
		validation.Regex.Pattern = v
	}
	if v, ok := m["flags"].(string); ok {
		if validation.Regex == nil {
			validation.Regex = &contentful.Regex{}
		}
		validation.Regex.Flags = v
	}
	if v, ok := m["err_message"].(string); ok {
		validation.ErrorMessage = v
	}
	return validation
}

func generateValidationSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"link": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"mime_type": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"dimension": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"min_width": {
							Type:     schema.TypeFloat,
							Required: true,
						},
						"max_width": {
							Type:     schema.TypeFloat,
							Required: true,
						},
						"min_height": {
							Type:     schema.TypeFloat,
							Required: true,
						},
						"max_height": {
							Type:     schema.TypeFloat,
							Required: true,
						},
						"err_message": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"size": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"min": {
							Type:     schema.TypeFloat,
							Optional: true,
						},
						"max": {
							Type:     schema.TypeFloat,
							Optional: true,
						},
						"err_message": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"file_size": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"min": {
							Type:     schema.TypeFloat,
							Optional: true,
						},
						"max": {
							Type:     schema.TypeFloat,
							Optional: true,
						},
						"err_message": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"unique": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"range": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"min": {
							Type:     schema.TypeFloat,
							Optional: true,
						},
						"max": {
							Type:     schema.TypeFloat,
							Optional: true,
						},
						"err_message": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"date": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"min": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"max": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"err_message": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"regex": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"pattern": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"flags": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"err_message": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
		},
	}
}

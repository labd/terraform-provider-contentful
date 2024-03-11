resource "contentful_contenttype" "{{ .identifier }}" {
  space_id              = "{{ .spaceId }}"
  environment   = "master"
  name                  = "tf_test1"
  description           = "Terraform Acc Test Content Type description change"
  display_field         = "field1"
  manage_field_controls = true
  fields = [{
    id       = "field1"
    name     = "Field 1 name change"
    required = true
    type     = "Text"
    },
    {
      id       = "field1"
      name     = "Field 1 name change"
      required = true
      type     = "Text"
    },
    {
    id       = "field3"
    name     = "Field 3 new field"
    required = true
    type     = "Integer"
    control = {
      widget_id        = "numberEditor"
      widget_namespace = "builtin"
      settings = {
        help_text    = "blabla"
        bulk_editing = true
      }
    },
  }]
}
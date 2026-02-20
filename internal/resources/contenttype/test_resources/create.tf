resource "contentful_contenttype" "{{ .identifier }}" {
  space_id      = "{{ .spaceId }}"
  environment   = "master-2026-02-20"
{{.id_definition}}
  name          = "tf_test1"
  description   = "{{.desc}}"
  display_field = "field1"
  fields = [{
    id       = "field1"
    name     = "Field 1 name change"
    required = true
    type     = "Text"
    }, {
    id       = "field3"
    name     = "Field 3 new field"
    required = true
    type     = "Integer"
  }, {
    id          = "field4"
    name        = "Field 4 new field"
    required    = true
    type        = "RichText"
    validations = [
      {
        enabled_marks = ["bold"]
        message       = "Supports only bold."
      },
      {
        enabled_node_types = ["embedded-asset-block"]
      }
    ]
  },
{
id       = "field5"
name     = "Field 5 new field"
type     = "Array"
items = {
type = "Symbol"
}
localized     = true
default_value = {
array : {
"en-US" : ["test"],
},
}
},
]
}

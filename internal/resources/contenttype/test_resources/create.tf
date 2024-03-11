resource "contentful_contenttype" "{{ .identifier }}" {
  space_id      = "{{ .spaceId }}"
  environment   = "master"
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
  }]
}
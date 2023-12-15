resource "contentful_contenttype" "{{ .identifier }}" {
  space_id      = "{{ .spaceId }}"
  name          = "tf_test1"
  description   = "Terraform Acc Test Content Type description change"
  display_field = "field1"
  fields = [{
    id       = "field3"
    name     = "Field 3 new field"
    required = true
    type     = "Integer"
    }, {
    id       = "field1"
    name     = "Field 1 name change"
    required = true
    type     = "Text"
  }]
}
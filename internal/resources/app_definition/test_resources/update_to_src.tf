resource "contentful_app_definition" "{{ .identifier }}" {
  name       = "tf_test1"
  use_bundle = false
  locations  = [{ location = "entry-field", "field_types" = [{ "type" = "Symbol" }] }]
  src        = "http://localhost"
}
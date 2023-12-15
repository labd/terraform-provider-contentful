resource "contentful_app_definition" "{{ .identifier }}" {
  name       = "tf_test1"
  use_bundle = true
  locations  = [{ location = "entry-field", "field_types" = [{ "type" = "Symbol" }] }, { location = "dialog" }]
}
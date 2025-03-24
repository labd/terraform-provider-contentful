resource "contentful_app_definition" "example_app_definition" {
  name       = "test_app_definition"
  use_bundle = false
  src        = "http://localhost:3000"
  locations  = [{ location = "app-config" }, { location = "dialog" }, { location = "entry-editor" }]
}
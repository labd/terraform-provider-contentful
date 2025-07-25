resource "contentful_app_definition" "example_app_definition" {
  name       = "test_app_definition"
  use_bundle = false
  src        = "http://localhost:3000"
  locations  = [{ location = "app-config" }, { location = "dialog" }, { location = "entry-editor" }]
}

resource "contentful_app_event_subscription" "example_app_event_subscription" {
  app_definition_id = contentful_app_definition.example_app_definition.id
  target_url        = "https://example.com/webhook"
  topics            = ["Entry.save"]
}

resource "contentful_app_installation" "example_app_installation" {
  space_id    = "space-id"
  environment = "master"

  app_definition_id = contentful_app_definition.example_app_definition.id

  parameters = jsonencode({
    "example" : "one",
    "nested" : {
      "example" : "two"
    }
  })
}
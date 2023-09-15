resource "contentful_apikey" "myapikey" {
  space_id = "space-id"

  name        = "api-key-name"
  description = "a-great-key"
}
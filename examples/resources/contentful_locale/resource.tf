resource "contentful_locale" "example_locale" {
  space_id = "spaced-id"

  name          = "locale-name"
  code          = "de"
  fallback_code = "en-US"
  optional      = false
  cda           = false
  cma           = true
}
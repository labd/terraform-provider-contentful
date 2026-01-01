resource "contentful_environment_alias" "master" {
  space_id       = "space-id"
  alias_id       = "master" # You must have at least one called master
  environment_id = "environment-id"
}

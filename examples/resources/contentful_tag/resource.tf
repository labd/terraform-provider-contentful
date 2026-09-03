resource "contentful_tag" "example" {
  space_id    = "space-id"
  environment = "master"

  id         = "campaign"
  name       = "Campaign"
  visibility = "public"
}

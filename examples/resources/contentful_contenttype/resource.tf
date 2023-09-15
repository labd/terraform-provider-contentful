resource "contentful_contenttype" "example_contenttype" {
  space_id      = "space-id"
  name          = "tf_linked"
  description   = "content type description"
  display_field = "asset_field"
  field {
    id   = "asset_field"
    name = "Asset Field"
    type = "Array"
    items {
      type      = "Link"
      link_type = "Asset"
    }
    required = true
  }
  field {
    id        = "entry_link_field"
    name      = "Entry Link Field"
    type      = "Link"
    link_type = "Entry"
    validations = [
      jsonencode({
        linkContentType = [
          contentful_contenttype.some_other_content_type.id
        ]
      })
    ]
    required = false
  }
}
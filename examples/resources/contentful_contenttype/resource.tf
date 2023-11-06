resource "contentful_contenttype" "example_contenttype" {
  space_id      = "space-id"
  name          = "tf_linked"
  description   = "content type description"
  display_field = "asset_field"
  fields =[{
    id   = "asset_field"
    name = "Asset Field"
    type = "Array"
    items = {
      type      = "Link"
      link_type = "Asset"
    }
    required = true
  }, {
    id        = "entry_link_field"
    name      = "Entry Link Field"
    type      = "Link"
    link_type = "Entry"
    validations = [{
        link_content_type = [
          contentful_contenttype.some_other_content_type.id
        ]
      }
    ]
    required = false
  }]
}
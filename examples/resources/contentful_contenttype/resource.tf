resource "contentful_contenttype" "example_contenttype" {
  space_id              = "space-id"
  environment           = "master"
  id                    = "tf_linked"
  name                  = "tf_linked"
  description           = "content type description"
  display_field         = "asset_field"
  manage_field_controls = true
  fields = [{
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
    }, {
    id       = "select",
    name     = "Select Field",
    type     = "Symbol",
    required = true,
    validations = [
      {
        in = [
          "choice 1",
          "choice 2",
          "choice 3",
          "choice 4"
        ]
      }
    ]
    control = {
      widget_id        = "radio"
      widget_namespace = "builtin"
    }
  }]
}
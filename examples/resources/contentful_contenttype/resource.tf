resource "contentful_contenttype" "some_other_content_type" {
  space_id      = "space-id"
  environment   = "provider-test"
  id            = "some_other_content_type"
  name          = "some_other_content_type"
  description   = "some other content type description"
  display_field = "content"

  fields = [{
    id       = "content"
    name     = "Content"
    type     = "RichText"
    required = true
  }]
}

resource "contentful_contenttype" "example_contenttype" {
  space_id      = "space-id"
  environment   = "master"
  id            = "tf_linked"
  name          = "tf_linked"
  description   = "content type description"
  display_field = "asset_field"

  fields = [
    {
      id   = "asset_field"
      name = "Asset Field"
      type = "Array"
      items = {
        type      = "Link"
        link_type = "Asset"
      }
      required = true
    },
    {
      id        = "entry_link_field"
      name      = "Entry Link Field"
      type      = "Link"
      link_type = "Entry"
      validations = [
        {
          link_content_type = [contentful_contenttype.some_other_content_type.id]
        }
      ]
      required = false
    },
    {
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
    },
    {
      id   = "themeColor"
      name = "Theme Color"
      type = "Symbol"
      validations = [{
        in = ["green", "pink", "turquoise", "yellow", "purple"]
      }]
      default_value = {
        string = {
          "en-US" = "green"
        }
      }
      required = false
    },
    {
      id   = "content"
      name = "Content"
      type = "RichText"
      validations = [
        {
          nodes = {
            entry_hyperlink = [
              {
                size = {
                  min = 1
                  max = 1
                },
                message = "test",
              },
              {
                link_content_type = [
                  contentful_contenttype.some_other_content_type.id
                ],
                message = "test"
              },
            ]
          }
        }
      ]
      required = false
    }
  ]
}

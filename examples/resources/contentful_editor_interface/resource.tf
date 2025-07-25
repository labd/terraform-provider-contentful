resource "contentful_contenttype" "example_contenttype" {
  space_id      = "space-id"
  environment   = "master"
  id            = "example_contenttype"
  name          = "name"
  description   = "content type description"
  display_field = "name"

  fields = [
    {
      id       = "name"
      name     = "Name"
      type     = "Text"
      required = true
    },
    {
      id       = "content"
      name     = "Content"
      type     = "RichText"
      required = false
    },
    {
      id   = "tags"
      name = "Tags"
      type = "Array"
      items = {
        type = "Symbol"
      }
      required = false
    }
  ]
}

resource "contentful_editor_interface" "example_editor_interface" {
  space_id     = "space-id"
  environment  = "master"
  content_type = contentful_contenttype.example_contenttype.id
  controls = [
    {
      field_id         = "name"
      widget_id        = "singleLine"
      widget_namespace = "builtin"
    },
    {
      field_id         = "content"
      widget_id        = "richTextEditor"
      widget_namespace = "builtin"
    },
    {
      field_id         = "tags"
      widget_id        = "listInput"
      widget_namespace = "builtin"
    }
  ]
  sidebar = [
    {
      widget_id        = "content-preview-widget"
      widget_namespace = "sidebar-builtin"
    },
    {
      widget_id        = "translation-widget"
      widget_namespace = "sidebar-builtin"
    }
  ]
  editors = [
    {
      widget_namespace = "editor-builtin",
      widget_id        = "default-editor",
      disabled         = true
    }
  ]
}

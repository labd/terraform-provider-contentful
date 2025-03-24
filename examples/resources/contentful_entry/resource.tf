resource "contentful_entry" "example_entry" {
  entry_id       = "mytestentry"
  space_id       = "space-id"
  contenttype_id = "type-id"
  environment    = "master"
  field {
    id      = "field1"
    content = "Hello, World!"
    locale  = "en-US"
  }
  field {
    id      = "field2"
    content = "Lettuce is healthy!"
    locale  = "en-US"
  }
  field {
    id     = "content"
    locale = "en-US"
    content = jsonencode({
      data = {},
      content = [
        {
          nodeType = "paragraph",
          content = [
            {
              nodeType = "text",
              marks    = [],
              value    = "This is a paragraph",
              data     = {},
            },
          ],
          data = {},
        }
      ],
      nodeType = "document"
    })
  }
  published  = false
  archived   = false
  depends_on = [contentful_contenttype.mycontenttype]
}

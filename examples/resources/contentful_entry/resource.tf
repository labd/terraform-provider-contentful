resource "contentful_entry" "example_entry" {
  entry_id       = "mytestentry"
  space_id       = "space-id"
  contenttype_id = "type-id"
  locale         = "en-US"
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
  published  = false
  archived   = false
  depends_on = [contentful_contenttype.mycontenttype]
}
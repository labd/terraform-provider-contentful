resource "contentful_webhook" "example_webhook" {
  space_id = "space-id"

  active = true

  name = "webhook-name"
  url  = "https://www.example.com/test"
  topics = [
    "Entry.create",
    "ContentType.create",
  ]

  # Use headers_v2 to support marking individual header values as secret.
  headers_v2 = [
    {
      key   = "X-Custom-Header"
      value = "header-value"
    },
    {
      key    = "Authorization"
      value  = "secret-token"
      secret = true
    },
  ]

  filters = jsonencode([
    { in : [{ "doc" : "sys.environment.sys.id" }, ["testing", "staging"]] },
    { not : { equals : [{ "doc" : "sys.environment.sys.id" }, "master"] } },
  ])
}

# Using the deprecated headers attribute (plain key/value map, no secret support).
# Migrate to headers_v2 to gain secret header support.
resource "contentful_webhook" "example_webhook_legacy_headers" {
  space_id = "space-id"

  active = true

  name = "webhook-name-legacy"
  url  = "https://www.example.com/test"
  topics = [
    "Entry.create",
    "ContentType.create",
  ]

  # Deprecated: use headers_v2 instead.
  headers = {
    header1 = "header1-value"
    header2 = "header2-value"
  }
}

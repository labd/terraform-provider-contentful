resource "contentful_webhook" "example_webhook" {
  space_id = "space-id"

  name = "webhook-name"
  url  = "https://www.example.com/test"
  topics = [
    "Entry.create",
    "ContentType.create",
  ]
  headers = {
    header1 = "header1-value"
    header2 = "header2-value"
  }
  http_basic_auth_username = "username"
  http_basic_auth_password = "password"
}
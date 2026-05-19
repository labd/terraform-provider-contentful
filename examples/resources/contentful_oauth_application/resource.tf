resource "contentful_oauth_application" "example" {
  name         = "Authoring server"
  description  = "OAuth app used by the authoring server to sign editors in"
  redirect_uri = "https://authoring.example.com/oauth/callback"
  scopes       = ["content_management_manage"]
  confidential = true
}

# The client_secret is only returned by Contentful at creation time and cannot
# be retrieved later. Wire it directly to your secret store rather than relying
# on Terraform state as the source of truth.
output "client_id" {
  value = contentful_oauth_application.example.client_id
}

output "client_secret" {
  value     = contentful_oauth_application.example.client_secret
  sensitive = true
}

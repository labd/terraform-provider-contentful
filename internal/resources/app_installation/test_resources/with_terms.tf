resource "contentful_app_installation" "{{ .identifier }}" {

  space_id    = "{{ .spaceId }}"
  environment = "{{ .environment }}"

  app_definition_id = "{{ .appDefinitionId }}"

  parameters = jsonencode({
    cpaToken = "not-working-ever"
  })

  accepted_terms = ["i-accept-end-user-license-agreement", "i-accept-marketplace-terms-of-service", "i-accept-privacy-policy"]

}
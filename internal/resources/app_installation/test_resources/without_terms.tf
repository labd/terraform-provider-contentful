resource "contentful_app_installation" "{{ .identifier }}" {

  space_id    = "{{ .spaceId }}"
  environment = "{{ .environment }}"

  app_definition_id = "{{ .appDefinitionId }}"

  parameters = jsonencode({})

}
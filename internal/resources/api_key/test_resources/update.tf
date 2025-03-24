resource "contentful_apikey" "myapikey" {
  space_id      = "{{ .spaceId }}"

  name = "{{ .name }}-updated"
  description = "{{ .description }}-updated"
}
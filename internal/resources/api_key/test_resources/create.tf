resource "contentful_apikey" "myapikey" {
  space_id      = "{{ .spaceId }}"

  name = "{{ .name }}"
  description = "{{ .description }}"

  {{if .environments}}

  environments = [ {{range .environments}}"{{.}}",{{end}}]

  {{end}}

}
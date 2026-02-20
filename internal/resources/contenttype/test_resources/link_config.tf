resource "contentful_contenttype" "{{ .identifier }}" {
  space_id = "{{ .spaceId }}"
  environment   = "master-2026-02-20"
  name = "tf_test1"
  description = "Terraform Acc Test Content Type description change"
  display_field = "field1"
  fields = [
    {
      id        = "field1"
      name      = "Field 1 name change"
      required  = true
      type      = "Text"
    },
    {
      id        = "field3"
      name      = "Field 3 new field"
      required  = true
      type      = "Integer"
    }
  ]
}

resource "contentful_contenttype" "{{ .linkIdentifier }}" {
  space_id = "{{ .spaceId }}"
  name          = "tf_linked"
  environment   = "master-2026-02-20"
  description   = "Terraform Acc Test Content Type with links"
  display_field = "asset_field"
  fields = [
    {
      id   = "asset_field"
      name = "Asset Field"
      type = "Array"
      items = {
        type      = "Link"
        link_type = "Asset"
      }
      required = true
    },
    {
      id        = "entry_link_field"
      name      = "Entry Link Field"
      type      = "Link"
      link_type = "Entry"
      validations = [
        {
          link_content_type = [contentful_contenttype.{{ .identifier }}.id ]
        }
      ]
    }
  ]
}

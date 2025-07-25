---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "contentful_contenttype Resource - terraform-provider-contentful"
subcategory: ""
description: |-
  Todo for explaining contenttype
---

# contentful_contenttype (Resource)

Todo for explaining contenttype

## Example Usage

```terraform
resource "contentful_contenttype" "some_other_content_type" {
  space_id      = "space-id"
  environment   = "provider-test"
  id            = "some_other_content_type"
  name          = "some_other_content_type"
  description   = "some other content type description"
  display_field = "content"

  fields = [{
    id       = "content"
    name     = "Content"
    type     = "RichText"
    required = true
  }]
}

resource "contentful_contenttype" "example_contenttype" {
  space_id      = "space-id"
  environment   = "master"
  id            = "tf_linked"
  name          = "tf_linked"
  description   = "content type description"
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
          link_content_type = [contentful_contenttype.some_other_content_type.id]
        }
      ]
      required = false
    },
    {
      id       = "select",
      name     = "Select Field",
      type     = "Symbol",
      required = true,
      validations = [
        {
          in = [
            "choice 1",
            "choice 2",
            "choice 3",
            "choice 4"
          ]
        }
      ]
    },
    {
      id   = "content"
      name = "Content"
      type = "RichText"
      validations = [
        {
          nodes = {
            entry_hyperlink = [
              {
                size = {
                  min = 1
                  max = 1
                },
                message = "test",
              },
              {
                link_content_type = [
                  contentful_contenttype.some_other_content_type.id
                ],
                message = "test"
              },
            ]
          }
        }
      ]
      required = false
    }
  ]
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `display_field` (String)
- `environment` (String)
- `fields` (Attributes List) (see [below for nested schema](#nestedatt--fields))
- `name` (String)
- `space_id` (String) space id

### Optional

- `description` (String)
- `id` (String) content type id

### Read-Only

- `version` (Number)

<a id="nestedatt--fields"></a>
### Nested Schema for `fields`

Required:

- `id` (String)
- `name` (String)
- `type` (String)

Optional:

- `default_value` (Attributes) (see [below for nested schema](#nestedatt--fields--default_value))
- `disabled` (Boolean)
- `items` (Attributes) (see [below for nested schema](#nestedatt--fields--items))
- `link_type` (String)
- `localized` (Boolean)
- `omitted` (Boolean)
- `required` (Boolean)
- `validations` (Attributes List) (see [below for nested schema](#nestedatt--fields--validations))

<a id="nestedatt--fields--default_value"></a>
### Nested Schema for `fields.default_value`

Optional:

- `bool` (Map of Boolean)
- `string` (Map of String)


<a id="nestedatt--fields--items"></a>
### Nested Schema for `fields.items`

Required:

- `type` (String)

Optional:

- `link_type` (String)
- `validations` (Attributes List) (see [below for nested schema](#nestedatt--fields--items--validations))

<a id="nestedatt--fields--items--validations"></a>
### Nested Schema for `fields.items.validations`

Optional:

- `asset_file_size` (Attributes) (see [below for nested schema](#nestedatt--fields--items--validations--asset_file_size))
- `enabled_marks` (List of String)
- `enabled_node_types` (List of String)
- `in` (List of String)
- `link_content_type` (List of String)
- `link_mimetype_group` (List of String)
- `message` (String) Defines the message that is shown to the user when the validation fails. It can be used to provide more information about the validation.
- `nodes` (Attributes) (see [below for nested schema](#nestedatt--fields--items--validations--nodes))
- `range` (Attributes) (see [below for nested schema](#nestedatt--fields--items--validations--range))
- `regexp` (Attributes) (see [below for nested schema](#nestedatt--fields--items--validations--regexp))
- `size` (Attributes) (see [below for nested schema](#nestedatt--fields--items--validations--size))
- `unique` (Boolean)

<a id="nestedatt--fields--items--validations--asset_file_size"></a>
### Nested Schema for `fields.items.validations.asset_file_size`

Optional:

- `max` (Number)
- `min` (Number)


<a id="nestedatt--fields--items--validations--nodes"></a>
### Nested Schema for `fields.items.validations.nodes`

Optional:

- `asset_hyperlink` (Attributes List) (see [below for nested schema](#nestedatt--fields--items--validations--unique--asset_hyperlink))
- `embedded_asset_block` (Attributes List) (see [below for nested schema](#nestedatt--fields--items--validations--unique--embedded_asset_block))
- `embedded_entry_block` (Attributes List) (see [below for nested schema](#nestedatt--fields--items--validations--unique--embedded_entry_block))
- `embedded_entry_inline` (Attributes List) (see [below for nested schema](#nestedatt--fields--items--validations--unique--embedded_entry_inline))
- `embedded_resource_block` (Attributes) (see [below for nested schema](#nestedatt--fields--items--validations--unique--embedded_resource_block))
- `embedded_resource_inline` (Attributes) (see [below for nested schema](#nestedatt--fields--items--validations--unique--embedded_resource_inline))
- `entry_hyperlink` (Attributes List) (see [below for nested schema](#nestedatt--fields--items--validations--unique--entry_hyperlink))
- `resource_hyperlink` (Attributes) (see [below for nested schema](#nestedatt--fields--items--validations--unique--resource_hyperlink))

<a id="nestedatt--fields--items--validations--unique--asset_hyperlink"></a>
### Nested Schema for `fields.items.validations.unique.asset_hyperlink`

Optional:

- `message` (String) Defines the message that is shown to the user when the validation fails. It can be used to provide more information about the validation.
- `size` (Attributes) (see [below for nested schema](#nestedatt--fields--items--validations--unique--asset_hyperlink--size))

<a id="nestedatt--fields--items--validations--unique--asset_hyperlink--size"></a>
### Nested Schema for `fields.items.validations.unique.asset_hyperlink.size`

Optional:

- `max` (Number)
- `min` (Number)



<a id="nestedatt--fields--items--validations--unique--embedded_asset_block"></a>
### Nested Schema for `fields.items.validations.unique.embedded_asset_block`

Optional:

- `message` (String) Defines the message that is shown to the user when the validation fails. It can be used to provide more information about the validation.
- `size` (Attributes) (see [below for nested schema](#nestedatt--fields--items--validations--unique--embedded_asset_block--size))

<a id="nestedatt--fields--items--validations--unique--embedded_asset_block--size"></a>
### Nested Schema for `fields.items.validations.unique.embedded_asset_block.size`

Optional:

- `max` (Number)
- `min` (Number)



<a id="nestedatt--fields--items--validations--unique--embedded_entry_block"></a>
### Nested Schema for `fields.items.validations.unique.embedded_entry_block`

Optional:

- `link_content_type` (List of String)
- `message` (String) Defines the message that is shown to the user when the validation fails. It can be used to provide more information about the validation.
- `size` (Attributes) (see [below for nested schema](#nestedatt--fields--items--validations--unique--embedded_entry_block--size))

<a id="nestedatt--fields--items--validations--unique--embedded_entry_block--size"></a>
### Nested Schema for `fields.items.validations.unique.embedded_entry_block.size`

Optional:

- `max` (Number)
- `min` (Number)



<a id="nestedatt--fields--items--validations--unique--embedded_entry_inline"></a>
### Nested Schema for `fields.items.validations.unique.embedded_entry_inline`

Optional:

- `link_content_type` (List of String)
- `message` (String) Defines the message that is shown to the user when the validation fails. It can be used to provide more information about the validation.
- `size` (Attributes) (see [below for nested schema](#nestedatt--fields--items--validations--unique--embedded_entry_inline--size))

<a id="nestedatt--fields--items--validations--unique--embedded_entry_inline--size"></a>
### Nested Schema for `fields.items.validations.unique.embedded_entry_inline.size`

Optional:

- `max` (Number)
- `min` (Number)



<a id="nestedatt--fields--items--validations--unique--embedded_resource_block"></a>
### Nested Schema for `fields.items.validations.unique.embedded_resource_block`

Optional:

- `allowed_resources` (Attributes List) Defines the entities that can be referenced by the field. It is only used for cross-space references. (see [below for nested schema](#nestedatt--fields--items--validations--unique--embedded_resource_block--allowed_resources))
- `validations` (Attributes List) (see [below for nested schema](#nestedatt--fields--items--validations--unique--embedded_resource_block--validations))

<a id="nestedatt--fields--items--validations--unique--embedded_resource_block--allowed_resources"></a>
### Nested Schema for `fields.items.validations.unique.embedded_resource_block.allowed_resources`

Optional:

- `content_types` (List of String)
- `source` (String)
- `type` (String)


<a id="nestedatt--fields--items--validations--unique--embedded_resource_block--validations"></a>
### Nested Schema for `fields.items.validations.unique.embedded_resource_block.validations`

Optional:

- `message` (String) Defines the message that is shown to the user when the validation fails. It can be used to provide more information about the validation.
- `size` (Attributes) (see [below for nested schema](#nestedatt--fields--items--validations--unique--embedded_resource_block--validations--size))

<a id="nestedatt--fields--items--validations--unique--embedded_resource_block--validations--size"></a>
### Nested Schema for `fields.items.validations.unique.embedded_resource_block.validations.size`

Optional:

- `max` (Number)
- `min` (Number)




<a id="nestedatt--fields--items--validations--unique--embedded_resource_inline"></a>
### Nested Schema for `fields.items.validations.unique.embedded_resource_inline`

Optional:

- `allowed_resources` (Attributes List) Defines the entities that can be referenced by the field. It is only used for cross-space references. (see [below for nested schema](#nestedatt--fields--items--validations--unique--embedded_resource_inline--allowed_resources))
- `validations` (Attributes List) (see [below for nested schema](#nestedatt--fields--items--validations--unique--embedded_resource_inline--validations))

<a id="nestedatt--fields--items--validations--unique--embedded_resource_inline--allowed_resources"></a>
### Nested Schema for `fields.items.validations.unique.embedded_resource_inline.allowed_resources`

Optional:

- `content_types` (List of String)
- `source` (String)
- `type` (String)


<a id="nestedatt--fields--items--validations--unique--embedded_resource_inline--validations"></a>
### Nested Schema for `fields.items.validations.unique.embedded_resource_inline.validations`

Optional:

- `message` (String) Defines the message that is shown to the user when the validation fails. It can be used to provide more information about the validation.
- `size` (Attributes) (see [below for nested schema](#nestedatt--fields--items--validations--unique--embedded_resource_inline--validations--size))

<a id="nestedatt--fields--items--validations--unique--embedded_resource_inline--validations--size"></a>
### Nested Schema for `fields.items.validations.unique.embedded_resource_inline.validations.size`

Optional:

- `max` (Number)
- `min` (Number)




<a id="nestedatt--fields--items--validations--unique--entry_hyperlink"></a>
### Nested Schema for `fields.items.validations.unique.entry_hyperlink`

Optional:

- `link_content_type` (List of String)
- `message` (String) Defines the message that is shown to the user when the validation fails. It can be used to provide more information about the validation.
- `size` (Attributes) (see [below for nested schema](#nestedatt--fields--items--validations--unique--entry_hyperlink--size))

<a id="nestedatt--fields--items--validations--unique--entry_hyperlink--size"></a>
### Nested Schema for `fields.items.validations.unique.entry_hyperlink.size`

Optional:

- `max` (Number)
- `min` (Number)



<a id="nestedatt--fields--items--validations--unique--resource_hyperlink"></a>
### Nested Schema for `fields.items.validations.unique.resource_hyperlink`

Optional:

- `allowed_resources` (Attributes List) Defines the entities that can be referenced by the field. It is only used for cross-space references. (see [below for nested schema](#nestedatt--fields--items--validations--unique--resource_hyperlink--allowed_resources))
- `validations` (Attributes List) (see [below for nested schema](#nestedatt--fields--items--validations--unique--resource_hyperlink--validations))

<a id="nestedatt--fields--items--validations--unique--resource_hyperlink--allowed_resources"></a>
### Nested Schema for `fields.items.validations.unique.resource_hyperlink.allowed_resources`

Optional:

- `content_types` (List of String)
- `source` (String)
- `type` (String)


<a id="nestedatt--fields--items--validations--unique--resource_hyperlink--validations"></a>
### Nested Schema for `fields.items.validations.unique.resource_hyperlink.validations`

Optional:

- `message` (String) Defines the message that is shown to the user when the validation fails. It can be used to provide more information about the validation.
- `size` (Attributes) (see [below for nested schema](#nestedatt--fields--items--validations--unique--resource_hyperlink--validations--size))

<a id="nestedatt--fields--items--validations--unique--resource_hyperlink--validations--size"></a>
### Nested Schema for `fields.items.validations.unique.resource_hyperlink.validations.size`

Optional:

- `max` (Number)
- `min` (Number)





<a id="nestedatt--fields--items--validations--range"></a>
### Nested Schema for `fields.items.validations.range`

Optional:

- `max` (Number)
- `min` (Number)


<a id="nestedatt--fields--items--validations--regexp"></a>
### Nested Schema for `fields.items.validations.regexp`

Optional:

- `pattern` (String)


<a id="nestedatt--fields--items--validations--size"></a>
### Nested Schema for `fields.items.validations.size`

Optional:

- `max` (Number)
- `min` (Number)




<a id="nestedatt--fields--validations"></a>
### Nested Schema for `fields.validations`

Optional:

- `asset_file_size` (Attributes) (see [below for nested schema](#nestedatt--fields--validations--asset_file_size))
- `enabled_marks` (List of String)
- `enabled_node_types` (List of String)
- `in` (List of String)
- `link_content_type` (List of String)
- `link_mimetype_group` (List of String)
- `message` (String) Defines the message that is shown to the user when the validation fails. It can be used to provide more information about the validation.
- `nodes` (Attributes) (see [below for nested schema](#nestedatt--fields--validations--nodes))
- `range` (Attributes) (see [below for nested schema](#nestedatt--fields--validations--range))
- `regexp` (Attributes) (see [below for nested schema](#nestedatt--fields--validations--regexp))
- `size` (Attributes) (see [below for nested schema](#nestedatt--fields--validations--size))
- `unique` (Boolean)

<a id="nestedatt--fields--validations--asset_file_size"></a>
### Nested Schema for `fields.validations.asset_file_size`

Optional:

- `max` (Number)
- `min` (Number)


<a id="nestedatt--fields--validations--nodes"></a>
### Nested Schema for `fields.validations.nodes`

Optional:

- `asset_hyperlink` (Attributes List) (see [below for nested schema](#nestedatt--fields--validations--nodes--asset_hyperlink))
- `embedded_asset_block` (Attributes List) (see [below for nested schema](#nestedatt--fields--validations--nodes--embedded_asset_block))
- `embedded_entry_block` (Attributes List) (see [below for nested schema](#nestedatt--fields--validations--nodes--embedded_entry_block))
- `embedded_entry_inline` (Attributes List) (see [below for nested schema](#nestedatt--fields--validations--nodes--embedded_entry_inline))
- `embedded_resource_block` (Attributes) (see [below for nested schema](#nestedatt--fields--validations--nodes--embedded_resource_block))
- `embedded_resource_inline` (Attributes) (see [below for nested schema](#nestedatt--fields--validations--nodes--embedded_resource_inline))
- `entry_hyperlink` (Attributes List) (see [below for nested schema](#nestedatt--fields--validations--nodes--entry_hyperlink))
- `resource_hyperlink` (Attributes) (see [below for nested schema](#nestedatt--fields--validations--nodes--resource_hyperlink))

<a id="nestedatt--fields--validations--nodes--asset_hyperlink"></a>
### Nested Schema for `fields.validations.nodes.asset_hyperlink`

Optional:

- `message` (String) Defines the message that is shown to the user when the validation fails. It can be used to provide more information about the validation.
- `size` (Attributes) (see [below for nested schema](#nestedatt--fields--validations--nodes--resource_hyperlink--size))

<a id="nestedatt--fields--validations--nodes--resource_hyperlink--size"></a>
### Nested Schema for `fields.validations.nodes.resource_hyperlink.size`

Optional:

- `max` (Number)
- `min` (Number)



<a id="nestedatt--fields--validations--nodes--embedded_asset_block"></a>
### Nested Schema for `fields.validations.nodes.embedded_asset_block`

Optional:

- `message` (String) Defines the message that is shown to the user when the validation fails. It can be used to provide more information about the validation.
- `size` (Attributes) (see [below for nested schema](#nestedatt--fields--validations--nodes--resource_hyperlink--size))

<a id="nestedatt--fields--validations--nodes--resource_hyperlink--size"></a>
### Nested Schema for `fields.validations.nodes.resource_hyperlink.size`

Optional:

- `max` (Number)
- `min` (Number)



<a id="nestedatt--fields--validations--nodes--embedded_entry_block"></a>
### Nested Schema for `fields.validations.nodes.embedded_entry_block`

Optional:

- `link_content_type` (List of String)
- `message` (String) Defines the message that is shown to the user when the validation fails. It can be used to provide more information about the validation.
- `size` (Attributes) (see [below for nested schema](#nestedatt--fields--validations--nodes--resource_hyperlink--size))

<a id="nestedatt--fields--validations--nodes--resource_hyperlink--size"></a>
### Nested Schema for `fields.validations.nodes.resource_hyperlink.size`

Optional:

- `max` (Number)
- `min` (Number)



<a id="nestedatt--fields--validations--nodes--embedded_entry_inline"></a>
### Nested Schema for `fields.validations.nodes.embedded_entry_inline`

Optional:

- `link_content_type` (List of String)
- `message` (String) Defines the message that is shown to the user when the validation fails. It can be used to provide more information about the validation.
- `size` (Attributes) (see [below for nested schema](#nestedatt--fields--validations--nodes--resource_hyperlink--size))

<a id="nestedatt--fields--validations--nodes--resource_hyperlink--size"></a>
### Nested Schema for `fields.validations.nodes.resource_hyperlink.size`

Optional:

- `max` (Number)
- `min` (Number)



<a id="nestedatt--fields--validations--nodes--embedded_resource_block"></a>
### Nested Schema for `fields.validations.nodes.embedded_resource_block`

Optional:

- `allowed_resources` (Attributes List) Defines the entities that can be referenced by the field. It is only used for cross-space references. (see [below for nested schema](#nestedatt--fields--validations--nodes--resource_hyperlink--allowed_resources))
- `validations` (Attributes List) (see [below for nested schema](#nestedatt--fields--validations--nodes--resource_hyperlink--validations))

<a id="nestedatt--fields--validations--nodes--resource_hyperlink--allowed_resources"></a>
### Nested Schema for `fields.validations.nodes.resource_hyperlink.allowed_resources`

Optional:

- `content_types` (List of String)
- `source` (String)
- `type` (String)


<a id="nestedatt--fields--validations--nodes--resource_hyperlink--validations"></a>
### Nested Schema for `fields.validations.nodes.resource_hyperlink.validations`

Optional:

- `message` (String) Defines the message that is shown to the user when the validation fails. It can be used to provide more information about the validation.
- `size` (Attributes) (see [below for nested schema](#nestedatt--fields--validations--nodes--resource_hyperlink--validations--size))

<a id="nestedatt--fields--validations--nodes--resource_hyperlink--validations--size"></a>
### Nested Schema for `fields.validations.nodes.resource_hyperlink.validations.size`

Optional:

- `max` (Number)
- `min` (Number)




<a id="nestedatt--fields--validations--nodes--embedded_resource_inline"></a>
### Nested Schema for `fields.validations.nodes.embedded_resource_inline`

Optional:

- `allowed_resources` (Attributes List) Defines the entities that can be referenced by the field. It is only used for cross-space references. (see [below for nested schema](#nestedatt--fields--validations--nodes--resource_hyperlink--allowed_resources))
- `validations` (Attributes List) (see [below for nested schema](#nestedatt--fields--validations--nodes--resource_hyperlink--validations))

<a id="nestedatt--fields--validations--nodes--resource_hyperlink--allowed_resources"></a>
### Nested Schema for `fields.validations.nodes.resource_hyperlink.allowed_resources`

Optional:

- `content_types` (List of String)
- `source` (String)
- `type` (String)


<a id="nestedatt--fields--validations--nodes--resource_hyperlink--validations"></a>
### Nested Schema for `fields.validations.nodes.resource_hyperlink.validations`

Optional:

- `message` (String) Defines the message that is shown to the user when the validation fails. It can be used to provide more information about the validation.
- `size` (Attributes) (see [below for nested schema](#nestedatt--fields--validations--nodes--resource_hyperlink--validations--size))

<a id="nestedatt--fields--validations--nodes--resource_hyperlink--validations--size"></a>
### Nested Schema for `fields.validations.nodes.resource_hyperlink.validations.size`

Optional:

- `max` (Number)
- `min` (Number)




<a id="nestedatt--fields--validations--nodes--entry_hyperlink"></a>
### Nested Schema for `fields.validations.nodes.entry_hyperlink`

Optional:

- `link_content_type` (List of String)
- `message` (String) Defines the message that is shown to the user when the validation fails. It can be used to provide more information about the validation.
- `size` (Attributes) (see [below for nested schema](#nestedatt--fields--validations--nodes--resource_hyperlink--size))

<a id="nestedatt--fields--validations--nodes--resource_hyperlink--size"></a>
### Nested Schema for `fields.validations.nodes.resource_hyperlink.size`

Optional:

- `max` (Number)
- `min` (Number)



<a id="nestedatt--fields--validations--nodes--resource_hyperlink"></a>
### Nested Schema for `fields.validations.nodes.resource_hyperlink`

Optional:

- `allowed_resources` (Attributes List) Defines the entities that can be referenced by the field. It is only used for cross-space references. (see [below for nested schema](#nestedatt--fields--validations--nodes--resource_hyperlink--allowed_resources))
- `validations` (Attributes List) (see [below for nested schema](#nestedatt--fields--validations--nodes--resource_hyperlink--validations))

<a id="nestedatt--fields--validations--nodes--resource_hyperlink--allowed_resources"></a>
### Nested Schema for `fields.validations.nodes.resource_hyperlink.allowed_resources`

Optional:

- `content_types` (List of String)
- `source` (String)
- `type` (String)


<a id="nestedatt--fields--validations--nodes--resource_hyperlink--validations"></a>
### Nested Schema for `fields.validations.nodes.resource_hyperlink.validations`

Optional:

- `message` (String) Defines the message that is shown to the user when the validation fails. It can be used to provide more information about the validation.
- `size` (Attributes) (see [below for nested schema](#nestedatt--fields--validations--nodes--resource_hyperlink--validations--size))

<a id="nestedatt--fields--validations--nodes--resource_hyperlink--validations--size"></a>
### Nested Schema for `fields.validations.nodes.resource_hyperlink.validations.size`

Optional:

- `max` (Number)
- `min` (Number)





<a id="nestedatt--fields--validations--range"></a>
### Nested Schema for `fields.validations.range`

Optional:

- `max` (Number)
- `min` (Number)


<a id="nestedatt--fields--validations--regexp"></a>
### Nested Schema for `fields.validations.regexp`

Optional:

- `pattern` (String)


<a id="nestedatt--fields--validations--size"></a>
### Nested Schema for `fields.validations.size`

Optional:

- `max` (Number)
- `min` (Number)

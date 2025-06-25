resource "contentful_role" "example_role" {
  space_id = "space-id"

  name        = "custom-role-name"
  description = "Custom Role Description"

  permission {
    id     = "ContentModel"
    values = ["read"]
  }

  permission {
    id    = "ContentDelivery"
    value = "all"
  }

  permission {
    id    = "Environments"
    value = "all"
  }

  policy {
    effect = "allow"
    actions = {
      value = "all"
    }

    constraint = jsonencode({
      and = [
        [
          "equals",
            { doc = "sys.type" },
            "Entry"
        ]
      ]
    })
  }

  policy {
    effect = "allow"

    actions = {
      values = ["create"]
    }

    constraint = jsonencode({
      and = [
        [
          "equals",
            { doc = "sys.type" },
            "Entry"
        ]
      ]
    })
  }
}


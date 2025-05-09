resource "contentful_role" "example_role" {
  space_id = "space-id"

  name        = "custom-role-name"
  description = "Custom Role Description"

  permissions = {
    ContentModel    = ["read", "delete", "publish"]
    ContentDelivery = "all"
    Environments    = "all"
  }

  policies = [
    {
      effect = "allow"
      actions = [
        "read",
        "create",
        "update",
        "delete",
        "publish",
        "unpublish",
        "archive",
        "unarchive",
      ]
      constraint = {
        and = [
          [
            "equals",
            {
              doc = "sys.type"
            },
            "Entry"
          ]
        ]
      }
    }
  ]
}


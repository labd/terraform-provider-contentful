package role_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/assert"

	"github.com/labd/terraform-provider-contentful/internal/acctest"
	"github.com/labd/terraform-provider-contentful/internal/provider"
	"github.com/labd/terraform-provider-contentful/internal/sdk"
)

func TestRoleResource_Basic(t *testing.T) {
	spaceID := os.Getenv("CONTENTFUL_SPACE_ID")
	resourceName := "contentful_role.example_role"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.TestAccPreCheck(t) },
		CheckDestroy: testAccCheckContentfulRoleDestroy,
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"contentful": providerserver.NewProtocol6WithError(provider.New("test", true)()),
		},
		Steps: []resource.TestStep{
			{
				Config: testEntryConfig(spaceID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContentfulRoleExists(t, resourceName, func(t *testing.T, role *sdk.Role) {
						assert.Equal(t, "[automated] Custom Role", role.Name)
						assert.Equal(t, spaceID, role.Sys.Space.Sys.Id)
					}),
				),
			},
			{
				Config: testEntryUpdateConfig(spaceID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContentfulRoleExists(t, resourceName, func(t *testing.T, role *sdk.Role) {
						assert.Equal(t, "custom-role-name", role.Name)
						assert.Equal(t, spaceID, role.Sys.Space.Sys.Id)
					}),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"role_id"},
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources[resourceName]
					if !ok {
						return "", fmt.Errorf("not found: %s", resourceName)
					}
					return fmt.Sprintf("%s:%s",
						rs.Primary.Attributes["role_id"],
						rs.Primary.Attributes["space_id"],
					), nil
				},
			},
		},
	})
}

type assertFunc func(*testing.T, *sdk.Role)

func testAccCheckContentfulRoleExists(t *testing.T, resourceName string, assertFunc assertFunc) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		role, err := getRoleFromState(s, resourceName)
		if err != nil {
			return err
		}

		assertFunc(t, role)
		return nil
	}
}

func getRoleFromState(s *terraform.State, resourceName string) (*sdk.Role, error) {
	rs, ok := s.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("entry not found in state: %s", resourceName)
	}

	if rs.Primary.ID == "" {
		return nil, fmt.Errorf("no role ID found")
	}

	spaceID := rs.Primary.Attributes["space_id"]
	if spaceID == "" {
		return nil, fmt.Errorf("no space_id is set")
	}

	fmt.Println(rs.Primary.ID)

	client := acctest.GetClient()
	resp, err := client.GetRoleWithResponse(context.Background(), spaceID, rs.Primary.ID)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("entry not found: %s", rs.Primary.ID)
	}

	return resp.JSON200, nil
}

func testAccCheckContentfulRoleDestroy(s *terraform.State) error {
	client := acctest.GetClient()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "contentful_role" {
			continue
		}

		spaceID := rs.Primary.Attributes["space_id"]
		if spaceID == "" {
			return fmt.Errorf("no space_id is set")
		}

		resp, err := client.GetRoleWithResponse(context.Background(), spaceID, rs.Primary.ID)
		if err != nil {
			return err
		}

		if resp.StatusCode() == 404 {
			return nil
		}

		return fmt.Errorf("entry still exists with id: %s", rs.Primary.ID)
	}

	return nil
}

func testEntryConfig(spaceID string) string {
	return fmt.Sprintf(`
resource "contentful_role" "example_role" {
  space_id = "%s"

  name        = "[automated] Custom Role"
  description = "Custom Role Description"

  permission {
    id     = "ContentModel"
    values = ["read"]
  }

  policy {
    effect = "allow"
    actions = {
    	values = [
			"read",
			"create",
			"update",
			"delete",
			"publish",
			"unpublish",
			"archive",
			"unarchive"
		]
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
`, spaceID)
}

func testEntryUpdateConfig(spaceID string) string {
	return fmt.Sprintf(`
resource "contentful_role" "example_role" {
  space_id = "%s"


  name        = "custom-role-name"
  description = "Custom Role Description"

  permission {
    id     = "ContentModel"
    values = ["read"]
  }

  policy {
    effect = "allow"
    actions = {
    	values = [
			"read",
			"create",
			"update",
			"delete",
			"publish",
			"unpublish",
			"archive",
			"unarchive"
		]
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
`, spaceID)
}

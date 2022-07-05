package drone

import (
	"fmt"
	"testing"

	"github.com/drone/drone-go/drone"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDroneOrgsecretBasic(t *testing.T) {
	// generate a random name to avoid collisions from multiple concurrent tests.
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDroneOrgsecretDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDroneOrgsecretConfigBasic(
					"test",
					rName,
					"thisissecret",
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDroneOrgsecretExists("drone_orgsecret.secret"),
					resource.TestCheckResourceAttr(
						"drone_orgsecret.secret",
						"name",
						rName,
					),
					resource.TestCheckResourceAttr(
						"drone_orgsecret.secret",
						"namespace",
						"test",
					),
					resource.TestCheckResourceAttr(
						"drone_orgsecret.secret",
						"value",
						"thisissecret",
					),
				),
			},
		},
	})
}

func testAccCheckDroneOrgsecretDestroy(s *terraform.State) error {
	c := testAccProvider.Meta().(drone.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "drone_orgsecret" {
			continue
		}

		name := rs.Primary.Attributes["name"]
		namespace := rs.Primary.Attributes["namespace"]

		err := c.OrgSecretDelete(namespace, name)
		if err == nil {
			return fmt.Errorf("Organization secret (%s/%s) still exists.", namespace, name)
		}
	}

	return nil
}

func testAccCheckDroneOrgsecretConfigBasic(namespace, name, value string) string {
	return fmt.Sprintf(`
	resource "drone_orgsecret" "secret" {
		namespace = "%s"
		name      = "%s"
		value     = "%s"
	}
	`,
		namespace,
		name,
		value,
	)
}

func testAccCheckDroneOrgsecretExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID set")
		}

		return nil
	}
}

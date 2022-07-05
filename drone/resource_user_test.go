package drone

import (
	"fmt"
	"testing"

	"github.com/drone/drone-go/drone"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDroneUserBasic(t *testing.T) {
	// generate a random name to avoid collisions from multiple concurrent tests.
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDroneUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDroneUserConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDroneUserExists("drone_user.user"),
					resource.TestCheckResourceAttr(
						"drone_user.user",
						"login",
						rName,
					),
				),
			},
		},
	})
}

func testAccCheckDroneUserDestroy(s *terraform.State) error {
	c := testAccProvider.Meta().(drone.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "drone_user" {
			continue
		}

		login := rs.Primary.Attributes["login"]

		err := c.UserDelete(login)
		if err == nil {
			return fmt.Errorf("User (%s) still exists.", login)
		}
	}

	return nil
}

func testAccCheckDroneUserConfigBasic(n string) string {
	return fmt.Sprintf(`
	resource "drone_user" "user" {
		login = "%s"
		active = true
		admin = false
	}
	`, n)
}

func testAccCheckDroneUserExists(n string) resource.TestCheckFunc {
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

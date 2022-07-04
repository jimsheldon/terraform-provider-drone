package drone

import (
	"fmt"
	"testing"

	"github.com/drone/drone-go/drone"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDroneTemplateBasic(t *testing.T) {
	// generate a random name to avoid collisions from multiple concurrent tests.
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDroneTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDroneTemplateConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDroneTemplateExists("drone_template.new"),
					resource.TestCheckResourceAttr("drone_template.new", "name", rName+".yaml"),
				),
			},
		},
	})
}

func testAccCheckDroneTemplateDestroy(s *terraform.State) error {
	c := testAccProvider.Meta().(drone.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "drone_template" {
			continue
		}

		namespace := rs.Primary.Attributes["namespace"]
		name := rs.Primary.Attributes["name"]

		err := c.TemplateDelete(namespace, name)
		if err == nil {
			return fmt.Errorf("Template (%s/%s) still exists.", namespace, name)
		}
	}

	return nil
}

func testAccCheckDroneTemplateConfigBasic(n string) string {
	return fmt.Sprintf(`
	resource "drone_template" "new" {
		name = "%s.yaml"
		namespace = "test"
		data = "kind: pipeline"
	}
	`, n)
}

func testAccCheckDroneTemplateExists(n string) resource.TestCheckFunc {
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

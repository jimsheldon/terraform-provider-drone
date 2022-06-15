package drone

import (
	"fmt"
	"testing"

	"github.com/drone/drone-go/drone"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDroneTemplateBasic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDroneTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDroneTemplateConfigBasic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDroneTemplateExists("drone_template.new"),
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

		_, err := c.Template(namespace, name)
		if err == nil {
			return fmt.Errorf("Template (%s/%s) still exists.", namespace, name)
		}
	}

	return nil
}

func testAccCheckDroneTemplateConfigBasic() string {
	return fmt.Sprintf(`
	resource "drone_template" "new" {
		name = "test.yaml"
		namespace = "jimsheldon"
		data = "kind: pipeline"
	}
	`)
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

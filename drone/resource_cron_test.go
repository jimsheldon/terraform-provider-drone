package drone

import (
	"fmt"
	"os"
	"testing"

	"terraform-provider-drone/drone/utils"

	"github.com/drone/drone-go/drone"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDroneCronBasic(t *testing.T) {
	scmAvail := os.Getenv("SCM_AVAIL")
	if scmAvail == "" {
		t.Skip("set SCM_AVAIL to run this test")
	}

	// generate a random name to avoid collisions from multiple concurrent tests.
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDroneCronDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDroneCronConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDroneCronExists("drone_cron.new"),
					resource.TestCheckResourceAttr("drone_cron.new", "name", rName),
				),
			},
		},
	})
}

func testAccCheckDroneCronDestroy(s *terraform.State) error {
	c := testAccProvider.Meta().(drone.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "drone_cron" {
			continue
		}

		cronName := rs.Primary.Attributes["name"]
		repository := rs.Primary.Attributes["repository"]
		owner, repo, err := utils.ParseRepo(repository)

		err = c.CronDelete(owner, repo, cronName)
		if err == nil {
			return fmt.Errorf("Cron (%s/%s/%s) still exists.", owner, repo, cronName)
		}
	}

	return nil
}

func testAccCheckDroneCronConfigBasic(n string) string {
	return fmt.Sprintf(`
	resource "drone_cron" "new" {
		repository = "jimsheldon/drone-quickstart"
		name = "%s"
		event = "push"
	}
	`, n)
}

func testAccCheckDroneCronExists(n string) resource.TestCheckFunc {
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

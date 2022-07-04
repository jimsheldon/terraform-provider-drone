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

func TestAccDroneRepoBasic(t *testing.T) {
	// testing requires a valid repository, currently I only have this working
	// in my own local environment
	scmAvail := os.Getenv("SCM_AVAIL")
	if scmAvail == "" {
		t.Skip("set SCM_AVAIL to run this test")
	}

	// generate a random name to avoid collisions from multiple concurrent tests.
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDroneRepoDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDroneRepoConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDroneRepoExists("drone_repo.new"),
					resource.TestCheckResourceAttr("drone_repo.new", "configuration", rName+".yaml"),
				),
			},
		},
	})
}

func testAccCheckDroneRepoDestroy(s *terraform.State) error {
	c := testAccProvider.Meta().(drone.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "drone_repo" {
			continue
		}

		repository := rs.Primary.Attributes["repository"]
		owner, repo, err := utils.ParseRepo(repository)

		repositories, err := c.RepoList()

		for _, repository := range repositories {
			if (repository.Namespace == owner) && (repository.Name == repo) {
				err = c.RepoDisable(owner, repo)
				if err != nil {
					return fmt.Errorf("Repo still exists: %s/%s", owner, repo)
				}
			}
		}
	}

	return nil
}

func testAccCheckDroneRepoConfigBasic(n string) string {
	return fmt.Sprintf(`
	resource "drone_repo" "new" {
		repository = "jimsheldon/drone-quickstart"
		configuration = "%s.yaml"
	}
	`, n)
}

func testAccCheckDroneRepoExists(n string) resource.TestCheckFunc {
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

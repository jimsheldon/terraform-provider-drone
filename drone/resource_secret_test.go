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

func TestAccDroneSecretBasic(t *testing.T) {
	// testing secrets requires a valid repository, currently I only have this working
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
		CheckDestroy: testAccCheckDroneSecretDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDroneSecretConfigBasic(
					testDroneUser,
					"repository-1",
					rName,
					"thisissecret",
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDroneSecretExists("drone_secret.secret"),
					resource.TestCheckResourceAttr(
						"drone_secret.secret",
						"repository",
						fmt.Sprintf("%s/repository-1", testDroneUser),
					),
					resource.TestCheckResourceAttr(
						"drone_secret.secret",
						"name",
						rName,
					),
					resource.TestCheckResourceAttr(
						"drone_secret.secret",
						"value",
						"thisissecret",
					),
				),
			},
		},
	})
}

func testAccCheckDroneSecretDestroy(s *terraform.State) error {
	c := testAccProvider.Meta().(drone.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "drone_secret" {
			continue
		}

		name := rs.Primary.Attributes["name"]
		repository := rs.Primary.Attributes["repository"]
		owner, repo, err := utils.ParseRepo(repository)

		err = c.SecretDelete(owner, repo, name)
		if err == nil {
			return fmt.Errorf("Secret (%s/%s/%s) still exists.", owner, repo, name)
		}
	}

	return nil
}

func testAccCheckDroneSecretConfigBasic(user, repo, name, value string) string {
	return fmt.Sprintf(`
    resource "drone_repo" "repo" {
		repository = "%s/%s"
	  }
	  
	  resource "drone_secret" "secret" {
		repository = drone_repo.repo.repository
		name       = "%s"
		value      = "%s"
	  }
	  `,
		user,
		repo,
		name,
		value,
	)
}

func testAccCheckDroneSecretExists(n string) resource.TestCheckFunc {
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

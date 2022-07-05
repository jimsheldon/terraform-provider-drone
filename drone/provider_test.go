package drone

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var (
	testDroneUser    string = os.Getenv("DRONE_USER")
	testAccProviders map[string]*schema.Provider
	testAccProvider  *schema.Provider
)

func init() {
	testAccProvider = Provider()
	testAccProviders = map[string]*schema.Provider{
		"drone": testAccProvider,
	}
}

func TestProvider(t *testing.T) {
	if err := Provider().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvider_impl(t *testing.T) {
	var _ *schema.Provider = Provider()
}

func testAccPreCheck(t *testing.T) {
	if err := os.Getenv("DRONE_SERVER"); err == "" {
		t.Fatal("DRONE_SERVER must be set for acceptance tests")
	}
	if err := os.Getenv("DRONE_TOKEN"); err == "" {
		t.Fatal("DRONE_TOKEN must be set for acceptance tests")
	}
	if v := os.Getenv("DRONE_USER"); v == "" {
		t.Fatal("DRONE_USER must be set for acceptance tests")
	}
}

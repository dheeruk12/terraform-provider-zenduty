package zenduty

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var testAccProviders map[string]*schema.Provider
var testAccProvider *schema.Provider

func init() {
	testAccProvider = Provider()
	testAccProviders = map[string]*schema.Provider{
		"zenduty": testAccProvider,
	}
}

func TestProvider(t *testing.T) {
	if err := Provider().InternalValidate(); err != nil {
		t.Fatalf("provider validation failed: %s", err)
	}
}

// testAccPreCheck guards acceptance tests, which talk to the live API.
func testAccPreCheck(t *testing.T) {
	t.Helper()
	if os.Getenv("ZENDUTY_API_KEY") == "" {
		t.Fatal("ZENDUTY_API_KEY must be set for acceptance tests")
	}
}

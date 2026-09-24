package zendesk

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var testAccProviders map[string]*schema.Provider
var testAccProvider *schema.Provider

func init() {
	testAccProvider = Provider()
	testAccProviders = map[string]*schema.Provider{
		"zendesk": testAccProvider,
	}
}

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv(emailVar); v == "" {
		t.Fatalf("%s must be set for acceptance tests", emailVar)
	}
	if v := os.Getenv(tokenVar); v == "" {
		t.Fatalf("%s must be set for acceptance tests", tokenVar)
	}
	if v := os.Getenv(accountVar); v == "" {
		t.Fatalf("%s must be set for acceptance tests", accountVar)
	}
}

func TestProvider(t *testing.T) {
	if err := Provider().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProviderAcceptsOAuthWithoutAPITokenVariables(t *testing.T) {
	t.Setenv(emailVar, "")
	t.Setenv(tokenVar, "")
	config := terraform.NewResourceConfigRaw(map[string]interface{}{
		"account":             "example",
		"oauth_client_id":     "id",
		"oauth_client_secret": "secret",
	})

	diags := Provider().Validate(config)

	if diags.HasError() {
		t.Fatalf("an OAuth-only configuration failed validation: %v", diags)
	}
}

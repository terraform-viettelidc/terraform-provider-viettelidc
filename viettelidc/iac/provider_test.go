package iac

import (
	"context"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// TestCSAProviderInterface verifies the CSAProvider satisfies provider.Provider
// at runtime (the var _ assertion in provider.go covers compile-time).
func TestCSAProviderInterface(t *testing.T) {
	var _ provider.Provider = New()()
}

// TestCSAProviderMetadata verifies the provider TypeName matches the muxed
// VCD provider so both servers respond to "viettelidc" address.
func TestCSAProviderMetadata(t *testing.T) {
	p := New()()
	resp := &provider.MetadataResponse{}
	p.Metadata(context.Background(), provider.MetadataRequest{}, resp)

	if resp.TypeName != "viettelidc" {
		t.Errorf("TypeName = %q, want %q", resp.TypeName, "viettelidc")
	}
	if resp.Version == "" {
		t.Error("Version must not be empty")
	}
}

// TestCSAProviderResourcesRegistered asserts the expected number of
// resources and data sources are wired up.
func TestCSAProviderResourcesRegistered(t *testing.T) {
	p := New()()
	if got, want := len(p.Resources(context.Background())), 3; got != want {
		t.Errorf("Resources() len = %d, want %d", got, want)
	}
	if got, want := len(p.DataSources(context.Background())), 4; got != want {
		t.Errorf("DataSources() len = %d, want %d", got, want)
	}
}

// TestCSAProviderSchema asserts the four required attributes are declared
// with correct optionality and sensitivity.
func TestCSAProviderSchema(t *testing.T) {
	p := New()()
	resp := &provider.SchemaResponse{}
	p.Schema(context.Background(), provider.SchemaRequest{}, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("schema diagnostics: %v", resp.Diagnostics)
	}
	attrs := resp.Schema.Attributes
	for _, name := range []string{"base_url", "customer_id", "token", "vpc_id", "username", "atm_token", "atm_authentication_token"} {
		if _, ok := attrs[name]; !ok {
			t.Errorf("missing attribute %q", name)
		}
	}
	for _, name := range []string{"token", "atm_token", "atm_authentication_token"} {
		if a, ok := attrs[name]; ok && !a.IsSensitive() {
			t.Errorf("%s must be Sensitive", name)
		}
	}
}

func TestStringOrEnv_HCLWins(t *testing.T) {
	t.Setenv("VIETTELIDC_TEST_VAR", "from-env")
	got := stringOrEnv(types.StringValue("from-hcl"), "VIETTELIDC_TEST_VAR")
	if got != "from-hcl" {
		t.Errorf("got %q, want from-hcl", got)
	}
}

func TestStringOrEnv_NullFallsBackToEnv(t *testing.T) {
	t.Setenv("VIETTELIDC_TEST_VAR", "from-env")
	got := stringOrEnv(types.StringNull(), "VIETTELIDC_TEST_VAR")
	if got != "from-env" {
		t.Errorf("got %q, want from-env", got)
	}
}

func TestStringOrEnv_EmptyHCLFallsBackToEnv(t *testing.T) {
	t.Setenv("VIETTELIDC_TEST_VAR", "from-env")
	got := stringOrEnv(types.StringValue(""), "VIETTELIDC_TEST_VAR")
	if got != "from-env" {
		t.Errorf("empty HCL string should fall back to env; got %q", got)
	}
}

func TestStringOrEnv_BothEmpty(t *testing.T) {
	_ = os.Unsetenv("VIETTELIDC_TEST_UNSET")
	got := stringOrEnv(types.StringNull(), "VIETTELIDC_TEST_UNSET")
	if got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}

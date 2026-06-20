// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

const (
	// providerConfig is a shared configuration to combine with the actual
	// test configuration so the HashiCups client is properly configured.
	// It is also possible to use the HASHICUPS_ environment variables instead,
	// such as updating the Makefile and running the testing through that tool.
	providerConfig = `
provider "wordpress" {
	host = "http://localhost:8888/wp-json/wp/v2"
	username = "admin"
}
`
)

// testAccProtoV6ProviderFactories are used to instantiate a provider during
// acceptance testing. The factory function will be invoked for every Terraform
// CLI command executed to create a provider server to which the CLI can
// reattach.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"wordpress": providerserver.NewProtocol6WithError(New("test")()),
}

func TestConfigValuePrefersExplicitConfig(t *testing.T) {
	t.Setenv("WP_TF_PROVIDER_HOST", "http://env.example/wp-json/wp/v2")
	t.Setenv("WORDPRESS_HOST", "http://legacy.example/wp-json/wp/v2")

	value := configValue(types.StringValue("http://config.example/wp-json/wp/v2"), "WP_TF_PROVIDER_HOST", "WORDPRESS_HOST")

	if value != "http://config.example/wp-json/wp/v2" {
		t.Fatalf("expected config value to win, got %q", value)
	}
}

func TestConfigValueUsesProviderEnvironmentVariables(t *testing.T) {
	t.Setenv("WP_TF_PROVIDER_PASSWORD", "env-password")
	t.Setenv("WORDPRESS_PASSWORD", "legacy-password")

	value := configValue(types.String{}, "WP_TF_PROVIDER_PASSWORD", "WORDPRESS_PASSWORD")

	if value != "env-password" {
		t.Fatalf("expected provider env var to win, got %q", value)
	}
}

func TestConfigValueFallsBackToLegacyEnvironmentVariables(t *testing.T) {
	t.Setenv("WORDPRESS_USERNAME", "legacy-user")

	value := configValue(types.String{}, "WP_TF_PROVIDER_USERNAME", "WORDPRESS_USERNAME")

	if value != "legacy-user" {
		t.Fatalf("expected legacy env var to be used, got %q", value)
	}
}

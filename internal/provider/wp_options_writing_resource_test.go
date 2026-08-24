package provider

import (
	"context"
	"fmt"
	"os"
	"testing"

	"terraform-provider-wordpress/internal/wpappauth"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccWPOptionsWritingResource(t *testing.T) {
	if os.Getenv("TF_ACC") != "1" {
		t.Skip("TF_ACC=1 must be set for acceptance tests")
	}
	password := os.Getenv("WP_TF_PROVIDER_USER_PASSWORD")
	if password == "" {
		password = os.Getenv("WORDPRESS_USER_PASSWORD")
	}
	if password == "" {
		t.Skip("WP_TF_PROVIDER_USER_PASSWORD or WORDPRESS_USER_PASSWORD must be set")
	}

	service := &wpappauth.Service{
		BaseURL:  "http://localhost:8888/wp-json/wp/v2",
		Username: "admin",
		Password: password,
	}
	original, err := service.GetWritingOptions(context.Background())
	if err != nil {
		t.Fatalf("unable to read original writing options: %v", err)
	}
	t.Cleanup(func() {
		if err := service.UpdateWritingOptions(context.Background(), *original); err != nil {
			t.Errorf("unable to restore writing options: %v", err)
		}
	})

	const (
		defaultCategory      = 1
		defaultPostFormat    = "0"
		mailserverURL        = "mail.example.com"
		mailserverPort       = 110
		mailserverLogin      = "login@example.com"
		defaultEmailCategory = 1
		pingSites            = "https://example.com/ping"
	)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{{
			Config: fmt.Sprintf(`provider "wordpress" {
	host = "http://localhost:8888/wp-json/wp/v2"
	user_auth {
		username = "admin"
		password = %q
	}
}

resource "wordpress_wp_options_writing" "test" {
	default_category       = %d
	default_post_format    = %q
	mailserver_url         = %q
	mailserver_port        = %d
	mailserver_login       = %q
	mailserver_pass        = ""
	default_email_category = %d
	ping_sites             = %q
}
`, password, defaultCategory, defaultPostFormat, mailserverURL, mailserverPort, mailserverLogin, defaultEmailCategory, pingSites),
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("wordpress_wp_options_writing.test", "id", "writing"),
				resource.TestCheckResourceAttr("wordpress_wp_options_writing.test", "default_category", "1"),
				resource.TestCheckResourceAttr("wordpress_wp_options_writing.test", "mailserver_port", "110"),
				func(_ *terraform.State) error {
					options, err := service.GetWritingOptions(context.Background())
					if err != nil {
						return fmt.Errorf("GetWritingOptions failed: %w", err)
					}
					if options.DefaultCategory != defaultCategory || options.DefaultPostFormat != defaultPostFormat || options.MailserverURL != mailserverURL || options.MailserverPort != mailserverPort || options.MailserverLogin != mailserverLogin || options.DefaultEmailCategory != defaultEmailCategory || options.PingSites != pingSites {
						return fmt.Errorf("unexpected writing options: %#v", options)
					}
					return nil
				},
			),
		}},
	})
}

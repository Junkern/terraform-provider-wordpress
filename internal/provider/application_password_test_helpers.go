package provider

import (
	"context"
	"fmt"
	"os"
	"testing"

	"terraform-provider-wordpress/internal/wpapi"
	"terraform-provider-wordpress/internal/wpappauth"
)

const applicationPasswordTestUserID = 1

func testAccApplicationPasswordPassword(t *testing.T) string {
	t.Helper()

	if value, ok := os.LookupEnv("WP_TF_PROVIDER_PASSWORD"); ok && value != "" {
		return value
	}

	if value, ok := os.LookupEnv("WORDPRESS_PASSWORD"); ok && value != "" {
		return value
	}

	t.Skip("WP_TF_PROVIDER_PASSWORD or WORDPRESS_PASSWORD must be set for application password acceptance tests")
	return ""
}

func testAccApplicationPasswordProviderConfig(t *testing.T) string {
	t.Helper()

	password := testAccApplicationPasswordPassword(t)
	return fmt.Sprintf(`provider "wordpress" {
	host = "http://localhost:8888/wp-json/wp/v2"
	username = "admin"
	password = %q
}
`, password)
}

func testAccApplicationPasswordClient(t *testing.T) *wpapi.Client {
	t.Helper()

	client, err := wpapi.New("http://localhost:8888/wp-json/wp/v2", "admin", testAccApplicationPasswordPassword(t))
	if err != nil {
		t.Fatalf("unable to create wpapi client: %v", err)
	}

	return client
}

func testAccSeedApplicationPassword(t *testing.T, name string) *wpapi.ApplicationPassword {
	t.Helper()

	client := testAccApplicationPasswordClient(t)
	service := &wpappauth.Service{
		BaseURL:  "http://localhost:8888/wp-json/wp/v2",
		Username: "admin",
		Password: testAccApplicationPasswordPassword(t),
	}
	ctx := context.Background()
	createdResult, err := service.CreateApplicationPassword(ctx, name)
	if err != nil {
		t.Fatalf("unable to seed application password %q: %v", name, err)
	}
	created := &wpapi.ApplicationPassword{UUID: createdResult.UUID, AppID: createdResult.AppID, Name: createdResult.Name, Password: createdResult.Password}

	t.Cleanup(func() {
		if deleteErr := client.DeleteApplicationPassword(ctx, applicationPasswordTestUserID, created.UUID); deleteErr != nil {
			t.Logf("cleanup delete failed for %s: %v", created.UUID, deleteErr)
		}
	})

	return created
}

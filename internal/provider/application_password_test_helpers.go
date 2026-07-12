package provider

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"terraform-provider-wordpress/internal/wpapi"
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
	appID := fmt.Sprintf("terraform-%d", time.Now().UnixNano())
	ctx := context.Background()
	created, err := client.CreateApplicationPassword(ctx, applicationPasswordTestUserID, wpapi.ApplicationPasswordInput{
		Name:  stringPtrTest(name),
		AppID: stringPtrTest(appID),
	})
	if err != nil {
		t.Fatalf("unable to seed application password %q: %v", name, err)
	}

	t.Cleanup(func() {
		if deleteErr := client.DeleteApplicationPassword(ctx, applicationPasswordTestUserID, created.UUID); deleteErr != nil {
			t.Logf("cleanup delete failed for %s: %v", created.UUID, deleteErr)
		}
	})

	return created
}

func stringPtrTest(value string) *string {
	return &value
}

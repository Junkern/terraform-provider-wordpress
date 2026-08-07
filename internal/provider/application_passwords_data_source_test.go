package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccApplicationPasswordsDataSource(t *testing.T) {
	created := testAccSeedApplicationPassword(t, "terraform-list-application-password")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationPasswordProviderConfig(t) + `
data "wordpress_application_passwords" "test" {
	user_id = 1
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.wordpress_application_passwords.test", "application_passwords.#", "1"),
					resource.TestCheckResourceAttr("data.wordpress_application_passwords.test", "application_passwords.0.uuid", created.UUID),
					resource.TestCheckResourceAttr("data.wordpress_application_passwords.test", "application_passwords.0.name", "terraform-list-application-password"),
				),
			},
		},
	})
}

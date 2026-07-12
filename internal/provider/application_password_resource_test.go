package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccApplicationPasswordResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationPasswordProviderConfig(t) + `
resource "wordpress_application_password" "test" {
	user_id = 1
	name    = "terraform-application-password"
	app_id  = "terraform-application-password"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("wordpress_application_password.test", "user_id", "1"),
					resource.TestCheckResourceAttr("wordpress_application_password.test", "name", "terraform-application-password"),
					resource.TestCheckResourceAttrSet("wordpress_application_password.test", "uuid"),
					resource.TestCheckResourceAttrSet("wordpress_application_password.test", "password"),
				),
			},
			{
				Config: testAccApplicationPasswordProviderConfig(t) + `
resource "wordpress_application_password" "test" {
	user_id = 1
	name    = "terraform-application-password-updated"
	app_id  = "terraform-application-password"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("wordpress_application_password.test", "name", "terraform-application-password-updated"),
				),
			},
		},
	})
}

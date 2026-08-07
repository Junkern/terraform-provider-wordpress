package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccApplicationPasswordDataSource(t *testing.T) {
	created := testAccSeedApplicationPassword(t, "terraform-single-application-password")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationPasswordProviderConfig(t) + `
data "wordpress_application_password" "test" {
	user_id = 1
	uuid    = "` + created.UUID + `"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.wordpress_application_password.test", "uuid", created.UUID),
					resource.TestCheckResourceAttr("data.wordpress_application_password.test", "name", "terraform-single-application-password"),
					resource.TestCheckResourceAttr("data.wordpress_application_password.test", "app_id", created.AppID),
				),
			},
		},
	})
}

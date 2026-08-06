package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccThemeResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
provider "wordpress" {
	host = "http://localhost:8888/wp-json/wp/v2"
	user_auth {
		username = "admin"
	}
}

resource "wordpress_theme" "test" {
	slug = "astra"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("wordpress_theme.test", "slug", "astra"),
				),
			},
		},
	})
}

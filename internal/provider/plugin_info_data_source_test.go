package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccPluginInfoDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `
data "wordpress_plugin_info" "test" {
  slug = "woocommerce"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.wordpress_plugin_info.test", "slug", "woocommerce"),
					resource.TestCheckResourceAttrSet("data.wordpress_plugin_info.test", "name"),
					resource.TestCheckResourceAttrSet("data.wordpress_plugin_info.test", "version"),
					resource.TestCheckResourceAttrSet("data.wordpress_plugin_info.test", "download_link"),
				),
			},
		},
	})
}

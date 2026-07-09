package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccPluginsDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `data "wordpress_plugins" "test" {}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.wordpress_plugins.test", "plugins.#"),
					resource.TestCheckResourceAttr("data.wordpress_plugins.test", "plugins.#", "1"),
					resource.TestCheckResourceAttr("data.wordpress_plugins.test", "plugins.0.name", "Hello Dolly"),
				),
			},
		},
	})
}

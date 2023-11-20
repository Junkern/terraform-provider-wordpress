package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccUserResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
resource "wordpress_user" "test" {
	username = "terraform-user"
	email = "terraform-user@example.com"
	password = "foobar123!"

	name = "Terraform User"
	first_name = "Terraform"
	last_name = "User"
	roles = ["author"]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("wordpress_user.test", "username", "terraform-user"),
					resource.TestCheckResourceAttr("wordpress_user.test", "email", "terraform-user@example.com"),
					resource.TestCheckResourceAttr("wordpress_user.test", "name", "Terraform User"),
					resource.TestCheckResourceAttr("wordpress_user.test", "roles.#", "1"),
				),
			},
			{
				Config: providerConfig + `
resource "wordpress_user" "test" {
	username = "terraform-user"
	email = "terraform-user@example.com"
	password = "foobar123!"

	name = "Terraform User Updated"
	first_name = "Terraform"
	last_name = "User"
	roles = ["author"]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("wordpress_user.test", "name", "Terraform User Updated"),
					resource.TestCheckResourceAttr("wordpress_user.test", "username", "terraform-user"),
				),
			},
		},
	})
}

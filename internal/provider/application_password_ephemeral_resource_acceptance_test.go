package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

func TestAccApplicationPasswordEphemeralResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_10_0),
		},
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationPasswordProviderConfig(t) + `
ephemeral "wordpress_application_password_ephemeral" "test" {
	user_id = 1
	name    = "terraform-ephemeral-application-password"
	delete_on_close = true
}

output "ephemeral_application_password" {
	value     = ephemeral.wordpress_application_password_ephemeral.test.password
	sensitive = true
}
`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("ephemeral_application_password", knownvalue.NotNull()),
				},
			},
		},
	})
}

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestPluginResourceInputFromModelOmitsBlankStatus(t *testing.T) {
	input := pluginResourceInputFromModel(pluginResourceModel{
		Slug:   types.StringValue("hello-dolly"),
		Status: types.StringValue(""),
	})

	if input.Slug != "hello-dolly" {
		t.Fatalf("expected slug to be copied, got %q", input.Slug)
	}
	if input.Status != nil {
		t.Fatalf("expected status to be omitted, got %q", *input.Status)
	}
}

func TestPluginResourceInputFromModelKeepsExplicitStatus(t *testing.T) {
	input := pluginResourceInputFromModel(pluginResourceModel{
		Slug:   types.StringValue("hello-dolly"),
		Status: types.StringValue("active"),
	})

	if input.Status == nil || *input.Status != "active" {
		t.Fatalf("expected status to be active, got %#v", input.Status)
	}
}

func TestAccPluginResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `
resource "wordpress_plugin" "test" {
	slug = "hello-dolly"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("wordpress_plugin.test", "slug", "hello-dolly"),
				),
			},
		},
	})
}

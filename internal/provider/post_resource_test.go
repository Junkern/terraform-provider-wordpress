package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestPostResourceInputFromModelOmitsBlankCommentAndPingStatus(t *testing.T) {
	input := postResourceInputFromModel(postResourceModel{
		Comment_status: types.StringValue(""),
		Ping_status:    types.StringValue(""),
	})

	if input.CommentStatus != nil {
		t.Fatalf("expected comment_status to be omitted, got %q", *input.CommentStatus)
	}
	if input.PingStatus != nil {
		t.Fatalf("expected ping_status to be omitted, got %q", *input.PingStatus)
	}
}

func TestPostResourceInputFromModelKeepsExplicitStatuses(t *testing.T) {
	input := postResourceInputFromModel(postResourceModel{
		Comment_status: types.StringValue("open"),
		Ping_status:    types.StringValue("closed"),
	})

	if input.CommentStatus == nil || *input.CommentStatus != "open" {
		t.Fatalf("expected comment_status to be open, got %#v", input.CommentStatus)
	}
	if input.PingStatus == nil || *input.PingStatus != "closed" {
		t.Fatalf("expected ping_status to be closed, got %#v", input.PingStatus)
	}
}

func TestAccPostResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `
resource "wordpress_post" "test" {
	title = {
		raw = "foobar"
	}

	content =  {
		raw = "foobar"
	}
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("wordpress_post.test", "title.raw", "foobar"),
					resource.TestCheckResourceAttr("wordpress_post.test", "title.rendered", "foobar"),
					resource.TestCheckResourceAttr("wordpress_post.test", "content.raw", "foobar"),
					resource.TestCheckResourceAttr("wordpress_post.test", "content.rendered", "<p>foobar</p>\n"),
				),
			},
			{
				Config: providerConfig + `
resource "wordpress_post" "test" {
	title = {
		raw = "foobar2"
	}

	content =  {
		raw = "foobar2"
	}
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("wordpress_post.test", "title.raw", "foobar2"),
					resource.TestCheckResourceAttr("wordpress_post.test", "title.rendered", "foobar2"),
					resource.TestCheckResourceAttr("wordpress_post.test", "content.raw", "foobar2"),
					resource.TestCheckResourceAttr("wordpress_post.test", "content.rendered", "<p>foobar2</p>\n"),
				),
			},
		},
	})
}

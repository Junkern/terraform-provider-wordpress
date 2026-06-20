package provider

import (
	"context"
	"fmt"

	"terraform-provider-wordpress/internal/wpapi"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &postsDataSource{}
	_ datasource.DataSourceWithConfigure = &postsDataSource{}
)

type postsDataSource struct {
	client *wpapi.Client
}

type postsDataSourceModel struct {
	Posts []postModel `tfsdk:"posts"`
}

type postModel struct {
	ID             types.Int64            `tfsdk:"id"`
	Date           types.String           `tfsdk:"date"`
	Date_gmt       types.String           `tfsdk:"date_gmt"`
	Link           types.String           `tfsdk:"link"`
	Modified       types.String           `tfsdk:"modified"`
	Modified_gmt   types.String           `tfsdk:"modified_gmt"`
	Slug           types.String           `tfsdk:"slug"`
	Status         types.String           `tfsdk:"status"`
	Type           types.String           `tfsdk:"type"`
	Password       types.String           `tfsdk:"password"`
	Title          renderedModel          `tfsdk:"title"`
	Content        renderedProtectedModel `tfsdk:"content"`
	Author         types.Int64            `tfsdk:"author"`
	Excerpt        renderedProtectedModel `tfsdk:"excerpt"`
	Featured_media types.Int64            `tfsdk:"featured_media"`
	Comment_status types.String           `tfsdk:"comment_status"`
	Ping_status    types.String           `tfsdk:"ping_status"`
	Format         types.String           `tfsdk:"format"`
	Sticky         types.Bool             `tfsdk:"sticky"`
	Template       types.String           `tfsdk:"template"`
}

func NewPostsDataSource() datasource.DataSource {
	return &postsDataSource{}
}

func (d *postsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*wpapi.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *wpapi.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

func (d *postsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_posts"
}

func (d *postsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"posts": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":           schema.Int64Attribute{Computed: true},
						"date":         schema.StringAttribute{Computed: true},
						"date_gmt":     schema.StringAttribute{Computed: true},
						"link":         schema.StringAttribute{Computed: true},
						"modified":     schema.StringAttribute{Computed: true},
						"modified_gmt": schema.StringAttribute{Computed: true},
						"slug":         schema.StringAttribute{Computed: true},
						"status":       schema.StringAttribute{Computed: true},
						"type":         schema.StringAttribute{Computed: true},
						"password":     schema.StringAttribute{Computed: true},
						"title": schema.SingleNestedAttribute{
							Computed: true,
							Attributes: map[string]schema.Attribute{
								"rendered": schema.StringAttribute{Computed: true},
							},
						},
						"content": schema.SingleNestedAttribute{
							Computed: true,
							Attributes: map[string]schema.Attribute{
								"rendered":  schema.StringAttribute{Computed: true},
								"protected": schema.BoolAttribute{Computed: true},
							},
						},
						"author": schema.Int64Attribute{Computed: true},
						"excerpt": schema.SingleNestedAttribute{
							Computed: true,
							Attributes: map[string]schema.Attribute{
								"rendered":  schema.StringAttribute{Computed: true},
								"protected": schema.BoolAttribute{Computed: true},
							},
						},
						"featured_media": schema.Int64Attribute{Computed: true},
						"comment_status": schema.StringAttribute{Computed: true},
						"ping_status":    schema.StringAttribute{Computed: true},
						"format":         schema.StringAttribute{Computed: true},
						"sticky":         schema.BoolAttribute{Computed: true},
						"template":       schema.StringAttribute{Computed: true},
					},
				},
			},
		},
	}
}

func (d *postsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state postsDataSourceModel

	posts, err := d.client.ListPosts(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Posts from Wordpress instance",
			err.Error(),
		)
		return
	}

	for _, post := range posts {
		postsState := postModel{
			ID:           types.Int64Value(post.ID),
			Date:         types.StringValue(post.Date),
			Date_gmt:     types.StringValue(post.DateGMT),
			Link:         types.StringValue(post.Link),
			Modified:     types.StringValue(post.Modified),
			Modified_gmt: types.StringValue(post.ModifiedGMT),
			Slug:         types.StringValue(post.Slug),
			Status:       types.StringValue(post.Status),
			Type:         types.StringValue(post.Type),
			Password:     types.StringValue(post.Password),
			Title: renderedModel{
				Rendered: types.StringValue(post.Title.Rendered),
			},
			Content: renderedProtectedModel{
				Rendered:  types.StringValue(post.Content.Rendered),
				Protected: types.BoolValue(post.Content.Protected),
			},
			Author: types.Int64Value(post.Author),
			Excerpt: renderedProtectedModel{
				Rendered:  types.StringValue(post.Excerpt.Rendered),
				Protected: types.BoolValue(post.Excerpt.Protected),
			},
			Featured_media: types.Int64Value(post.FeaturedMedia),
			Comment_status: types.StringValue(post.CommentStatus),
			Ping_status:    types.StringValue(post.PingStatus),
			Format:         types.StringValue(post.Format),
			Sticky:         types.BoolValue(post.Sticky),
			Template:       types.StringValue(post.Template),
		}

		state.Posts = append(state.Posts, postsState)
	}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

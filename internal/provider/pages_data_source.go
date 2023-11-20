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
	_ datasource.DataSource              = &pagesDataSource{}
	_ datasource.DataSourceWithConfigure = &pagesDataSource{}
)

type pagesDataSource struct {
	client *wpapi.Client
}

type pagesDataSourceModel struct {
	Pages []pagesModel `tfsdk:"pages"`
}

type pagesModel struct {
	ID           types.Int64   `tfsdk:"id"`
	GUID         renderedModel `tfsdk:"guid"`
	Date         types.String  `tfsdk:"date"`
	Date_gmt     types.String  `tfsdk:"date_gmt"`
	Link         types.String  `tfsdk:"link"`
	Modified     types.String  `tfsdk:"modified"`
	Modified_gmt types.String  `tfsdk:"modified_gmt"`
	Slug         types.String  `tfsdk:"slug"`
	Status       types.String  `tfsdk:"status"`
	Type         types.String  `tfsdk:"type"`
	Password     types.String  `tfsdk:"password"`
	// TODO: not currently exposed by the data source model.
	// Permalink_template types.String `tfsdk:"permalink_template"`
	// TODO: not currently exposed by the data source model.
	// Generated_slug types.String `tfsdk:"generated_slug"`
	Parent         types.Int64            `tfsdk:"parent"`
	Author         types.Int64            `tfsdk:"author"`
	Featured_media types.Int64            `tfsdk:"featured_media"`
	Comment_status types.String           `tfsdk:"comment_status"`
	Ping_status    types.String           `tfsdk:"ping_status"`
	Menu_order     types.Int64            `tfsdk:"menu_order"`
	Template       types.String           `tfsdk:"template"`
	Title          renderedModel          `tfsdk:"title"`
	Content        renderedProtectedModel `tfsdk:"content"`
	Excerpt        renderedProtectedModel `tfsdk:"excerpt"`
	Meta           footnotesModel         `tfsdk:"meta"`
}

type renderedModel struct {
	Rendered types.String `tfsdk:"rendered"`
}

type renderedProtectedModel struct {
	Rendered  types.String `tfsdk:"rendered"`
	Protected types.Bool   `tfsdk:"protected"`
}

type footnotesModel struct {
	FootNotes types.String `tfsdk:"footnotes"`
}

func NewPagesDataSource() datasource.DataSource {
	return &pagesDataSource{}
}

func (d *pagesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *pagesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_pages"
}

// https://developer.wordpress.org/rest-api/reference/pages/
func (d *pagesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"pages": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							Computed: true,
						},
						"date": schema.StringAttribute{
							Computed: true,
						},
						"date_gmt": schema.StringAttribute{
							Computed: true,
						},
						"guid": schema.SingleNestedAttribute{
							Computed: true,
							Attributes: map[string]schema.Attribute{
								"rendered": schema.StringAttribute{
									Computed: true,
								},
							},
						},
						"link": schema.StringAttribute{
							Computed: true,
						},
						"modified": schema.StringAttribute{
							Computed: true,
						},
						"modified_gmt": schema.StringAttribute{
							Computed: true,
						},
						"slug": schema.StringAttribute{
							Computed: true,
						},
						"status": schema.StringAttribute{
							Computed: true,
						},
						"type": schema.StringAttribute{
							Computed: true,
						},
						"password": schema.StringAttribute{
							Computed: true,
						},
						// not currently exposed by the data source model.
						// "permalink_template": schema.StringAttribute{
						// 	Computed: true,
						// },
						// not currently exposed by the data source model.
						// "generated_slug": schema.StringAttribute{
						// 	Computed: true,
						// },
						"parent": schema.Int64Attribute{
							Computed: true,
						},
						"title": schema.SingleNestedAttribute{
							Computed: true,
							Attributes: map[string]schema.Attribute{
								"rendered": schema.StringAttribute{
									Computed: true,
								},
							},
						},
						"content": schema.SingleNestedAttribute{
							Computed: true,
							Attributes: map[string]schema.Attribute{
								"rendered": schema.StringAttribute{
									Computed: true,
								},
								"protected": schema.BoolAttribute{
									Computed: true,
								},
							},
						},
						"author": schema.Int64Attribute{
							Computed: true,
						},
						"excerpt": schema.SingleNestedAttribute{
							Computed: true,
							Attributes: map[string]schema.Attribute{
								"rendered": schema.StringAttribute{
									Computed: true,
								},
								"protected": schema.BoolAttribute{
									Computed: true,
								},
							},
						},
						"featured_media": schema.Int64Attribute{
							Computed: true,
						},
						"comment_status": schema.StringAttribute{
							Computed: true,
						},
						"ping_status": schema.StringAttribute{
							Computed: true,
						},
						"menu_order": schema.Int64Attribute{
							Computed: true,
						},
						"meta": schema.SingleNestedAttribute{
							Computed: true,
							Attributes: map[string]schema.Attribute{
								"footnotes": schema.StringAttribute{
									Computed: true,
								},
							},
						},
						"template": schema.StringAttribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func (d *pagesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state pagesDataSourceModel

	pages, err := d.client.ListPages(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Pages from Wordpress instance",
			err.Error(),
		)
		return
	}

	for _, page := range pages {
		pagesState := pagesModel{
			ID:             types.Int64Value(int64(page.ID)),
			Date:           types.StringValue(page.Date),
			Date_gmt:       types.StringValue(page.DateGMT),
			Link:           types.StringValue(page.Link),
			Modified:       types.StringValue(page.Modified),
			Modified_gmt:   types.StringValue(page.ModifiedGMT),
			Slug:           types.StringValue(page.Slug),
			Status:         types.StringValue(page.Status),
			Type:           types.StringValue(page.Type),
			Password:       types.StringValue(page.Password),
			Parent:         types.Int64Value(int64(page.Parent)),
			Author:         types.Int64Value(int64(page.Author)),
			Featured_media: types.Int64Value(int64(page.FeaturedMedia)),
			Comment_status: types.StringValue(page.CommentStatus),
			Ping_status:    types.StringValue(page.PingStatus),
			Menu_order:     types.Int64Value(int64(page.MenuOrder)),
			Template:       types.StringValue(page.Template),
			Title: renderedModel{
				Rendered: types.StringValue(page.Title.Rendered),
			},
			GUID: renderedModel{
				Rendered: types.StringValue(page.GUID.Rendered),
			},
			Content: renderedProtectedModel{
				Rendered:  types.StringValue(page.Content.Rendered),
				Protected: types.BoolValue(page.Content.Protected),
			},
			Excerpt: renderedProtectedModel{
				Rendered:  types.StringValue(page.Excerpt.Rendered),
				Protected: types.BoolValue(page.Excerpt.Protected),
			},
			Meta: footnotesModel{
				FootNotes: types.StringValue(""),
			},
		}

		state.Pages = append(state.Pages, pagesState)
	}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

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
	_ datasource.DataSource              = &pluginInfoDataSource{}
	_ datasource.DataSourceWithConfigure = &pluginInfoDataSource{}
)

type pluginInfoDataSource struct {
	client *wpapi.Client
}

type pluginInfoDataSourceModel struct {
	Slug             types.String `tfsdk:"slug"`
	Name             types.String `tfsdk:"name"`
	Version          types.String `tfsdk:"version"`
	Author           types.String `tfsdk:"author"`
	AuthorProfile    types.String `tfsdk:"author_profile"`
	Requires         types.String `tfsdk:"requires"`
	Tested           types.String `tfsdk:"tested"`
	RequiresPHP      types.String `tfsdk:"requires_php"`
	Rating           types.Int64  `tfsdk:"rating"`
	NumRatings       types.Int64  `tfsdk:"num_ratings"`
	ActiveInstalls   types.Int64  `tfsdk:"active_installs"`
	LastUpdated      types.String `tfsdk:"last_updated"`
	Homepage         types.String `tfsdk:"homepage"`
	DownloadLink     types.String `tfsdk:"download_link"`
	ShortDescription types.String `tfsdk:"short_description"`
}

func NewPluginInfoDataSource() datasource.DataSource {
	return &pluginInfoDataSource{}
}

func (d *pluginInfoDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *pluginInfoDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_plugin_info"
}

func (d *pluginInfoDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads public plugin metadata from the WordPress.org plugin registry by slug. This resource does not require a WordPress site to be configured, as it reads data from the WordPress.org plugin registry. If you only use this data source, you can pass empty values for the provider configuration.",
		Attributes: map[string]schema.Attribute{
			"slug": schema.StringAttribute{
				MarkdownDescription: "WordPress.org plugin slug, such as `woocommerce`.",
				Required:            true,
			},
			"name":              schema.StringAttribute{Computed: true},
			"version":           schema.StringAttribute{Computed: true},
			"author":            schema.StringAttribute{Computed: true},
			"author_profile":    schema.StringAttribute{Computed: true},
			"requires":          schema.StringAttribute{Computed: true},
			"tested":            schema.StringAttribute{Computed: true},
			"requires_php":      schema.StringAttribute{Computed: true},
			"rating":            schema.Int64Attribute{Computed: true},
			"num_ratings":       schema.Int64Attribute{Computed: true},
			"active_installs":   schema.Int64Attribute{Computed: true},
			"last_updated":      schema.StringAttribute{Computed: true},
			"homepage":          schema.StringAttribute{Computed: true},
			"download_link":     schema.StringAttribute{Computed: true},
			"short_description": schema.StringAttribute{Computed: true},
		},
	}
}

func (d *pluginInfoDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state pluginInfoDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	info, err := d.client.GetPluginInfo(ctx, state.Slug.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Plugin Info from WordPress.org",
			err.Error(),
		)
		return
	}

	state.Name = types.StringValue(info.Name)
	state.Version = types.StringValue(info.Version)
	state.Author = types.StringValue(info.Author)
	state.AuthorProfile = types.StringValue(info.AuthorProfile)
	state.Requires = types.StringValue(info.Requires)
	state.Tested = types.StringValue(info.Tested)
	state.RequiresPHP = types.StringValue(info.RequiresPHP)
	state.Rating = types.Int64Value(info.Rating)
	state.NumRatings = types.Int64Value(info.NumRatings)
	state.ActiveInstalls = types.Int64Value(info.ActiveInstalls)
	state.LastUpdated = types.StringValue(info.LastUpdated)
	state.Homepage = types.StringValue(info.Homepage)
	state.DownloadLink = types.StringValue(info.DownloadLink)
	state.ShortDescription = types.StringValue(info.ShortDescription)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

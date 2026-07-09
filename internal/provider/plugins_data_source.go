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
	_ datasource.DataSource              = &pluginsDataSource{}
	_ datasource.DataSourceWithConfigure = &pluginsDataSource{}
)

type pluginsDataSource struct {
	client *wpapi.Client
}

type pluginsDataSourceModel struct {
	Plugins []pluginDataSourceModel `tfsdk:"plugins"`
}

type pluginDataSourceModel struct {
	Plugin      types.String `tfsdk:"plugin"`
	Status      types.String `tfsdk:"status"`
	Name        types.String `tfsdk:"name"`
	PluginURI   types.String `tfsdk:"plugin_uri"`
	AuthorURI   types.String `tfsdk:"author_uri"`
	Version     types.String `tfsdk:"version"`
	NetworkOnly types.Bool   `tfsdk:"network_only"`
	RequiresWP  types.String `tfsdk:"requires_wp"`
	RequiresPHP types.String `tfsdk:"requires_php"`
	Textdomain  types.String `tfsdk:"textdomain"`
}

// NewPluginsDataSource returns the plugins data source implementation.
func NewPluginsDataSource() datasource.DataSource {
	return &pluginsDataSource{}
}

func (d *pluginsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *pluginsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_plugins"
}

func (d *pluginsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Data source for retrieving a list of installed WordPress plugins.",
		Attributes: map[string]schema.Attribute{
			"plugins": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"plugin":       schema.StringAttribute{Computed: true},
						"status":       schema.StringAttribute{Computed: true},
						"name":         schema.StringAttribute{Computed: true},
						"plugin_uri":   schema.StringAttribute{Computed: true},
						"author_uri":   schema.StringAttribute{Computed: true},
						"version":      schema.StringAttribute{Computed: true},
						"network_only": schema.BoolAttribute{Computed: true},
						"requires_wp":  schema.StringAttribute{Computed: true},
						"requires_php": schema.StringAttribute{Computed: true},
						"textdomain":   schema.StringAttribute{Computed: true},
					},
				},
			},
		},
	}
}

func (d *pluginsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state pluginsDataSourceModel

	plugins, err := d.client.ListPlugins(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Plugins from Wordpress instance",
			err.Error(),
		)
		return
	}

	for _, plugin := range plugins {
		state.Plugins = append(state.Plugins, pluginDataSourceModel{
			Plugin:      types.StringValue(plugin.Plugin),
			Status:      types.StringValue(plugin.Status),
			Name:        types.StringValue(plugin.Name),
			PluginURI:   types.StringValue(plugin.PluginURI),
			AuthorURI:   types.StringValue(plugin.AuthorURI),
			Version:     types.StringValue(plugin.Version),
			NetworkOnly: types.BoolValue(plugin.NetworkOnly),
			RequiresWP:  types.StringValue(plugin.RequiresWP),
			RequiresPHP: types.StringValue(plugin.RequiresPHP),
			Textdomain:  types.StringValue(plugin.Textdomain),
		})
	}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

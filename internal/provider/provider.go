// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"os"

	"terraform-provider-wordpress/internal/wpapi"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure ScaffoldingProvider satisfies various provider interfaces.
var _ provider.Provider = &WordpressProvider{}

// ScaffoldingProvider defines the provider implementation.
type WordpressProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// ScaffoldingProviderModel describes the provider data model.
type WordpressProviderModel struct {
	Host     types.String `tfsdk:"host"`
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
}

func configValue(config types.String, envNames ...string) string {
	if !config.IsNull() {
		return config.ValueString()
	}

	for _, envName := range envNames {
		if value, ok := os.LookupEnv(envName); ok {
			return value
		}
	}

	return ""
}

func (p *WordpressProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "wordpress"
	resp.Version = p.version
}

func (p *WordpressProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Provider for managing WordPress resources.\n\nProvider settings can also be read from environment variables: `WP_TF_PROVIDER_HOST`, `WP_TF_PROVIDER_USERNAME`, and `WP_TF_PROVIDER_PASSWORD`.",
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				MarkdownDescription: "The base URL of the WordPress site, including the REST API endpoint. Example: `http://localhost:8888/wp-json/wp/v2`. Can also be set via the `WP_TF_PROVIDER_HOST` or `WORDPRESS_HOST` environment variables.",
				Optional:            true,
			},
			"username": schema.StringAttribute{
				MarkdownDescription: "Username to authenticate with. Example: `admin`. Can also be set via the `WP_TF_PROVIDER_USERNAME` or `WORDPRESS_USERNAME` environment variables.",
				Optional:            true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "Password to authenticate with. Needs to be an application password, otherwise you will encounter `401 Unauthorized` errors. See https://make.wordpress.org/core/2020/11/05/application-passwords-integration-guide/ for more information. Can also be set via the `WP_TF_PROVIDER_PASSWORD` or `WORDPRESS_PASSWORD` environment variables.",
				Optional:            true,
			},
		},
	}
}

func (p *WordpressProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data WordpressProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	host := configValue(data.Host, "WP_TF_PROVIDER_HOST", "WORDPRESS_HOST")
	username := configValue(data.Username, "WP_TF_PROVIDER_USERNAME", "WORDPRESS_USERNAME")
	password := configValue(data.Password, "WP_TF_PROVIDER_PASSWORD", "WORDPRESS_PASSWORD")

	// Example client configuration for data sources and resources
	client, err := wpapi.New(host, username, password)
	if err != nil {
		resp.Diagnostics.AddError("Unable to configure WordPress client", err.Error())
		return
	}

	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *WordpressProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewPageResource,
		NewUserResource,
	}
}

func (p *WordpressProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewPagesDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &WordpressProvider{
			version: version,
		}
	}
}

// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"os"

	"terraform-provider-wordpress/internal/wpapi"
	"terraform-provider-wordpress/internal/wpappauth"

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
	Host     types.String                `tfsdk:"host"`
	AppAuth  *WordpressProviderAuthModel `tfsdk:"app_auth"`
	UserAuth *WordpressProviderAuthModel `tfsdk:"user_auth"`
}

type WordpressProviderAuthModel struct {
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
}

type providerData struct {
	AppClient   *wpapi.Client
	UserClient  *wpappauth.Service
	HasAppAuth  bool
	HasUserAuth bool
}

func configValue(config types.String, envNames ...string) string {
	if !config.IsNull() && !config.IsUnknown() {
		return config.ValueString()
	}

	for _, envName := range envNames {
		if value, ok := os.LookupEnv(envName); ok {
			return value
		}
	}

	return ""
}

func authConfigValue(config *WordpressProviderAuthModel, selector func(*WordpressProviderAuthModel) types.String, envNames ...string) string {
	if config != nil {
		return configValue(selector(config), envNames...)
	}

	for _, envName := range envNames {
		if value, ok := os.LookupEnv(envName); ok {
			return value
		}
	}

	return ""
}

func hasCredentialPair(username, password string) bool {
	return username != "" && password != ""
}

func (p *WordpressProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "wordpress"
	resp.Version = p.version
}

func (p *WordpressProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Provider for managing WordPress resources.\n\nProvider settings can also be read from environment variables. Configure `app_auth` for REST resources and data sources, and `user_auth` for nonce/AJAX workflows.",
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				MarkdownDescription: "The base URL of the WordPress site, including the REST API endpoint. Example: `http://localhost:8888/wp-json/wp/v2`. Can also be set via the `WP_TF_PROVIDER_HOST` or `WORDPRESS_HOST` environment variables.",
				Optional:            true,
			},
		},
		Blocks: map[string]schema.Block{
			"app_auth": schema.SingleNestedBlock{
				Description: "Application password authentication used by REST resources and data sources.",
				Attributes: map[string]schema.Attribute{
					"username": schema.StringAttribute{
						Optional: true,
						MarkdownDescription: "Username used together with the application password. Can also be set via `WP_TF_PROVIDER_APP_USERNAME` or `WORDPRESS_APP_USERNAME`.",
					},
					"password": schema.StringAttribute{
						Optional: true,
						Sensitive:           true,
						MarkdownDescription: "Application password for REST authentication. Can also be set via `WP_TF_PROVIDER_APP_PASSWORD` or `WORDPRESS_APP_PASSWORD`.",
					},
				},
			},
			"user_auth": schema.SingleNestedBlock{
				Description: "Normal WordPress user/password authentication used for nonce/AJAX workflows.",
				Attributes: map[string]schema.Attribute{
					"username": schema.StringAttribute{
						Optional: true,
						MarkdownDescription: "Username for nonce/AJAX authentication. Can also be set via `WP_TF_PROVIDER_USER_USERNAME` or `WORDPRESS_USER_USERNAME`.",
					},
					"password": schema.StringAttribute{
						Optional: true,
						Sensitive:           true,
						MarkdownDescription: "Normal user password for nonce/AJAX authentication. Can also be set via `WP_TF_PROVIDER_USER_PASSWORD` or `WORDPRESS_USER_PASSWORD`.",
					},
				},
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

	appUsername := authConfigValue(data.AppAuth, func(model *WordpressProviderAuthModel) types.String { return model.Username }, "WP_TF_PROVIDER_APP_USERNAME", "WORDPRESS_APP_USERNAME")
	appPassword := authConfigValue(data.AppAuth, func(model *WordpressProviderAuthModel) types.String { return model.Password }, "WP_TF_PROVIDER_APP_PASSWORD", "WORDPRESS_APP_PASSWORD")

	userUsername := authConfigValue(data.UserAuth, func(model *WordpressProviderAuthModel) types.String { return model.Username }, "WP_TF_PROVIDER_USER_USERNAME", "WORDPRESS_USER_USERNAME")
	userPassword := authConfigValue(data.UserAuth, func(model *WordpressProviderAuthModel) types.String { return model.Password }, "WP_TF_PROVIDER_USER_PASSWORD", "WORDPRESS_USER_PASSWORD")

	hasAppAuth := hasCredentialPair(appUsername, appPassword)
	hasUserAuth := hasCredentialPair(userUsername, userPassword)

	if host == "" && (hasAppAuth || hasUserAuth) {
		resp.Diagnostics.AddError(
			"Invalid provider configuration",
			"Host is required when app_auth or user_auth credentials are configured.",
		)
		return
	}

	dataSourceData := &providerData{HasAppAuth: hasAppAuth, HasUserAuth: hasUserAuth}
	resourceData := &providerData{HasAppAuth: hasAppAuth, HasUserAuth: hasUserAuth}

	if host != "" {
		appClient, err := wpapi.New(host, appUsername, appPassword)
		if err != nil {
			resp.Diagnostics.AddError("Unable to configure WordPress app_auth client", err.Error())
			return
		}

		dataSourceData.AppClient = appClient
		resourceData.AppClient = appClient

		if hasUserAuth {
			userClient := &wpappauth.Service{
				BaseURL:  host,
				Username: userUsername,
				Password: userPassword,
			}

			dataSourceData.UserClient = userClient
			resourceData.UserClient = userClient
		}
	}

	resp.DataSourceData = dataSourceData
	resp.ResourceData = resourceData
}

func (p *WordpressProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewPageResource,
		NewPluginResource,
		NewPostResource,
		NewThemeResource,
		NewUserResource,
	}
}

func (p *WordpressProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewPluginInfoDataSource,
		NewPagesDataSource,
		NewPluginsDataSource,
		NewPostsDataSource,
		NewUsersDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &WordpressProvider{
			version: version,
		}
	}
}

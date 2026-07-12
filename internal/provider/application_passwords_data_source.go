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
	_ datasource.DataSource              = &applicationPasswordsDataSource{}
	_ datasource.DataSourceWithConfigure = &applicationPasswordsDataSource{}
)

type applicationPasswordsDataSource struct {
	client *wpapi.Client
}

type applicationPasswordsDataSourceModel struct {
	UserID               types.Int64                     `tfsdk:"user_id"`
	ApplicationPasswords []applicationPasswordDataSource `tfsdk:"application_passwords"`
}

type applicationPasswordDataSource struct {
	UUID     types.String `tfsdk:"uuid"`
	AppID    types.String `tfsdk:"app_id"`
	Name     types.String `tfsdk:"name"`
	Created  types.String `tfsdk:"created"`
	LastUsed types.String `tfsdk:"last_used"`
	LastIP   types.String `tfsdk:"last_ip"`
}

func NewApplicationPasswordsDataSource() datasource.DataSource {
	return &applicationPasswordsDataSource{}
}

func (d *applicationPasswordsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *applicationPasswordsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_application_passwords"
}

func (d *applicationPasswordsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads all application password metadata records for a WordPress user.",
		Attributes: map[string]schema.Attribute{
			"user_id": schema.Int64Attribute{
				MarkdownDescription: "WordPress user ID that owns the application passwords.",
				Required:            true,
			},
			"application_passwords": schema.ListNestedAttribute{
				MarkdownDescription: "List of application passwords for the user.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"uuid": schema.StringAttribute{
							MarkdownDescription: "Unique identifier of the application password.",
							Computed:            true,
						},
						"app_id": schema.StringAttribute{
							MarkdownDescription: "Application-provided identifier for the password, if present.",
							Computed:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "Human-readable name of the application password.",
							Computed:            true,
						},
						"created": schema.StringAttribute{
							MarkdownDescription: "GMT timestamp of when the application password was created.",
							Computed:            true,
						},
						"last_used": schema.StringAttribute{
							MarkdownDescription: "GMT timestamp of last use, or null if never used.",
							Computed:            true,
						},
						"last_ip": schema.StringAttribute{
							MarkdownDescription: "IP address from last use, or null if never used.",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *applicationPasswordsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state applicationPasswordsDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	passwords, err := d.client.ListApplicationPasswords(ctx, state.UserID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Application Passwords from Wordpress instance",
			err.Error(),
		)
		return
	}

	state.ApplicationPasswords = make([]applicationPasswordDataSource, 0, len(passwords))
	for _, password := range passwords {
		state.ApplicationPasswords = append(state.ApplicationPasswords, applicationPasswordDataSource{
			UUID:     types.StringValue(password.UUID),
			AppID:    types.StringValue(password.AppID),
			Name:     types.StringValue(password.Name),
			Created:  types.StringValue(password.Created),
			LastUsed: nullableStringValue(password.LastUsed),
			LastIP:   nullableStringValue(password.LastIP),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

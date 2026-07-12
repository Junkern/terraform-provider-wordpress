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
	_ datasource.DataSource              = &applicationPasswordDataSourceOne{}
	_ datasource.DataSourceWithConfigure = &applicationPasswordDataSourceOne{}
)

type applicationPasswordDataSourceOne struct {
	client *wpapi.Client
}

type applicationPasswordDataSourceOneModel struct {
	UserID   types.Int64  `tfsdk:"user_id"`
	UUID     types.String `tfsdk:"uuid"`
	AppID    types.String `tfsdk:"app_id"`
	Name     types.String `tfsdk:"name"`
	Created  types.String `tfsdk:"created"`
	LastUsed types.String `tfsdk:"last_used"`
	LastIP   types.String `tfsdk:"last_ip"`
}

func NewApplicationPasswordDataSource() datasource.DataSource {
	return &applicationPasswordDataSourceOne{}
}

func (d *applicationPasswordDataSourceOne) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *applicationPasswordDataSourceOne) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_application_password"
}

func (d *applicationPasswordDataSourceOne) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads a single WordPress application password metadata record by user and UUID.",
		Attributes: map[string]schema.Attribute{
			"user_id": schema.Int64Attribute{
				MarkdownDescription: "WordPress user ID that owns the application password.",
				Required:            true,
			},
			"uuid": schema.StringAttribute{
				MarkdownDescription: "Unique identifier of the application password.",
				Required:            true,
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
	}
}

func (d *applicationPasswordDataSourceOne) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state applicationPasswordDataSourceOneModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	password, err := d.client.GetApplicationPassword(ctx, state.UserID.ValueInt64(), state.UUID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Application Password from Wordpress instance",
			err.Error(),
		)
		return
	}

	state.AppID = types.StringValue(password.AppID)
	state.Name = types.StringValue(password.Name)
	state.Created = types.StringValue(password.Created)
	state.LastUsed = nullableStringValue(password.LastUsed)
	state.LastIP = nullableStringValue(password.LastIP)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

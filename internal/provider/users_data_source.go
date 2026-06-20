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
	_ datasource.DataSource              = &usersDataSource{}
	_ datasource.DataSourceWithConfigure = &usersDataSource{}
)

type usersDataSource struct {
	client *wpapi.Client
}

type usersDataSourceModel struct {
	Users []userModel `tfsdk:"users"`
}

type userModel struct {
	ID             types.Int64  `tfsdk:"id"`
	Username       types.String `tfsdk:"username"`
	Name           types.String `tfsdk:"name"`
	FirstName      types.String `tfsdk:"first_name"`
	LastName       types.String `tfsdk:"last_name"`
	Email          types.String `tfsdk:"email"`
	URL            types.String `tfsdk:"url"`
	Description    types.String `tfsdk:"description"`
	Link           types.String `tfsdk:"link"`
	Locale         types.String `tfsdk:"locale"`
	Nickname       types.String `tfsdk:"nickname"`
	Slug           types.String `tfsdk:"slug"`
	RegisteredDate types.String `tfsdk:"registered_date"`
	Roles          types.List   `tfsdk:"roles"`
}

func NewUsersDataSource() datasource.DataSource {
	return &usersDataSource{}
}

func (d *usersDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *usersDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_users"
}

func (d *usersDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"users": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":              schema.Int64Attribute{Computed: true},
						"username":        schema.StringAttribute{Computed: true},
						"name":            schema.StringAttribute{Computed: true},
						"first_name":      schema.StringAttribute{Computed: true},
						"last_name":       schema.StringAttribute{Computed: true},
						"email":           schema.StringAttribute{Computed: true},
						"url":             schema.StringAttribute{Computed: true},
						"description":     schema.StringAttribute{Computed: true},
						"link":            schema.StringAttribute{Computed: true},
						"locale":          schema.StringAttribute{Computed: true},
						"nickname":        schema.StringAttribute{Computed: true},
						"slug":            schema.StringAttribute{Computed: true},
						"registered_date": schema.StringAttribute{Computed: true},
						"roles":           schema.ListAttribute{Computed: true, ElementType: types.StringType},
					},
				},
			},
		},
	}
}

func (d *usersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state usersDataSourceModel

	users, err := d.client.ListUsers(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Users from Wordpress instance",
			err.Error(),
		)
		return
	}

	for _, user := range users {
		usersState := userModel{
			ID:             types.Int64Value(user.ID),
			Username:       types.StringValue(user.Username),
			Name:           types.StringValue(user.Name),
			FirstName:      types.StringValue(user.FirstName),
			LastName:       types.StringValue(user.LastName),
			Email:          types.StringValue(user.Email),
			URL:            types.StringValue(user.URL),
			Description:    types.StringValue(user.Description),
			Link:           types.StringValue(user.Link),
			Locale:         types.StringValue(user.Locale),
			Nickname:       types.StringValue(user.Nickname),
			Slug:           types.StringValue(user.Slug),
			RegisteredDate: types.StringValue(user.RegisteredDate),
			Roles:          stringListValue(user.Roles),
		}

		state.Users = append(state.Users, usersState)
	}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

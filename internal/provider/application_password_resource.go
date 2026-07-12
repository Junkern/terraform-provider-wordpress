package provider

import (
	"context"
	"fmt"

	"terraform-provider-wordpress/internal/wpapi"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource              = &applicationPasswordResource{}
	_ resource.ResourceWithConfigure = &applicationPasswordResource{}
)

func NewApplicationPasswordResource() resource.Resource {
	return &applicationPasswordResource{}
}

type applicationPasswordResource struct {
	client *wpapi.Client
}

type applicationPasswordResourceModel struct {
	ID       types.String `tfsdk:"id"`
	UserID   types.Int64  `tfsdk:"user_id"`
	UUID     types.String `tfsdk:"uuid"`
	AppID    types.String `tfsdk:"app_id"`
	Name     types.String `tfsdk:"name"`
	Password types.String `tfsdk:"password"`
	Created  types.String `tfsdk:"created"`
	LastUsed types.String `tfsdk:"last_used"`
	LastIP   types.String `tfsdk:"last_ip"`
}

func (r *applicationPasswordResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_application_password"
}

func (r *applicationPasswordResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a WordPress application password for a specific user.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Terraform resource identifier in the format `user_id:uuid`.",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"user_id": schema.Int64Attribute{
				MarkdownDescription: "WordPress user ID that owns the application password.",
				Required:            true,
				PlanModifiers:       []planmodifier.Int64{int64planmodifier.RequiresReplace()},
			},
			"uuid": schema.StringAttribute{
				MarkdownDescription: "Unique identifier for the application password.",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"app_id": schema.StringAttribute{
				MarkdownDescription: "Optional UUID provided by the client application to identify itself.",
				Optional:            true,
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Human-readable name for this application password.",
				Required:            true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "Generated application password. This value is only returned by WordPress when the password is created.",
				Computed:            true,
				Sensitive:           true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"created": schema.StringAttribute{
				MarkdownDescription: "GMT timestamp of when the application password was created.",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"last_used": schema.StringAttribute{
				MarkdownDescription: "GMT timestamp of the last use, or null if never used.",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"last_ip": schema.StringAttribute{
				MarkdownDescription: "IP address from the last use, or null if never used.",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
		},
	}
}

func (r *applicationPasswordResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*wpapi.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *wpapi.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

func applicationPasswordModelFromAPI(password *wpapi.ApplicationPassword, prior applicationPasswordResourceModel) applicationPasswordResourceModel {
	if password == nil {
		return applicationPasswordResourceModel{}
	}

	model := applicationPasswordResourceModel{
		UserID:   prior.UserID,
		UUID:     types.StringValue(password.UUID),
		AppID:    types.StringValue(password.AppID),
		Name:     types.StringValue(password.Name),
		Created:  types.StringValue(password.Created),
		LastUsed: nullableStringValue(password.LastUsed),
		LastIP:   nullableStringValue(password.LastIP),
	}

	model.ID = types.StringValue(fmt.Sprintf("%d:%s", model.UserID.ValueInt64(), password.UUID))

	if password.Password != "" {
		model.Password = types.StringValue(password.Password)
	} else if !prior.Password.IsNull() && !prior.Password.IsUnknown() {
		model.Password = prior.Password
	} else {
		model.Password = types.StringNull()
	}

	return model
}

func applicationPasswordInputFromModel(model applicationPasswordResourceModel) wpapi.ApplicationPasswordInput {
	return wpapi.ApplicationPasswordInput{
		AppID: stringValuePointer(model.AppID),
		Name:  stringValuePointer(model.Name),
	}
}

func nullableStringValue(value *string) types.String {
	if value == nil {
		return types.StringNull()
	}
	return types.StringValue(*value)
}

func (r *applicationPasswordResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan applicationPasswordResourceModel

	tflog.Debug(ctx, "Wordpress application password create")
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	password, err := r.client.CreateApplicationPassword(ctx, plan.UserID.ValueInt64(), applicationPasswordInputFromModel(plan))
	if err != nil {
		resp.Diagnostics.AddError("Error creating application password", "Could not create application password, unexpected error: "+err.Error())
		return
	}

	plan = applicationPasswordModelFromAPI(password, plan)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *applicationPasswordResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state applicationPasswordResourceModel

	tflog.Debug(ctx, "Wordpress application password read")
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	password, err := r.client.GetApplicationPassword(ctx, state.UserID.ValueInt64(), state.UUID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Wordpress Application Password",
			"Could not read Wordpress Application Password "+state.UUID.ValueString()+": "+err.Error(),
		)
		return
	}

	state = applicationPasswordModelFromAPI(password, state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *applicationPasswordResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan applicationPasswordResourceModel

	tflog.Debug(ctx, "Wordpress application password update")
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	password, err := r.client.UpdateApplicationPassword(ctx, plan.UserID.ValueInt64(), plan.UUID.ValueString(), applicationPasswordInputFromModel(plan))
	if err != nil {
		resp.Diagnostics.AddError("Error updating application password", "Could not update application password, unexpected error: "+err.Error())
		return
	}

	plan = applicationPasswordModelFromAPI(password, plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *applicationPasswordResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state applicationPasswordResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteApplicationPassword(ctx, state.UserID.ValueInt64(), state.UUID.ValueString()); err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Wordpress Application Password",
			"Could not delete application password, unexpected error: "+err.Error(),
		)
		return
	}
}

package provider

import (
	"context"
	"fmt"
	"strconv"

	"terraform-provider-wordpress/internal/wpapi"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource              = &userResource{}
	_ resource.ResourceWithConfigure = &userResource{}
)

// NewUserResource is a helper function to simplify the provider implementation.
func NewUserResource() resource.Resource {
	return &userResource{}
}

type userResource struct {
	client *wpapi.Client
}

type userResourceModel struct {
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
	Password       types.String `tfsdk:"password"`
	ReassignTo     types.Int64  `tfsdk:"reassign_to"`
}

// Metadata returns the resource type name.
func (r *userResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

// Schema defines the schema for the resource.
func (r *userResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"username": schema.StringAttribute{
				Required: true,
			},
			"name": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"first_name": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"last_name": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"email": schema.StringAttribute{
				Required: true,
			},
			"url": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"description": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"link": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"locale": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"nickname": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"slug": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"registered_date": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"roles": schema.ListAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"password": schema.StringAttribute{
				Required:  true,
				Sensitive: true,
			},
			"reassign_to": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *userResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func userResourceModelFromUser(user *wpapi.User, prior userResourceModel) userResourceModel {
	if user == nil {
		return userResourceModel{}
	}

	model := userResourceModel{
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

	if !prior.Password.IsNull() && !prior.Password.IsUnknown() {
		model.Password = prior.Password
	}
	if !prior.ReassignTo.IsNull() && !prior.ReassignTo.IsUnknown() {
		model.ReassignTo = prior.ReassignTo
	}

	return model
}

func userResourceInputFromModel(ctx context.Context, model userResourceModel) wpapi.UserInput {
	input := wpapi.UserInput{
		Username:    stringValuePointer(model.Username),
		Name:        stringValuePointer(model.Name),
		FirstName:   stringValuePointer(model.FirstName),
		LastName:    stringValuePointer(model.LastName),
		Email:       stringValuePointer(model.Email),
		URL:         stringValuePointer(model.URL),
		Description: stringValuePointer(model.Description),
		Locale:      stringValuePointer(model.Locale),
		Nickname:    stringValuePointer(model.Nickname),
		Slug:        stringValuePointer(model.Slug),
		Password:    stringValuePointer(model.Password),
	}

	if roles := stringListFromValue(ctx, model.Roles); len(roles) > 0 {
		input.Roles = roles
	}

	return input
}

func stringValuePointer(value types.String) *string {
	if value.IsNull() || value.IsUnknown() {
		return nil
	}

	result := value.ValueString()
	return &result
}

func stringListValue(values []string) types.List {
	listValues := make([]attr.Value, 0, len(values))
	for _, value := range values {
		listValues = append(listValues, types.StringValue(value))
	}

	return types.ListValueMust(types.StringType, listValues)
}

func stringListFromValue(ctx context.Context, value types.List) []string {
	if value.IsNull() || value.IsUnknown() {
		return nil
	}

	var result []string
	if diags := value.ElementsAs(ctx, &result, false); diags.HasError() {
		return nil
	}

	return result
}

// Create creates the resource and sets the initial Terraform state.
func (r *userResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan userResourceModel

	tflog.Debug(ctx, fmt.Sprintf("Wordpress user create"))
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	user, err := r.client.CreateUser(ctx, userResourceInputFromModel(ctx, plan))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating user",
			"Could not create user, unexpected error: "+err.Error(),
		)
		return
	}

	plan = userResourceModelFromUser(user, plan)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *userResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Debug(ctx, fmt.Sprintf("Wordpress user read"))
	var state userResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	user, err := r.client.GetUser(ctx, state.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Wordpress User",
			"Could not read Wordpress User "+strconv.Itoa(int(state.ID.ValueInt64()))+": "+err.Error(),
		)
		return
	}

	state = userResourceModelFromUser(user, state)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *userResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan userResourceModel

	tflog.Debug(ctx, fmt.Sprintf("Wordpress user update"))
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	user, err := r.client.UpdateUser(ctx, plan.ID.ValueInt64(), userResourceInputFromModel(ctx, plan))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating user",
			"Could not update user, unexpected error: "+err.Error(),
		)
		return
	}

	plan = userResourceModelFromUser(user, plan)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *userResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state userResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	reassignTo := int64(1)
	if !state.ReassignTo.IsNull() && !state.ReassignTo.IsUnknown() {
		reassignTo = state.ReassignTo.ValueInt64()
	}

	if err := r.client.DeleteUser(ctx, state.ID.ValueInt64(), reassignTo); err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Wordpress User",
			"Could not delete user, unexpected error: "+err.Error(),
		)
		return
	}
}

package provider

import (
	"context"

	"terraform-provider-wordpress/internal/wpappauth"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource              = &themeResource{}
	_ resource.ResourceWithConfigure = &themeResource{}
)

// NewThemeResource returns the theme resource implementation.
func NewThemeResource() resource.Resource {
	return &themeResource{}
}

type themeResource struct {
	client *wpappauth.Service
}

type themeResourceModel struct {
	ID   types.String `tfsdk:"id"`
	Slug types.String `tfsdk:"slug"`
}

// Metadata returns the resource type name.
func (r *themeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_theme"
}

// Schema defines the schema for the resource.
func (r *themeResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a WordPress theme by slug. Creating this resource installs the theme via wp-admin AJAX and deleting it removes the theme. This resource does not cover activating or deactivating the theme. This resource needs the `user_auth` provider configuration block because it uses wp-admin AJAX requests to install and delete the theme.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"slug": schema.StringAttribute{
				Required:    true,
				Description: "The theme slug, for example `astra`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *themeResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, err := userClientForProviderData(req.ProviderData)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Configure Resource",
			err.Error(),
		)

		return
	}

	r.client = client
}

// Create installs the theme and stores the slug in state.
func (r *themeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan themeResourceModel

	tflog.Debug(ctx, "Wordpress theme create")
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	slug := plan.Slug.ValueString()
	if err := r.client.InstallTheme(ctx, slug); err != nil {
		resp.Diagnostics.AddError(
			"Error creating theme",
			"Could not install theme, unexpected error: "+err.Error(),
		)
		return
	}

	plan.ID = types.StringValue(slug)
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// Read keeps the managed slug in state.
func (r *themeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Debug(ctx, "Wordpress theme read")
	var state themeResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.ID.IsNull() || state.ID.IsUnknown() {
		state.ID = state.Slug
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// Update is a no-op because slug changes require replacement.
func (r *themeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan themeResourceModel

	tflog.Debug(ctx, "Wordpress theme update")
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.ID = plan.Slug
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// Delete removes the theme from the WordPress installation.
func (r *themeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state themeResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	slug := state.Slug.ValueString()
	if slug == "" {
		slug = state.ID.ValueString()
	}
	if slug == "" {
		resp.Diagnostics.AddError("Error Deleting Wordpress Theme", "theme slug is missing from state")
		return
	}

	if err := r.client.DeleteTheme(ctx, slug); err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Wordpress Theme",
			"Could not delete theme, unexpected error: "+err.Error(),
		)
		return
	}
}

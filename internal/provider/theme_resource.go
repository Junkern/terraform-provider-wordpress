package provider

import (
	"context"

	"terraform-provider-wordpress/internal/wpapi"
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
	client    *wpappauth.Service
	appClient *wpapi.Client
}

type themeResourceModel struct {
	ID     types.String `tfsdk:"id"`
	Slug   types.String `tfsdk:"slug"`
	Active types.Bool   `tfsdk:"active"`
}

// Metadata returns the resource type name.
func (r *themeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_theme"
}

// Schema defines the schema for the resource.
func (r *themeResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a WordPress theme by slug. Creating this resource installs the theme via wp-admin AJAX and deleting it removes the theme. Set `active` to true to activate the theme. This resource needs `user_auth` for wp-admin operations and `app_auth` to read theme status.",
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
			"active": schema.BoolAttribute{
				Optional:    true,
				Description: "Whether the theme should be active.",
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
	if data, ok := req.ProviderData.(*providerData); ok {
		r.appClient = data.AppClient
	}
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
	if plan.Active.ValueBool() {
		if err := r.client.ActivateTheme(ctx, slug); err != nil {
			resp.Diagnostics.AddError("Error activating theme", "Could not activate theme, unexpected error: "+err.Error())
			return
		}
	}

	plan.ID = types.StringValue(slug)
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// Read refreshes the managed slug and activation status.
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
	if r.appClient != nil {
		themes, err := r.appClient.ListThemes(ctx)
		if err != nil {
			resp.Diagnostics.AddError("Error reading theme status", "Could not read themes, unexpected error: "+err.Error())
			return
		}
		for _, theme := range themes {
			if theme.Stylesheet == state.Slug.ValueString() {
				state.Active = types.BoolValue(theme.Status == "active")
				break
			}
		}
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
	if plan.Active.ValueBool() {
		if err := r.client.ActivateTheme(ctx, plan.Slug.ValueString()); err != nil {
			resp.Diagnostics.AddError("Error activating theme", "Could not activate theme, unexpected error: "+err.Error())
			return
		}
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

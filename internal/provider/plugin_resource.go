package provider

import (
	"context"
	"fmt"

	"terraform-provider-wordpress/internal/wpapi"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource              = &pluginResource{}
	_ resource.ResourceWithConfigure = &pluginResource{}
)

// NewPluginResource returns the plugin resource implementation.
func NewPluginResource() resource.Resource {
	return &pluginResource{}
}

type pluginResource struct {
	client *wpapi.Client
}

type pluginResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Plugin      types.String `tfsdk:"plugin"`
	Slug        types.String `tfsdk:"slug"`
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

// Metadata returns the resource type name.
func (r *pluginResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_plugin"
}

// Schema defines the schema for the resource.
func (r *pluginResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Wordpress plugin. Creating this resource will install the plugin, and deleting it will uninstall the plugin. The plugin can be activated or deactivated by setting the `status` attribute to `active` or `inactive`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"plugin": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"slug": schema.StringAttribute{
				Required: true,
				Description: "The slug of the plugin. This is used to identify the plugin in WordPress and to install it. For example, the slug for the Hello Dolly plugin is `hello-dolly`",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"status": schema.StringAttribute{
				Optional: true,
				Description: "The status of the plugin. Can be 'active' or 'inactive'.",
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"plugin_uri": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"author_uri": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"version": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"network_only": schema.BoolAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"requires_wp": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"requires_php": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"textdomain": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *pluginResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func pluginResourceModelFromPlugin(plugin *wpapi.Plugin, prior pluginResourceModel) pluginResourceModel {
	if plugin == nil {
		return pluginResourceModel{}
	}

	model := pluginResourceModel{
		ID:          types.StringValue(plugin.Plugin),
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
	}

	if !prior.Slug.IsNull() && !prior.Slug.IsUnknown() {
		model.Slug = prior.Slug
	}

	return model
}

func pluginResourceInputFromModel(model pluginResourceModel) wpapi.PluginInput {
	input := wpapi.PluginInput{Slug: model.Slug.ValueString()}
	if status := stringValuePointer(model.Status); status != nil && *status != "" {
		input.Status = status
	}

	return input
}

func pluginIdentifier(model pluginResourceModel) string {
	if !model.ID.IsNull() && !model.ID.IsUnknown() && model.ID.ValueString() != "" {
		return model.ID.ValueString()
	}

	if !model.Plugin.IsNull() && !model.Plugin.IsUnknown() {
		return model.Plugin.ValueString()
	}

	return ""
}

// Create installs the plugin and stores the resulting plugin file in state.
func (r *pluginResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan pluginResourceModel

	tflog.Debug(ctx, "Wordpress plugin create")
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	plugin, err := r.client.CreatePlugin(ctx, pluginResourceInputFromModel(plan))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating plugin",
			"Could not install plugin, unexpected error: "+err.Error(),
		)
		return
	}

	plan = pluginResourceModelFromPlugin(plugin, plan)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest plugin data.
func (r *pluginResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Debug(ctx, "Wordpress plugin read")
	var state pluginResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	identifier := pluginIdentifier(state)
	if identifier == "" {
		resp.Diagnostics.AddError("Error Reading Wordpress Plugin", "plugin identifier is missing from state")
		return
	}

	plugin, err := r.client.GetPlugin(ctx, identifier)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Wordpress Plugin",
			"Could not read Wordpress Plugin "+identifier+": "+err.Error(),
		)
		return
	}

	state = pluginResourceModelFromPlugin(plugin, state)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// Update applies plugin activation changes.
func (r *pluginResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan pluginResourceModel

	tflog.Debug(ctx, "Wordpress plugin update")
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	identifier := pluginIdentifier(plan)
	if identifier == "" {
		resp.Diagnostics.AddError("Error Updating Wordpress Plugin", "plugin identifier is missing from state")
		return
	}

	plugin, err := r.client.UpdatePlugin(ctx, identifier, pluginResourceInputFromModel(plan))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating plugin",
			"Could not update plugin, unexpected error: "+err.Error(),
		)
		return
	}

	plan = pluginResourceModelFromPlugin(plugin, plan)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// Delete removes the plugin from the WordPress installation.
func (r *pluginResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state pluginResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	identifier := pluginIdentifier(state)
	if identifier == "" {
		resp.Diagnostics.AddError("Error Deleting Wordpress Plugin", "plugin identifier is missing from state")
		return
	}

	if err := r.client.DeletePlugin(ctx, identifier); err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Wordpress Plugin",
			"Could not delete plugin, unexpected error: "+err.Error(),
		)
		return
	}
}

package provider

import (
	"context"

	"terraform-provider-wordpress/internal/wpapi"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource              = &wpAPISettingsResource{}
	_ resource.ResourceWithConfigure = &wpAPISettingsResource{}
)

// NewWPAPISettingsResource returns the WordPress REST API settings resource.
func NewWPAPISettingsResource() resource.Resource { return &wpAPISettingsResource{} }

type wpAPISettingsResource struct {
	client *wpapi.Client
}

type wpAPISettingsResourceModel struct {
	ID                   types.String `tfsdk:"id"`
	Title                types.String `tfsdk:"title"`
	Description          types.String `tfsdk:"description"`
	URL                  types.String `tfsdk:"url"`
	Email                types.String `tfsdk:"email"`
	Timezone             types.String `tfsdk:"timezone"`
	DateFormat           types.String `tfsdk:"date_format"`
	TimeFormat           types.String `tfsdk:"time_format"`
	StartOfWeek          types.Int64  `tfsdk:"start_of_week"`
	Language             types.String `tfsdk:"language"`
	UseSmilies           types.Bool   `tfsdk:"use_smilies"`
	DefaultCategory      types.Int64  `tfsdk:"default_category"`
	DefaultPostFormat    types.String `tfsdk:"default_post_format"`
	PostsPerPage         types.Int64  `tfsdk:"posts_per_page"`
	ShowOnFront          types.String `tfsdk:"show_on_front"`
	PageOnFront          types.Int64  `tfsdk:"page_on_front"`
	PageForPosts         types.Int64  `tfsdk:"page_for_posts"`
	DefaultPingStatus    types.String `tfsdk:"default_ping_status"`
	DefaultCommentStatus types.String `tfsdk:"default_comment_status"`
	SiteLogo             types.Int64  `tfsdk:"site_logo"`
	SiteIcon             types.Int64  `tfsdk:"site_icon"`
}

func (r *wpAPISettingsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_wp_api_settings"
}

func (r *wpAPISettingsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages WordPress site settings through the REST API (https://developer.wordpress.org/rest-api/reference/settings/). WordPress does not support deleting site settings; destroying this resource only removes it from Terraform state.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"title":                  settingStringAttribute("Site title."),
			"description":            settingStringAttribute("Site tagline."),
			"url":                    settingStringAttribute("Site URL."),
			"email":                  settingStringAttribute("Administration email address."),
			"timezone":               settingStringAttribute("Site timezone."),
			"date_format":            settingStringAttribute("Site date format."),
			"time_format":            settingStringAttribute("Site time format."),
			"language":               settingStringAttribute("WordPress locale code."),
			"default_post_format":    settingStringAttribute("Default post format."),
			"show_on_front":          settingStringAttribute("Front page display mode."),
			"default_ping_status":    settingStringAttribute("Default ping status."),
			"default_comment_status": settingStringAttribute("Default comment status."),
			"start_of_week":          settingInt64Attribute("First day of the week."),
			"default_category":       settingInt64Attribute("Default post category."),
			"posts_per_page":         settingInt64Attribute("Maximum posts shown on a blog page."),
			"page_on_front":          settingInt64Attribute("Front page ID."),
			"page_for_posts":         settingInt64Attribute("Posts page ID."),
			"site_logo":              settingInt64Attribute("Site logo media ID."),
			"site_icon":              settingInt64Attribute("Site icon media ID."),
			"use_smilies": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether to convert emoticons to graphics.",
			},
		},
	}
}

func settingStringAttribute(description string) schema.StringAttribute {
	return schema.StringAttribute{Optional: true, Computed: true, Description: description}
}

func settingInt64Attribute(description string) schema.Int64Attribute {
	return schema.Int64Attribute{Optional: true, Computed: true, Description: description}
}

func (r *wpAPISettingsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, err := appClientForProviderData(req.ProviderData)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Configure Resource", err.Error())
		return
	}
	r.client = client
}

func (r *wpAPISettingsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan wpAPISettingsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	settings, err := r.client.UpdateSettings(ctx, wpAPISettingsInputFromModel(plan))
	if err != nil {
		resp.Diagnostics.AddError("Error creating WordPress API settings", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, wpAPISettingsModelFromSettings(settings))...)
}

func (r *wpAPISettingsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state wpAPISettingsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	settings, err := r.client.GetSettings(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error reading WordPress API settings", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, wpAPISettingsModelFromSettings(settings))...)
}

func (r *wpAPISettingsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan wpAPISettingsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	settings, err := r.client.UpdateSettings(ctx, wpAPISettingsInputFromModel(plan))
	if err != nil {
		resp.Diagnostics.AddError("Error updating WordPress API settings", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, wpAPISettingsModelFromSettings(settings))...)
}

// Delete only removes the singleton from Terraform state; WordPress has no settings delete endpoint.
func (r *wpAPISettingsResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
}

func wpAPISettingsModelFromSettings(s *wpapi.Settings) wpAPISettingsResourceModel {
	return wpAPISettingsResourceModel{
		ID: types.StringValue("settings"), Title: types.StringValue(s.Title), Description: types.StringValue(s.Description), URL: types.StringValue(s.URL), Email: types.StringValue(s.Email), Timezone: types.StringValue(s.Timezone), DateFormat: types.StringValue(s.DateFormat), TimeFormat: types.StringValue(s.TimeFormat), StartOfWeek: types.Int64Value(s.StartOfWeek), Language: types.StringValue(s.Language), UseSmilies: types.BoolValue(s.UseSmilies), DefaultCategory: types.Int64Value(s.DefaultCategory), DefaultPostFormat: types.StringValue(s.DefaultPostFormat), PostsPerPage: types.Int64Value(s.PostsPerPage), ShowOnFront: types.StringValue(s.ShowOnFront), PageOnFront: types.Int64Value(s.PageOnFront), PageForPosts: types.Int64Value(s.PageForPosts), DefaultPingStatus: types.StringValue(s.DefaultPingStatus), DefaultCommentStatus: types.StringValue(s.DefaultCommentStatus), SiteLogo: types.Int64Value(s.SiteLogo), SiteIcon: types.Int64Value(s.SiteIcon),
	}
}

func wpAPISettingsInputFromModel(m wpAPISettingsResourceModel) wpapi.SettingsInput {
	return wpapi.SettingsInput{
		Title: stringValuePointer(m.Title), Description: stringValuePointer(m.Description), URL: stringValuePointer(m.URL), Email: stringValuePointer(m.Email), Timezone: stringValuePointer(m.Timezone), DateFormat: stringValuePointer(m.DateFormat), TimeFormat: stringValuePointer(m.TimeFormat), StartOfWeek: int64ValuePointer(m.StartOfWeek), Language: stringValuePointer(m.Language), UseSmilies: boolValuePointer(m.UseSmilies), DefaultCategory: int64ValuePointer(m.DefaultCategory), DefaultPostFormat: stringValuePointer(m.DefaultPostFormat), PostsPerPage: int64ValuePointer(m.PostsPerPage), ShowOnFront: stringValuePointer(m.ShowOnFront), PageOnFront: int64ValuePointer(m.PageOnFront), PageForPosts: int64ValuePointer(m.PageForPosts), DefaultPingStatus: stringValuePointer(m.DefaultPingStatus), DefaultCommentStatus: stringValuePointer(m.DefaultCommentStatus), SiteLogo: int64ValuePointer(m.SiteLogo), SiteIcon: int64ValuePointer(m.SiteIcon),
	}
}

func int64ValuePointer(value types.Int64) *int64 {
	if value.IsNull() || value.IsUnknown() {
		return nil
	}
	result := value.ValueInt64()
	return &result
}

func boolValuePointer(value types.Bool) *bool {
	if value.IsNull() || value.IsUnknown() {
		return nil
	}
	result := value.ValueBool()
	return &result
}

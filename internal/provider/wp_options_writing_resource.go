package provider

import (
	"context"

	"terraform-provider-wordpress/internal/wpappauth"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &wpOptionsWritingResource{}
var _ resource.ResourceWithConfigure = &wpOptionsWritingResource{}

func NewWPOptionsWritingResource() resource.Resource { return &wpOptionsWritingResource{} }

type wpOptionsWritingResource struct{ client *wpappauth.Service }

type wpOptionsWritingResourceModel struct {
	ID                   types.String `tfsdk:"id"`
	DefaultCategory      types.Int64  `tfsdk:"default_category"`
	DefaultPostFormat    types.String `tfsdk:"default_post_format"`
	MailserverURL        types.String `tfsdk:"mailserver_url"`
	MailserverPort       types.Int64  `tfsdk:"mailserver_port"`
	MailserverLogin      types.String `tfsdk:"mailserver_login"`
	MailserverPass       types.String `tfsdk:"mailserver_pass"`
	DefaultEmailCategory types.Int64  `tfsdk:"default_email_category"`
	PingSites            types.String `tfsdk:"ping_sites"`
}

func (r *wpOptionsWritingResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_wp_options_writing"
}

func (r *wpOptionsWritingResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	optionalString := func(description string) schema.StringAttribute {
		return schema.StringAttribute{Optional: true, Computed: true, Description: description}
	}
	optionalInt := func(description string) schema.Int64Attribute {
		return schema.Int64Attribute{Optional: true, Computed: true, Description: description}
	}
	resp.Schema = schema.Schema{Description: "Manages WordPress Writing Settings through wp-admin. Requires user_auth credentials.", Attributes: map[string]schema.Attribute{
		"id":                     schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"default_category":       optionalInt("Default post category ID."),
		"default_post_format":    optionalString("Default post format."),
		"mailserver_url":         optionalString("Mail server hostname."),
		"mailserver_port":        optionalInt("Mail server port."),
		"mailserver_login":       optionalString("Mail server login."),
		"mailserver_pass":        schema.StringAttribute{Optional: true, Sensitive: true, Description: "Mail server password."},
		"default_email_category": optionalInt("Default category for posts sent by email."),
		"ping_sites":             optionalString("URLs to notify when publishing a post."),
	}}
}

func (r *wpOptionsWritingResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, err := userClientForProviderData(req.ProviderData)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Configure Resource", err.Error())
		return
	}
	r.client = client
}

func (r *wpOptionsWritingResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan wpOptionsWritingResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.UpdateWritingOptions(ctx, writingOptionsFromModel(plan)); err != nil {
		resp.Diagnostics.AddError("Error creating WordPress writing options", err.Error())
		return
	}
	plan.ID = types.StringValue("writing")
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *wpOptionsWritingResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state wpOptionsWritingResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	options, err := r.client.GetWritingOptions(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error reading WordPress writing options", err.Error())
		return
	}
	password := state.MailserverPass
	state = writingOptionsModel(options)
	state.MailserverPass = password
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *wpOptionsWritingResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan wpOptionsWritingResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.UpdateWritingOptions(ctx, writingOptionsFromModel(plan)); err != nil {
		resp.Diagnostics.AddError("Error updating WordPress writing options", err.Error())
		return
	}
	plan.ID = types.StringValue("writing")
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *wpOptionsWritingResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
}

func writingOptionsFromModel(m wpOptionsWritingResourceModel) wpappauth.WritingOptions {
	return wpappauth.WritingOptions{DefaultCategory: m.DefaultCategory.ValueInt64(), DefaultPostFormat: m.DefaultPostFormat.ValueString(), MailserverURL: m.MailserverURL.ValueString(), MailserverPort: m.MailserverPort.ValueInt64(), MailserverLogin: m.MailserverLogin.ValueString(), MailserverPass: m.MailserverPass.ValueString(), DefaultEmailCategory: m.DefaultEmailCategory.ValueInt64(), PingSites: m.PingSites.ValueString()}
}

func writingOptionsModel(o *wpappauth.WritingOptions) wpOptionsWritingResourceModel {
	return wpOptionsWritingResourceModel{ID: types.StringValue("writing"), DefaultCategory: types.Int64Value(o.DefaultCategory), DefaultPostFormat: types.StringValue(o.DefaultPostFormat), MailserverURL: types.StringValue(o.MailserverURL), MailserverPort: types.Int64Value(o.MailserverPort), MailserverLogin: types.StringValue(o.MailserverLogin), MailserverPass: types.StringValue(o.MailserverPass), DefaultEmailCategory: types.Int64Value(o.DefaultEmailCategory), PingSites: types.StringValue(o.PingSites)}
}

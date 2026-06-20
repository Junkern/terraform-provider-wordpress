package provider

import (
	"context"
	"fmt"
	"strconv"

	"terraform-provider-wordpress/internal/wpapi"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource               = &postResource{}
	_ resource.ResourceWithConfigure  = &postResource{}
	_ resource.ResourceWithModifyPlan = &postResource{}
)

// NewPostResource is a helper function to simplify the provider implementation.
func NewPostResource() resource.Resource {
	return &postResource{}
}

type postResource struct {
	client *wpapi.Client
}

type postResourceModel struct {
	ID             types.Int64                     `tfsdk:"id"`
	Date           types.String                    `tfsdk:"date"`
	Date_gmt       types.String                    `tfsdk:"date_gmt"`
	Link           types.String                    `tfsdk:"link"`
	Modified       types.String                    `tfsdk:"modified"`
	Slug           types.String                    `tfsdk:"slug"`
	Status         types.String                    `tfsdk:"status"`
	Type           types.String                    `tfsdk:"type"`
	Password       types.String                    `tfsdk:"password"`
	Author         types.Int64                     `tfsdk:"author"`
	Featured_media types.Int64                     `tfsdk:"featured_media"`
	Comment_status types.String                    `tfsdk:"comment_status"`
	Ping_status    types.String                    `tfsdk:"ping_status"`
	Format         types.String                    `tfsdk:"format"`
	Sticky         types.Bool                      `tfsdk:"sticky"`
	Template       types.String                    `tfsdk:"template"`
	Title          *renderedResourceModel          `tfsdk:"title"`
	Content        *renderedProtectedResourceModel `tfsdk:"content"`
	Excerpt        *renderedProtectedResourceModel `tfsdk:"excerpt"`
}

// Metadata returns the resource type name.
func (r *postResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_post"
}

// Schema defines the schema for the resource.
func (r *postResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"date": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"date_gmt": schema.StringAttribute{
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
			"modified": schema.StringAttribute{
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
			"status": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"type": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"password": schema.StringAttribute{
				Optional:  true,
				Computed:  true,
				Sensitive: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"title": schema.SingleNestedAttribute{
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"rendered": schema.StringAttribute{
						Computed: true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"raw": schema.StringAttribute{
						Optional: true,
					},
				},
			},
			"content": schema.SingleNestedAttribute{
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"rendered": schema.StringAttribute{
						Computed: true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"protected": schema.BoolAttribute{
						Computed: true,
						PlanModifiers: []planmodifier.Bool{
							boolplanmodifier.UseStateForUnknown(),
						},
					},
					"raw": schema.StringAttribute{
						Optional: true,
					},
				},
			},
			"author": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"excerpt": schema.SingleNestedAttribute{
				Optional: true,
				Computed: true,
				Default: objectdefault.StaticValue(
					types.ObjectValueMust(
						map[string]attr.Type{
							"rendered":  types.StringType,
							"raw":       types.StringType,
							"protected": types.BoolType,
						},
						map[string]attr.Value{
							"rendered":  types.StringValue(""),
							"raw":       types.StringValue(""),
							"protected": types.BoolValue(false),
						},
					),
				),
				Attributes: map[string]schema.Attribute{
					"rendered": schema.StringAttribute{
						Computed: true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"protected": schema.BoolAttribute{
						Computed: true,
						PlanModifiers: []planmodifier.Bool{
							boolplanmodifier.UseStateForUnknown(),
						},
					},
					"raw": schema.StringAttribute{
						Optional: true,
					},
				},
			},
			"featured_media": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"comment_status": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"ping_status": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"format": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"sticky": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"template": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (d *postResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func postResourceModelFromPost(post *wpapi.Post, prior postResourceModel) postResourceModel {
	if post == nil {
		return postResourceModel{}
	}

	titleRaw := types.StringNull()
	if prior.Title != nil {
		titleRaw = prior.Title.Raw
	}

	contentRaw := types.StringNull()
	if prior.Content != nil {
		contentRaw = prior.Content.Raw
	}

	excerptRaw := types.StringNull()
	if prior.Excerpt != nil {
		excerptRaw = prior.Excerpt.Raw
	}

	return postResourceModel{
		ID:             types.Int64Value(post.ID),
		Date:           types.StringValue(post.Date),
		Date_gmt:       types.StringValue(post.DateGMT),
		Link:           types.StringValue(post.Link),
		Modified:       types.StringValue(post.Modified),
		Slug:           types.StringValue(post.Slug),
		Status:         types.StringValue(post.Status),
		Type:           types.StringValue(post.Type),
		Password:       types.StringValue(post.Password),
		Author:         types.Int64Value(post.Author),
		Featured_media: types.Int64Value(post.FeaturedMedia),
		Comment_status: types.StringValue(post.CommentStatus),
		Ping_status:    types.StringValue(post.PingStatus),
		Format:         types.StringValue(post.Format),
		Sticky:         types.BoolValue(post.Sticky),
		Template:       types.StringValue(post.Template),
		Title: &renderedResourceModel{
			Rendered: types.StringValue(post.Title.Rendered),
			Raw:      titleRaw,
		},
		Content: &renderedProtectedResourceModel{
			Rendered:  types.StringValue(post.Content.Rendered),
			Raw:       contentRaw,
			Protected: types.BoolValue(false),
		},
		Excerpt: &renderedProtectedResourceModel{
			Rendered:  types.StringValue(post.Excerpt.Rendered),
			Raw:       excerptRaw,
			Protected: types.BoolValue(false),
		},
	}
}

func postResourceHasConfigChange(plan postResourceModel, state postResourceModel) bool {
	return resourceStringValue(plan.Title).ValueString() != resourceStringValue(state.Title).ValueString() ||
		resourceProtectedStringValue(plan.Content).ValueString() != resourceProtectedStringValue(state.Content).ValueString() ||
		resourceProtectedStringValue(plan.Excerpt).ValueString() != resourceProtectedStringValue(state.Excerpt).ValueString() ||
		plan.Status.ValueString() != state.Status.ValueString() ||
		plan.Type.ValueString() != state.Type.ValueString() ||
		plan.Password.ValueString() != state.Password.ValueString() ||
		plan.Author.ValueInt64() != state.Author.ValueInt64() ||
		plan.Featured_media.ValueInt64() != state.Featured_media.ValueInt64() ||
		plan.Comment_status.ValueString() != state.Comment_status.ValueString() ||
		plan.Ping_status.ValueString() != state.Ping_status.ValueString() ||
		plan.Format.ValueString() != state.Format.ValueString() ||
		plan.Sticky.ValueBool() != state.Sticky.ValueBool() ||
		plan.Template.ValueString() != state.Template.ValueString()
}

func (r *postResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	var plan postResourceModel
	var state postResourceModel

	if req.Plan.Raw.IsNull() || req.State.Raw.IsNull() {
		return
	}

	planDiags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(planDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	stateDiags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(stateDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.ID.IsNull() || !postResourceHasConfigChange(plan, state) {
		return
	}

	plan.Date = types.StringUnknown()
	plan.Date_gmt = types.StringUnknown()
	plan.Modified = types.StringUnknown()
	if plan.Title != nil {
		plan.Title.Rendered = types.StringUnknown()
	}
	if plan.Content != nil {
		plan.Content.Rendered = types.StringUnknown()
	}
	if plan.Excerpt != nil {
		plan.Excerpt.Rendered = types.StringUnknown()
	}
	plan.Slug = types.StringUnknown()

	planDiags = resp.Plan.Set(ctx, plan)
	resp.Diagnostics.Append(planDiags...)
}

func postResourceInputFromModel(plan postResourceModel) wpapi.PostInput {
	input := wpapi.PostInput{
		Title:    stringPointer(resourceStringValue(plan.Title)),
		Content:  stringPointer(resourceProtectedStringValue(plan.Content)),
		Excerpt:  stringPointer(resourceProtectedStringValue(plan.Excerpt)),
		Status:   stringPointer(plan.Status),
		Type:     stringPointer(plan.Type),
		Password: stringPointer(plan.Password),
		Format:   stringPointer(plan.Format),
		Template: stringPointer(plan.Template),
	}

	if !plan.Author.IsNull() {
		value := plan.Author.ValueInt64()
		input.Author = &value
	}
	if !plan.Featured_media.IsNull() {
		value := plan.Featured_media.ValueInt64()
		input.FeaturedMedia = &value
	}
	if !plan.Comment_status.IsNull() {
		value := plan.Comment_status.ValueString()
		if value != "" {
			input.CommentStatus = &value
		}
	}
	if !plan.Ping_status.IsNull() {
		value := plan.Ping_status.ValueString()
		if value != "" {
			input.PingStatus = &value
		}
	}
	if !plan.Sticky.IsNull() {
		value := plan.Sticky.ValueBool()
		input.Sticky = &value
	}

	return input
}

// Create creates the resource and sets the initial Terraform state.
func (r *postResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan postResourceModel

	tflog.Debug(ctx, "Wordpress post create")
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Build post")
	post, err := r.client.CreatePost(ctx, postResourceInputFromModel(plan))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating post",
			"Could not create post, unexpected error: "+err.Error(),
		)
		return
	}

	plan = postResourceModelFromPost(post, plan)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *postResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Debug(ctx, "Wordpress post read")
	var state postResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	post, err := r.client.GetPost(ctx, state.ID.ValueInt64())

	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Wordpress Post",
			"Could not read Wordpress Post "+strconv.Itoa(int(state.ID.ValueInt64()))+": "+err.Error(),
		)
		return
	}

	state = postResourceModelFromPost(post, state)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *postResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan postResourceModel

	tflog.Debug(ctx, "Wordpress post update")
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	post, err := r.client.UpdatePost(ctx, plan.ID.ValueInt64(), postResourceInputFromModel(plan))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating post",
			"Could not update post, unexpected error: "+err.Error(),
		)
		return
	}

	plan = postResourceModelFromPost(post, plan)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *postResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state postResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeletePost(ctx, state.ID.ValueInt64()); err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Wordpress Post",
			"Could not delete post, unexpected error: "+err.Error(),
		)
		return
	}
}

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

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource               = &pageResource{}
	_ resource.ResourceWithConfigure  = &pageResource{}
	_ resource.ResourceWithModifyPlan = &pageResource{}
)

// NewPageResource is a helper function to simplify the provider implementation.
func NewPageResource() resource.Resource {
	return &pageResource{}
}

// pageResource is the resource implementation.
type pageResource struct {
	client *wpapi.Client
}

type pageResourceModel struct {
	ID types.Int64 `tfsdk:"id"`
	// GUID         renderedResourceModel `tfsdk:"guid"`
	Date     types.String `tfsdk:"date"`
	Date_gmt types.String `tfsdk:"date_gmt"`
	Link     types.String `tfsdk:"link"`
	Modified types.String `tfsdk:"modified"`
	// Modified_gmt types.String `tfsdk:"modified_gmt"`
	Slug     types.String `tfsdk:"slug"`
	Status   types.String `tfsdk:"status"`
	Type     types.String `tfsdk:"type"`
	Password types.String `tfsdk:"password"`
	// TODO: not currently exposed by the resource model.
	// Permalink_template types.String `tfsdk:"permalink_template"`
	// TODO: not currently exposed by the resource model.
	// Generated_slug types.String `tfsdk:"generated_slug"`
	Parent         types.Int64                     `tfsdk:"parent"`
	Author         types.Int64                     `tfsdk:"author"`
	Featured_media types.Int64                     `tfsdk:"featured_media"`
	Comment_status types.String                    `tfsdk:"comment_status"`
	Ping_status    types.String                    `tfsdk:"ping_status"`
	Menu_order     types.Int64                     `tfsdk:"menu_order"`
	Template       types.String                    `tfsdk:"template"`
	Title          *renderedResourceModel          `tfsdk:"title"`
	Content        *renderedProtectedResourceModel `tfsdk:"content"`
	Excerpt        *renderedProtectedResourceModel `tfsdk:"excerpt"`
}

type renderedResourceModel struct {
	Rendered types.String `tfsdk:"rendered"`
	Raw      types.String `tfsdk:"raw"`
}

type renderedProtectedResourceModel struct {
	Rendered  types.String `tfsdk:"rendered"`
	Protected types.Bool   `tfsdk:"protected"`
	Raw       types.String `tfsdk:"raw"`
}

// Metadata returns the resource type name.
func (r *pageResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_page"
}

// Schema defines the schema for the resource.
func (r *pageResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
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
			// "guid": schema.SingleNestedAttribute{
			// 	Computed: true,
			// 	Attributes: map[string]schema.Attribute{
			// 		"rendered": schema.StringAttribute{
			// 			Computed: true,
			// 		},
			// 		"raw": schema.StringAttribute{
			// 			Optional: true,
			// 		},
			// 	},
			// 	// Attributes: map[string]schema.Attribute{
			// 	// 	"rendered": schema.StringAttribute{
			// 	// 		PlanModifiers: []planmodifier.String{
			// 	// 			stringplanmodifier.UseStateForUnknown(),
			// 	// 		},
			// 	// 		Required: true,
			// 	// 	},
			// 	// 	"raw": schema.StringAttribute{
			// 	// 		PlanModifiers: []planmodifier.String{
			// 	// 			stringplanmodifier.UseStateForUnknown(),
			// 	// 		},
			// 	// 		Required: true,
			// 	// 	},
			// 	// },
			// },
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
			// "modified_gmt": schema.StringAttribute{
			// 	Computed: true,
			// },
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
				// TODO status can be "publish, future, draft, pending, private"
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
			// not currently exposed by the resource model.
			// "permalink_template": schema.StringAttribute{
			// 	Computed: true,
			// },
			// not currently exposed by the resource model.
			// "generated_slug": schema.StringAttribute{
			// 	Computed: true,
			// },
			"parent": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
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
						Computed: true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
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
				// TODO: One of "open, closed"
			},
			"ping_status": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				// TODO: One of "open, closed"
			},
			"menu_order": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			// "meta": schema.SingleNestedAttribute{
			// 	Optional: true,
			// 	Attributes: map[string]schema.Attribute{
			// 		"footnotes": schema.StringAttribute{
			// 			Computed: true,
			// 		},
			// 	},
			// },
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

func (d *pageResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func pageResourceModelFromPage(page *wpapi.Page, prior pageResourceModel) pageResourceModel {
	if page == nil {
		return pageResourceModel{}
	}
	titleRaw := types.StringNull()
	if prior.Title != nil {
		titleRaw = prior.Title.Raw
	}

	contentRaw := types.StringNull()
	if prior.Content != nil {
		contentRaw = prior.Content.Raw
	}

	return pageResourceModel{
		ID:             types.Int64Value(page.ID),
		Date:           types.StringValue(page.Date),
		Date_gmt:       types.StringValue(page.DateGMT),
		Link:           types.StringValue(page.Link),
		Modified:       types.StringValue(page.Modified),
		Slug:           types.StringValue(page.Slug),
		Status:         types.StringValue(page.Status),
		Type:           types.StringValue(page.Type),
		Password:       types.StringValue(page.Password),
		Parent:         types.Int64Value(page.Parent),
		Author:         types.Int64Value(page.Author),
		Featured_media: types.Int64Value(page.FeaturedMedia),
		Comment_status: types.StringValue(page.CommentStatus),
		Ping_status:    types.StringValue(page.PingStatus),
		Menu_order:     types.Int64Value(page.MenuOrder),
		Template:       types.StringValue(page.Template),
		Title: &renderedResourceModel{
			Rendered: types.StringValue(page.Title.Rendered),
			Raw:      titleRaw,
		},
		Content: &renderedProtectedResourceModel{
			Rendered:  types.StringValue(page.Content.Rendered),
			Raw:       contentRaw,
			Protected: types.BoolValue(false),
		},
		Excerpt: &renderedProtectedResourceModel{
			Rendered:  types.StringValue(page.Excerpt.Rendered),
			Raw:       types.StringValue(page.Excerpt.Raw),
			Protected: types.BoolValue(false),
		},
	}
}

func resourceStringValue(model *renderedResourceModel) types.String {
	if model == nil {
		return types.StringNull()
	}

	return model.Raw
}

func resourceProtectedStringValue(model *renderedProtectedResourceModel) types.String {
	if model == nil {
		return types.StringNull()
	}

	return model.Raw
}

func pageResourceHasConfigChange(plan pageResourceModel, state pageResourceModel) bool {
	return resourceStringValue(plan.Title).ValueString() != resourceStringValue(state.Title).ValueString() ||
		resourceProtectedStringValue(plan.Content).ValueString() != resourceProtectedStringValue(state.Content).ValueString() ||
		resourceProtectedStringValue(plan.Excerpt).ValueString() != resourceProtectedStringValue(state.Excerpt).ValueString() ||
		plan.Status.ValueString() != state.Status.ValueString() ||
		plan.Type.ValueString() != state.Type.ValueString() ||
		plan.Password.ValueString() != state.Password.ValueString() ||
		plan.Parent.ValueInt64() != state.Parent.ValueInt64() ||
		plan.Author.ValueInt64() != state.Author.ValueInt64() ||
		plan.Featured_media.ValueInt64() != state.Featured_media.ValueInt64() ||
		plan.Comment_status.ValueString() != state.Comment_status.ValueString() ||
		plan.Ping_status.ValueString() != state.Ping_status.ValueString() ||
		plan.Menu_order.ValueInt64() != state.Menu_order.ValueInt64() ||
		plan.Template.ValueString() != state.Template.ValueString()
}

func (r *pageResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	var plan pageResourceModel
	var state pageResourceModel

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

	if state.ID.IsNull() || !pageResourceHasConfigChange(plan, state) {
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

func pageResourceInputFromModel(plan pageResourceModel) wpapi.PageInput {
	input := wpapi.PageInput{
		Title:    stringPointer(resourceStringValue(plan.Title)),
		Content:  stringPointer(resourceProtectedStringValue(plan.Content)),
		Excerpt:  stringPointer(resourceProtectedStringValue(plan.Excerpt)),
		Status:   stringPointer(plan.Status),
		Type:     stringPointer(plan.Type),
		Password: stringPointer(plan.Password),
		Template: stringPointer(plan.Template),
	}

	if !plan.Parent.IsNull() {
		value := plan.Parent.ValueInt64()
		input.Parent = &value
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
	if !plan.Menu_order.IsNull() {
		value := plan.Menu_order.ValueInt64()
		input.MenuOrder = &value
	}

	return input
}

func stringPointer(value types.String) *string {
	if value.IsNull() || value.IsUnknown() {
		return nil
	}

	result := value.ValueString()
	return &result
}

// Create creates the resource and sets the initial Terraform state.
func (r *pageResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan pageResourceModel

	tflog.Debug(ctx, "Wordpress page create")
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Build page")
	page, err := r.client.CreatePage(ctx, pageResourceInputFromModel(plan))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating page",
			"Could not create page, unexpected error: "+err.Error(),
		)
		return
	}

	plan = pageResourceModelFromPage(page, plan)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *pageResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Debug(ctx, "Wordpress page read")
	var state pageResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	page, err := r.client.GetPage(ctx, state.ID.ValueInt64())

	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Wordpress Page",
			"Could not read Wordpress Page "+strconv.Itoa(int(state.ID.ValueInt64()))+": "+err.Error(),
		)
		return
	}

	state = pageResourceModelFromPage(page, state)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *pageResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan pageResourceModel

	tflog.Debug(ctx, "Wordpress page update")
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	page, err := r.client.UpdatePage(ctx, plan.ID.ValueInt64(), pageResourceInputFromModel(plan))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating page",
			"Could not update page, unexpected error: "+err.Error(),
		)
		return
	}

	plan = pageResourceModelFromPage(page, plan)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *pageResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state pageResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeletePage(ctx, state.ID.ValueInt64()); err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Wordpress Page",
			"Could not delete page, unexpected error: "+err.Error(),
		)
		return
	}
}

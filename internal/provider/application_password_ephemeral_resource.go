package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"terraform-provider-wordpress/internal/wpapi"

	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	eschema "github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ ephemeral.EphemeralResource              = &applicationPasswordEphemeralResource{}
	_ ephemeral.EphemeralResourceWithClose     = &applicationPasswordEphemeralResource{}
	_ ephemeral.EphemeralResourceWithConfigure = &applicationPasswordEphemeralResource{}
)

const applicationPasswordEphemeralPrivateCleanupKey = "application_password_cleanup"

func NewApplicationPasswordEphemeralResource() ephemeral.EphemeralResource {
	return &applicationPasswordEphemeralResource{}
}

type applicationPasswordEphemeralResource struct {
	client *wpapi.Client
}

type applicationPasswordEphemeralModel struct {
	UserID        types.Int64  `tfsdk:"user_id"`
	Name          types.String `tfsdk:"name"`
	AppID         types.String `tfsdk:"app_id"`
	DeleteOnClose types.Bool   `tfsdk:"delete_on_close"`
	UUID          types.String `tfsdk:"uuid"`
	Password      types.String `tfsdk:"password"`
	Created       types.String `tfsdk:"created"`
}

type applicationPasswordEphemeralCleanupState struct {
	DeleteOnClose bool   `json:"delete_on_close"`
	UserID        int64  `json:"user_id"`
	UUID          string `json:"uuid"`
}

func (r *applicationPasswordEphemeralResource) Metadata(_ context.Context, req ephemeral.MetadataRequest, resp *ephemeral.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_application_password_ephemeral"
}

func (r *applicationPasswordEphemeralResource) Schema(_ context.Context, _ ephemeral.SchemaRequest, resp *ephemeral.SchemaResponse) {
	resp.Schema = eschema.Schema{
		Description: "Creates a short-lived WordPress application password for ephemeral Terraform workflows without persisting the password in state.",
		Attributes: map[string]eschema.Attribute{
			"user_id": eschema.Int64Attribute{
				MarkdownDescription: "WordPress user ID that will own the generated application password.",
				Required:            true,
			},
			"name": eschema.StringAttribute{
				MarkdownDescription: "Human-readable name for the generated application password.",
				Required:            true,
			},
			"app_id": eschema.StringAttribute{
				MarkdownDescription: "Optional UUID provided by the caller to identify the client application.",
				Optional:            true,
			},
			"delete_on_close": eschema.BoolAttribute{
				MarkdownDescription: "Whether Terraform should delete the generated application password during ephemeral close. Defaults to `true`.",
				Optional:            true,
			},
			"uuid": eschema.StringAttribute{
				MarkdownDescription: "Unique identifier of the generated application password.",
				Computed:            true,
			},
			"password": eschema.StringAttribute{
				MarkdownDescription: "Generated application password value returned by WordPress.",
				Computed:            true,
				Sensitive:           true,
			},
			"created": eschema.StringAttribute{
				MarkdownDescription: "GMT timestamp of when the application password was created.",
				Computed:            true,
			},
		},
	}
}

func (r *applicationPasswordEphemeralResource) Configure(_ context.Context, req ephemeral.ConfigureRequest, resp *ephemeral.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*wpapi.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Ephemeral Resource Configure Type",
			fmt.Sprintf("Expected *wpapi.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *applicationPasswordEphemeralResource) Open(ctx context.Context, req ephemeral.OpenRequest, resp *ephemeral.OpenResponse) {
	var config applicationPasswordEphemeralModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deleteOnClose := true
	if !config.DeleteOnClose.IsNull() && !config.DeleteOnClose.IsUnknown() {
		deleteOnClose = config.DeleteOnClose.ValueBool()
	}

	created, err := r.client.CreateApplicationPassword(ctx, config.UserID.ValueInt64(), wpapi.ApplicationPasswordInput{
		Name:  stringValuePointer(config.Name),
		AppID: stringValuePointer(config.AppID),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create ephemeral application password",
			err.Error(),
		)
		return
	}

	result := applicationPasswordEphemeralModel{
		UserID:        config.UserID,
		Name:          types.StringValue(created.Name),
		AppID:         types.StringValue(created.AppID),
		DeleteOnClose: types.BoolValue(deleteOnClose),
		UUID:          types.StringValue(created.UUID),
		Password:      types.StringValue(created.Password),
		Created:       types.StringValue(created.Created),
	}

	cleanupStateBytes, err := json.Marshal(applicationPasswordEphemeralCleanupState{
		DeleteOnClose: deleteOnClose,
		UserID:        config.UserID.ValueInt64(),
		UUID:          created.UUID,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to prepare ephemeral application password cleanup",
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.Private.SetKey(ctx, applicationPasswordEphemeralPrivateCleanupKey, cleanupStateBytes)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.Result.Set(ctx, &result)...)
}

func (r *applicationPasswordEphemeralResource) Close(ctx context.Context, req ephemeral.CloseRequest, resp *ephemeral.CloseResponse) {
	cleanupStateBytes, diags := req.Private.GetKey(ctx, applicationPasswordEphemeralPrivateCleanupKey)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if len(cleanupStateBytes) == 0 {
		return
	}

	var cleanupState applicationPasswordEphemeralCleanupState
	if err := json.Unmarshal(cleanupStateBytes, &cleanupState); err != nil {
		resp.Diagnostics.AddError(
			"Unable to decode ephemeral application password cleanup data",
			err.Error(),
		)
		return
	}

	if err := r.deleteApplicationPasswordIfRequested(ctx, cleanupState.UserID, cleanupState.UUID, cleanupState.DeleteOnClose); err != nil {
		resp.Diagnostics.AddError(
			"Unable to delete ephemeral application password",
			err.Error(),
		)
	}
}

func (r *applicationPasswordEphemeralResource) deleteApplicationPasswordIfRequested(ctx context.Context, userID int64, uuid string, deleteOnClose bool) error {
	if !deleteOnClose {
		return nil
	}

	return r.client.DeleteApplicationPassword(ctx, userID, uuid)
}

// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/i1snow/terraform-provider-scylladb/scylla"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &keyspaceResource{}
var _ resource.ResourceWithConfigure = &keyspaceResource{}
var _ resource.ResourceWithImportState = &keyspaceResource{}

func NewKeyspaceResource() resource.Resource {
	return &keyspaceResource{}
}

// keyspaceResource defines the resource implementation.
type keyspaceResource struct {
	client *scylla.Cluster
}

// keyspaceResourceModel maps the resource schema data.
type keyspaceResourceModel struct {
	ID                types.String `tfsdk:"id"`
	LastUpdated       types.String `tfsdk:"last_updated"`
	Name              types.String `tfsdk:"name"`
	DurableWrites     types.Bool   `tfsdk:"durable_writes"`
	ReplicationClass  types.String `tfsdk:"replication_class"`
	ReplicationFactor types.Int64  `tfsdk:"replication_factor"`
}

// Metadata returns the resource type name.
func (r *keyspaceResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_keyspace"
}

// Schema defines the schema for the resource.
func (r *keyspaceResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a ScyllaDB keyspace.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The name of the keyspace (same as name).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"last_updated": schema.StringAttribute{
				Computed:    true,
				Description: "The time of the last update to the resource.",
			},
			"name": schema.StringAttribute{
				Description: "The name of the keyspace.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"durable_writes": schema.BoolAttribute{
				Description: "Whether durable writes are enabled. Defaults to true.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"replication_class": schema.StringAttribute{
				Description: "The replication strategy class. Defaults to SimpleStrategy.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("SimpleStrategy"),
			},
			"replication_factor": schema.Int64Attribute{
				Description: "The replication factor. Defaults to 1.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(1),
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *keyspaceResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*scylla.Cluster)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *scylla.Cluster, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

// Create creates the resource and sets the initial Terraform state.
func (r *keyspaceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan keyspaceResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	keyspace := keyspaceFromPlan(plan)

	err := r.client.CreateKeyspace(keyspace)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create keyspace",
			err.Error(),
		)
		return
	}

	// Set computed attributes
	plan.ID = types.StringValue(keyspace.Name)
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *keyspaceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state keyspaceResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	keyspace, err := r.client.GetKeyspace(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read keyspace",
			err.Error(),
		)
		return
	}

	// Update state with refreshed values, preserving LastUpdated
	state.Name = types.StringValue(keyspace.Name)
	state.DurableWrites = types.BoolValue(keyspace.DurableWrites)
	state.ReplicationClass = types.StringValue(keyspace.ReplicationClass)
	state.ReplicationFactor = types.Int64Value(int64(keyspace.ReplicationFactor))

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update updates the resource and sets the updated Terraform state.
func (r *keyspaceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan keyspaceResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	keyspace := keyspaceFromPlan(plan)

	err := r.client.UpdateKeyspace(keyspace)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to update keyspace",
			err.Error(),
		)
		return
	}

	// Set computed attributes
	plan.ID = types.StringValue(keyspace.Name)
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete deletes the resource.
func (r *keyspaceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state keyspaceResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	keyspace := scylla.Keyspace{
		Name: state.Name.ValueString(),
	}

	err := r.client.DeleteKeyspace(keyspace)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to delete keyspace",
			err.Error(),
		)
		return
	}
}

// ImportState imports an existing resource.
func (r *keyspaceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// keyspaceFromPlan converts a plan model to a scylla.Keyspace.
func keyspaceFromPlan(plan keyspaceResourceModel) scylla.Keyspace {
	return scylla.Keyspace{
		Name:              plan.Name.ValueString(),
		DurableWrites:     plan.DurableWrites.ValueBool(),
		ReplicationClass:  plan.ReplicationClass.ValueString(),
		ReplicationFactor: int(plan.ReplicationFactor.ValueInt64()),
	}
}

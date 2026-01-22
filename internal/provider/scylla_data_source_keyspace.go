// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/i1snow/terraform-provider-scylladb/scylla"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &keyspaceDataSource{}
	_ datasource.DataSourceWithConfigure = &keyspaceDataSource{}
)

// NewKeyspaceDataSource is a helper function to simplify the provider implementation.
func NewKeyspaceDataSource() datasource.DataSource {
	return &keyspaceDataSource{}
}

// keyspaceDataSource is the data source implementation.
type keyspaceDataSource struct {
	client *scylla.Cluster
}

// keyspaceDataSourceModel maps the data source schema data.
type keyspaceDataSourceModel struct {
	ID                types.String `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	DurableWrites     types.Bool   `tfsdk:"durable_writes"`
	ReplicationClass  types.String `tfsdk:"replication_class"`
	ReplicationFactor types.Int64  `tfsdk:"replication_factor"`
}

// Metadata returns the data source type name.
func (d *keyspaceDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_keyspace"
}

// Schema defines the schema for the data source.
func (d *keyspaceDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches information about a ScyllaDB keyspace.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The name of the keyspace to look up.",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Computed:    true,
				Description: "The name of the keyspace.",
			},
			"durable_writes": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether durable writes are enabled for this keyspace.",
			},
			"replication_class": schema.StringAttribute{
				Computed:    true,
				Description: "The replication strategy class (e.g., SimpleStrategy, NetworkTopologyStrategy).",
			},
			"replication_factor": schema.Int64Attribute{
				Computed:    true,
				Description: "The replication factor for the keyspace.",
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *keyspaceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config keyspaceDataSourceModel

	// Read config.
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	keyspace, err := d.client.GetKeyspace(config.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read keyspace",
			err.Error(),
		)
		return
	}

	// Map response body to model.
	state := keyspaceDataSourceModel{
		ID:                config.ID,
		Name:              types.StringValue(keyspace.Name),
		DurableWrites:     types.BoolValue(keyspace.DurableWrites),
		ReplicationClass:  types.StringValue(keyspace.ReplicationClass),
		ReplicationFactor: types.Int64Value(int64(keyspace.ReplicationFactor)),
	}

	// Set state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Configure adds the provider configured client to the data source.
func (d *keyspaceDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*scylla.Cluster)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *scylla.Cluster, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = client
}

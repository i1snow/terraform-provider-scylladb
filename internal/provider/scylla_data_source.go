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
	_ datasource.DataSource              = &roleDataSource{}
	_ datasource.DataSourceWithConfigure = &roleDataSource{}
)

// NewRoleDataSource is a helper function to simplify the provider implementation.
func NewRoleDataSource() datasource.DataSource {
	return &roleDataSource{}
}

// roleDataSource is the data source implementation.
type roleDataSource struct {
	client *scylla.Cluster
}

// roleDataSourceModel maps the data source schema data.
type roleDataSourceModel struct {
	Role        types.String   `tfsdk:"role"`
	CanLogin    types.Bool     `tfsdk:"can_login"`
	IsSuperuser types.Bool     `tfsdk:"is_superuser"`
	MemberOf    []types.String `tfsdk:"member_of"`
}

// Metadata returns the data source type name.
func (d *roleDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_role"
}

// Schema defines the schema for the data source.
func (d *roleDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"role": schema.StringAttribute{
				Computed: true,
			},
			"can_login": schema.BoolAttribute{
				Computed: true,
			},
			"is_superuser": schema.BoolAttribute{
				Computed: true,
			},
			"member_of": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *roleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state roleDataSourceModel

	curRole, err := d.client.GetRole("cassandra")
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read the role",
			err.Error(),
		)
		return
	}

	// Map response body to model
	state = roleDataSourceModel{
		Role:        types.StringValue(curRole.Role),
		CanLogin:    types.BoolValue(curRole.CanLogin),
		IsSuperuser: types.BoolValue(curRole.IsSuperuser),
	}
	for _, member := range curRole.MemberOf {
		state.MemberOf = append(state.MemberOf, types.StringValue(member))
	}

	// Set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Configure adds the provider configured client to the data source.
func (d *roleDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Add a nil check when handling ProviderData because Terraform
	// sets that data after it calls the ConfigureProvider RPC.
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

// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/i1snow/terraform-provider-scylladb/internal/consts"
	"github.com/i1snow/terraform-provider-scylladb/scylla"
)

// Ensure ScylladbProvider satisfies various provider interfaces.
var _ provider.Provider = &scylladbProvider{}
var _ provider.ProviderWithFunctions = &scylladbProvider{}
var _ provider.ProviderWithEphemeralResources = &scylladbProvider{}
var _ provider.ProviderWithActions = &scylladbProvider{}

// ScylladbProvider defines the provider implementation.
type scylladbProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// scylladbProviderModel describes the provider data model.
type scylladbProviderModel struct {
	Host              types.String `tfsdk:"host"`
	Port              types.Int64  `tfsdk:"port"`
	AuthLoginUserPass struct {
		Username types.String `tfsdk:"username"`
		Password types.String `tfsdk:"password"`
	} `tfsdk:"auth_login_userpass"`
}

func (p *scylladbProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "scylladb"
	resp.Version = p.version
}

func (p *scylladbProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Configure access to ScyllaDB.",
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				MarkdownDescription: "Hostname or IP address of the ScyllaDB instance.",
				Optional:            true,
			},
			"port": schema.Int64Attribute{
				MarkdownDescription: "Port number of the ScyllaDB instance.",
				Optional:            true,
			},
		},
		Blocks: map[string]schema.Block{
			consts.FieldAuthLoginUserpass: schema.SingleNestedBlock{
				Description: "Login to ScyllaDB using the userpass method",
				Attributes: map[string]schema.Attribute{
					"username": schema.StringAttribute{
						Description: "Login with username",
						Required:    false,
					},
					"password": schema.StringAttribute{
						Description: "Login with password",
						Required:    false,
						Sensitive:   true,
					},
				},
			},
		},
	}
}

func (p *scylladbProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	tflog.Info(ctx, "Configuring Scylla client")

	// Retrieve provider data from configuration.
	var data scylladbProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Info(ctx, "Provider configuration", map[string]interface{}{
		"host": data.Host.ValueString(),
		//"username": data.Username.ValueString(),
		// Do not log sensitive values such as passwords.
	})

	if data.Host.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Unknown HashiCups API Host",
			"The provider cannot create the HashiCups API client as there is an unknown configuration value for the HashiCups API host. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the HASHICUPS_HOST environment variable.",
		)
	}

	// if data.Username.IsUnknown() {
	// 	resp.Diagnostics.AddAttributeError(
	// 		path.Root("username"),
	// 		"Unknown HashiCups API Username",
	// 		"The provider cannot create the HashiCups API client as there is an unknown configuration value for the HashiCups API username. "+
	// 			"Either target apply the source of the value first, set the value statically in the configuration, or use the HASHICUPS_USERNAME environment variable.",
	// 	)
	// }

	// if data.Password.IsUnknown() {
	// 	resp.Diagnostics.AddAttributeError(
	// 		path.Root("password"),
	// 		"Unknown HashiCups API Password",
	// 		"The provider cannot create the HashiCups API client as there is an unknown configuration value for the HashiCups API password. "+
	// 			"Either target apply the source of the value first, set the value statically in the configuration, or use the HASHICUPS_PASSWORD environment variable.",
	// 	)
	// }

	if resp.Diagnostics.HasError() {
		return
	}

	// Default values to environment variables, but override
	// with Terraform configuration value if set.

	host := os.Getenv("HASHICUPS_HOST")
	username := os.Getenv("HASHICUPS_USERNAME")
	password := os.Getenv("HASHICUPS_PASSWORD")

	if !data.Host.IsNull() {
		host = data.Host.ValueString()
	}

	// if !data.Username.IsNull() {
	// 	username = data.Username.ValueString()
	// }

	// if !data.Password.IsNull() {
	// 	password = data.Password.ValueString()
	// }

	// If any of the expected configurations are missing, return
	// errors with provider-specific guidance.

	if host == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Missing HashiCups API Host",
			"The provider cannot create the HashiCups API client as there is a missing or empty value for the HashiCups API host. "+
				"Set the host value in the configuration or use the HASHICUPS_HOST environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	// if username == "" {
	// 	resp.Diagnostics.AddAttributeError(
	// 		path.Root("username"),
	// 		"Missing HashiCups API Username",
	// 		"The provider cannot create the HashiCups API client as there is a missing or empty value for the HashiCups API username. "+
	// 			"Set the username value in the configuration or use the HASHICUPS_USERNAME environment variable. "+
	// 			"If either is already set, ensure the value is not empty.",
	// 	)
	// }

	// if password == "" {
	// 	resp.Diagnostics.AddAttributeError(
	// 		path.Root("password"),
	// 		"Missing HashiCups API Password",
	// 		"The provider cannot create the HashiCups API client as there is a missing or empty value for the HashiCups API password. "+
	// 			"Set the password value in the configuration or use the HASHICUPS_PASSWORD environment variable. "+
	// 			"If either is already set, ensure the value is not empty.",
	// 	)
	// }

	if resp.Diagnostics.HasError() {
		return
	}

	client := scylla.NewClusterConfig([]string{host})
	if username != "" && password != "" {
		client.SetUserPasswordAuth(username, password)
	}

	err := client.CreateSession()
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create HashiCups API Client",
			"An unexpected error was encountered trying to create the HashiCups API client. "+
				"Please verify the provider configuration values are correct and try again. "+
				"If the problem persists, please contact HashiCorp support.\n\n"+
				err.Error(),
		)
		return
	}

	// Make the HashiCups client available during DataSource and Resource
	// type Configure methods.
	resp.DataSourceData = &client
	resp.ResourceData = &client

	tflog.Info(ctx, "Configured HashiCups client", map[string]any{"success": true})
}

func (p *scylladbProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewExampleResource,
	}
}

func (p *scylladbProvider) EphemeralResources(ctx context.Context) []func() ephemeral.EphemeralResource {
	return []func() ephemeral.EphemeralResource{
		NewExampleEphemeralResource,
	}
}

func (p *scylladbProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewRoleDataSource,
	}
}

func (p *scylladbProvider) Functions(ctx context.Context) []func() function.Function {
	return []func() function.Function{
		NewExampleFunction,
	}
}

func (p *scylladbProvider) Actions(ctx context.Context) []func() action.Action {
	return []func() action.Action{
		NewExampleAction,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &scylladbProvider{
			version: version,
		}
	}
}

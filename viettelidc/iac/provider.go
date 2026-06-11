// Package csa implements the ViettelIDC IaC (Cloud Service Application) resources
// for Terraform using the Plugin Framework v1.x.
//
// This package is muxed alongside the existing VCD provider (Plugin SDK v2)
// in main.go via tf6muxserver. Both share the provider TypeName "viettelidc"
// but expose different resource families:
//   - VCD provider (SDK v2):  vcd_*        (e.g. vcd_vapp, vcd_org)
//   - CSA provider (this):    viettelidc_* (e.g. viettelidc_subnet)
package iac

import (
	"context"
	"fmt"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-viettelidc/terraform-provider-viettelidc/v3/viettelidc/iac/client"
	"github.com/terraform-viettelidc/terraform-provider-viettelidc/v3/viettelidc/iac/networking"
	"github.com/terraform-viettelidc/terraform-provider-viettelidc/v3/viettelidc/iac/providerdata"
	"github.com/terraform-viettelidc/terraform-provider-viettelidc/v3/viettelidc/iac/vpc"
)

// Environment variable names consulted as fallbacks when a corresponding
// provider attribute is not set in HCL.
const (
	envBaseURL    = "VIETTELIDC_BASE_URL"
	envCustomerID = "VIETTELIDC_CUSTOMER_ID"
	envToken      = "VIETTELIDC_TOKEN"
	envVpcID      = "VIETTELIDC_VPC_ID"
	envUsername   = "VIETTELIDC_USERNAME"
	envPassword   = "VIETTELIDC_PASSWORD"
	envUserType   = "VIETTELIDC_USER_TYPE"
)

// Compile-time check that CSAProvider satisfies the provider.Provider interface.
var _ provider.Provider = (*CSAProvider)(nil)

// CSAProvider is the Terraform Plugin Framework provider for ViettelIDC IaC.
type CSAProvider struct {
	version string
}

// CSAProviderModel is the Go representation of the provider's HCL config block.
type CSAProviderModel struct {
	BaseURL    types.String `tfsdk:"base_url"`
	CustomerID types.String `tfsdk:"customer_id"`
	Token      types.String `tfsdk:"token"`
	VpcID      types.String `tfsdk:"vpc_id"`
	Username   types.String `tfsdk:"username"`
	Password   types.String `tfsdk:"password"`
	UserType   types.String `tfsdk:"user_type"`
}

// New returns a constructor for the CSA provider. It is consumed by
// providerserver.NewProtocol6 in main.go.
func New() func() provider.Provider {
	return func() provider.Provider {
		return &CSAProvider{version: "0.1.0"}
	}
}

// Metadata sets the provider type name. The name MUST match the VCD provider's
// type name ("viettelidc") because both are muxed under the same provider
// address. Resource and data source name prefixes (vcd_* vs viettelidc_*)
// disambiguate which underlying server handles each request.
func (p *CSAProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "viettelidc"
	resp.Version = p.version
}

// Schema declares the provider-level configuration block. All four attributes
// are marked Optional in schema (rather than Required) because environment
// variables provide a fallback path. Configure() validates the resolved
// (post-env) values.
func (p *CSAProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "ViettelIDC IaC provider (Plugin Framework). Configures the API client used by viettelidc_subnet, viettelidc_network_interface, and related resources.",
		Attributes: map[string]schema.Attribute{
			"base_url": schema.StringAttribute{
				Optional:    true,
				Description: "API Gateway base URL. Mặc định: https://iac.viettelidc.com.vn. Có thể ghi đè qua env " + envBaseURL + ".",
			},
			"customer_id": schema.StringAttribute{
				Optional:    true,
				Description: "ViettelIDC customer/tenant identifier. Falls back to env " + envCustomerID + ".",
			},
			"token": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "Pre-acquired CMP bearer token. When set, both the Authorization and X-Authorization headers use this value verbatim and no login call is made. Falls back to env " + envToken + ".",
			},
			"vpc_id": schema.StringAttribute{
				Optional:    true,
				Description: "Default VPC ID applied when a resource omits its own vpc_id. Falls back to env " + envVpcID + ".",
			},
			"username": schema.StringAttribute{
				Optional:    true,
				Description: "Operator email. Required with 'password' unless 'token' is set. Falls back to env " + envUsername + ".",
			},
			"password": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "Operator password for POST /iam/api/v1/authorization/login. Required with 'username' unless 'token' is set. Falls back to env " + envPassword + ".",
			},
			"user_type": schema.StringAttribute{
				Optional:    true,
				Description: "User type for login (default: ROOT_USER). Falls back to env " + envUserType + ".",
			},
		},
	}
}

// Configure resolves config + env-var fallbacks, builds the HTTP client, and
// publishes a *ProviderData to all resources and data sources.
//
// Auth resolution (in priority order):
//  1. If 'token' is set, use it directly for both Authorization and
//     X-Authorization headers. No login call is made.
//  2. Otherwise, perform the two-step login via POST /iam/api/v1/authorization/login
//     using username + password (+ optional user_type).
//
// base_url and customer_id are always required.
func (p *CSAProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var cfg CSAProviderModel
	diags := req.Config.Get(ctx, &cfg)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	baseURL := stringOrEnv(cfg.BaseURL, envBaseURL)
	customerID := stringOrEnv(cfg.CustomerID, envCustomerID)
	token := stringOrEnv(cfg.Token, envToken)
	vpcID := stringOrEnv(cfg.VpcID, envVpcID)
	username := stringOrEnv(cfg.Username, envUsername)
	password := stringOrEnv(cfg.Password, envPassword)
	userType := stringOrEnv(cfg.UserType, envUserType)

	// Default API base URL khi không cấu hình.
	if baseURL == "" {
		baseURL = "https://iac.viettelidc.com.vn"
	}
	if token == "" && (username == "" || password == "") {
		resp.Diagnostics.AddError("Missing auth credentials",
			"Set 'token' OR both 'username' and 'password' (or their env equivalents).")
	}
	if resp.Diagnostics.HasError() {
		return
	}

	var oldTok, accessTok string
	if token != "" {
		oldTok, accessTok = token, token
	} else {
		var err error
		oldTok, accessTok, err = client.LoginWithPassword(ctx, nil, baseURL, client.LoginCredentials{
			Username: username,
			Password: password,
			UserType: userType,
		})
		if err != nil {
			resp.Diagnostics.AddError("Login failed", err.Error())
			return
		}
	}

	// Auto-resolve customer_id from JWT when not explicitly set.
	if customerID == "" {
		extracted, err := client.ExtractCustomerIDFromJWT(oldTok)
		if err != nil {
			resp.Diagnostics.AddAttributeError(path.Root("customer_id"), "Cannot resolve customer_id",
				fmt.Sprintf("customer_id not set and could not be extracted from JWT: %s", err.Error()))
			return
		}
		if extracted == "" {
			resp.Diagnostics.AddAttributeError(path.Root("customer_id"), "Missing customer_id",
				"customer_id not set and not found in JWT claims. Set 'customer_id' in the provider block or export "+envCustomerID+".")
			return
		}
		customerID = extracted
	}

	pd := &providerdata.ProviderData{
		Client:       client.NewClientWithTokens(baseURL, oldTok, accessTok),
		CustomerID:   customerID,
		DefaultVpcID: vpcID,
	}
	resp.ResourceData = pd
	resp.DataSourceData = pd
}

// stringOrEnv returns the HCL string value, falling back to env var when the
// HCL value is null, unknown, or empty.
func stringOrEnv(v types.String, envName string) string {
	if !v.IsNull() && !v.IsUnknown() {
		if s := v.ValueString(); s != "" {
			return s
		}
	}
	return os.Getenv(envName)
}

// Resources returns the list of CSA resources.
func (p *CSAProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		networking.NewVPCResource,
		networking.NewSubnetResource,
		networking.NewNetworkInterfaceResource,
		networking.NewNetworkInterfaceAttachmentResource,
		networking.NewFloatingIPResource,
		networking.NewSecurityGroupResource,
		networking.NewSecurityGroupRuleResource,
		networking.NewKeyPairResource,
		networking.NewInstanceResource,
		networking.NewVolumeResource,
		networking.NewVolumeAttachmentResource,
		networking.NewRouteTableResource,
		networking.NewRouteTableAssociationResource,
		networking.NewNatGatewayResource,
		networking.NewLoadBalancerResource,
		networking.NewCertificateResource,
		networking.NewBackupPlanResource,
		// Phase 4 — VPC autoscaling resources
		vpc.NewLaunchTemplateResource,
		vpc.NewAutoscaleGroupResource,
	}
}

// DataSources returns the list of CSA data sources.
func (p *CSAProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		networking.NewVPCDataSource,
		networking.NewVFirewallsDataSource,
		networking.NewSubnetDataSource,
		networking.NewSubnetsDataSource,
		networking.NewNetworkInterfaceDataSource,
		networking.NewNetworkInterfacesDataSource,
		networking.NewFloatingIPDataSource,
		networking.NewSecurityGroupDataSource,
		networking.NewInstanceDataSource,
		networking.NewVMTemplatesDataSource,
		networking.NewRouteTableDataSource,
		networking.NewNatGatewayDataSource,
		networking.NewInternetGatewayDataSource,
		networking.NewLoadBalancerDataSource,
		networking.NewCertificateDataSource,
		networking.NewBackupRecordDataSource,
		// Phase 4 — VPC autoscaling data sources
		vpc.NewLaunchTemplateDataSource,
		vpc.NewLaunchTemplatesDataSource,
		vpc.NewAutoscaleGroupsDataSource,
	}
}

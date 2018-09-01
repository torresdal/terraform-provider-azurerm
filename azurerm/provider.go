package azurerm

import (
	"context"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2017-05-10/resources"
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/hashicorp/terraform/helper/mutexkv"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/authentication"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/suppress"
)

// Provider returns a terraform.ResourceProvider.
func Provider() terraform.ResourceProvider {
	var p *schema.Provider
	p = &schema.Provider{
		Schema: map[string]*schema.Schema{
			"subscription_id": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("ARM_SUBSCRIPTION_ID", ""),
			},

			"client_id": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("ARM_CLIENT_ID", ""),
			},

			"client_secret": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("ARM_CLIENT_SECRET", ""),
			},

			"tenant_id": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("ARM_TENANT_ID", ""),
			},

			"environment": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("ARM_ENVIRONMENT", "public"),
			},

			"skip_credentials_validation": {
				Type:        schema.TypeBool,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("ARM_SKIP_CREDENTIALS_VALIDATION", false),
			},

			"skip_provider_registration": {
				Type:        schema.TypeBool,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("ARM_SKIP_PROVIDER_REGISTRATION", false),
			},
			"use_msi": {
				Type:        schema.TypeBool,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("ARM_USE_MSI", false),
			},
			"msi_endpoint": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("ARM_MSI_ENDPOINT", ""),
			},
		},

		DataSourcesMap: map[string]*schema.Resource{
			"azurerm_azuread_application":                   dataSourceArmAzureADApplication(),
			"azurerm_azuread_service_principal":             dataSourceArmActiveDirectoryServicePrincipal(),
			"azurerm_api_management_service":                dataSourceApiManagementService(),
			"azurerm_application_security_group":            dataSourceArmApplicationSecurityGroup(),
			"azurerm_app_service":                           dataSourceArmAppService(),
			"azurerm_app_service_plan":                      dataSourceAppServicePlan(),
			"azurerm_builtin_role_definition":               dataSourceArmBuiltInRoleDefinition(),
			"azurerm_cdn_profile":                           dataSourceArmCdnProfile(),
			"azurerm_client_config":                         dataSourceArmClientConfig(),
			"azurerm_cosmosdb_account":                      dataSourceArmCosmosDBAccount(),
			"azurerm_container_registry":                    dataSourceArmContainerRegistry(),
			"azurerm_data_lake_store":                       dataSourceArmDataLakeStoreAccount(),
			"azurerm_dns_zone":                              dataSourceArmDnsZone(),
			"azurerm_eventhub_namespace":                    dataSourceEventHubNamespace(),
			"azurerm_image":                                 dataSourceArmImage(),
			"azurerm_key_vault":                             dataSourceArmKeyVault(),
			"azurerm_key_vault_access_policy":               dataSourceArmKeyVaultAccessPolicy(),
			"azurerm_key_vault_secret":                      dataSourceArmKeyVaultSecret(),
			"azurerm_kubernetes_cluster":                    dataSourceArmKubernetesCluster(),
			"azurerm_log_analytics_workspace":               dataSourceLogAnalyticsWorkspace(),
			"azurerm_logic_app_workflow":                    dataSourceArmLogicAppWorkflow(),
			"azurerm_managed_disk":                          dataSourceArmManagedDisk(),
			"azurerm_network_interface":                     dataSourceArmNetworkInterface(),
			"azurerm_network_security_group":                dataSourceArmNetworkSecurityGroup(),
			"azurerm_notification_hub":                      dataSourceNotificationHub(),
			"azurerm_notification_hub_namespace":            dataSourceNotificationHubNamespace(),
			"azurerm_platform_image":                        dataSourceArmPlatformImage(),
			"azurerm_public_ip":                             dataSourceArmPublicIP(),
			"azurerm_public_ips":                            dataSourceArmPublicIPs(),
			"azurerm_recovery_services_vault":               dataSourceArmRecoveryServicesVault(),
			"azurerm_resource_group":                        dataSourceArmResourceGroup(),
			"azurerm_role_definition":                       dataSourceArmRoleDefinition(),
			"azurerm_route_table":                           dataSourceArmRouteTable(),
			"azurerm_scheduler_job_collection":              dataSourceArmSchedulerJobCollection(),
			"azurerm_snapshot":                              dataSourceArmSnapshot(),
			"azurerm_storage_account":                       dataSourceArmStorageAccount(),
			"azurerm_storage_account_sas":                   dataSourceArmStorageAccountSharedAccessSignature(),
			"azurerm_subnet":                                dataSourceArmSubnet(),
			"azurerm_subscription":                          dataSourceArmSubscription(),
			"azurerm_subscriptions":                         dataSourceArmSubscriptions(),
			"azurerm_traffic_manager_geographical_location": dataSourceArmTrafficManagerGeographicalLocation(),
			"azurerm_virtual_network":                       dataSourceArmVirtualNetwork(),
			"azurerm_virtual_network_gateway":               dataSourceArmVirtualNetworkGateway(),
		},

		ResourcesMap: map[string]*schema.Resource{
			"azurerm_azuread_application":                     resourceArmActiveDirectoryApplication(),
			"azurerm_azuread_service_principal":               resourceArmActiveDirectoryServicePrincipal(),
			"azurerm_azuread_service_principal_password":      resourceArmActiveDirectoryServicePrincipalPassword(),
			"azurerm_application_gateway":                     resourceArmApplicationGateway(),
			"azurerm_application_insights":                    resourceArmApplicationInsights(),
			"azurerm_application_security_group":              resourceArmApplicationSecurityGroup(),
			"azurerm_app_service":                             resourceArmAppService(),
			"azurerm_app_service_plan":                        resourceArmAppServicePlan(),
			"azurerm_app_service_active_slot":                 resourceArmAppServiceActiveSlot(),
			"azurerm_app_service_custom_hostname_binding":     resourceArmAppServiceCustomHostnameBinding(),
			"azurerm_app_service_slot":                        resourceArmAppServiceSlot(),
			"azurerm_automation_account":                      resourceArmAutomationAccount(),
			"azurerm_automation_credential":                   resourceArmAutomationCredential(),
			"azurerm_automation_runbook":                      resourceArmAutomationRunbook(),
			"azurerm_automation_schedule":                     resourceArmAutomationSchedule(),
			"azurerm_autoscale_setting":                       resourceArmAutoScaleSetting(),
			"azurerm_availability_set":                        resourceArmAvailabilitySet(),
			"azurerm_cdn_endpoint":                            resourceArmCdnEndpoint(),
			"azurerm_cdn_profile":                             resourceArmCdnProfile(),
			"azurerm_container_registry":                      resourceArmContainerRegistry(),
			"azurerm_container_service":                       resourceArmContainerService(),
			"azurerm_container_group":                         resourceArmContainerGroup(),
			"azurerm_cosmosdb_account":                        resourceArmCosmosDBAccount(),
			"azurerm_data_lake_analytics_account":             resourceArmDataLakeAnalyticsAccount(),
			"azurerm_data_lake_analytics_firewall_rule":       resourceArmDataLakeAnalyticsFirewallRule(),
			"azurerm_data_lake_store":                         resourceArmDataLakeStore(),
			"azurerm_data_lake_store_file":                    resourceArmDataLakeStoreFile(),
			"azurerm_data_lake_store_firewall_rule":           resourceArmDataLakeStoreFirewallRule(),
			"azurerm_dns_a_record":                            resourceArmDnsARecord(),
			"azurerm_dns_aaaa_record":                         resourceArmDnsAAAARecord(),
			"azurerm_dns_caa_record":                          resourceArmDnsCaaRecord(),
			"azurerm_dns_cname_record":                        resourceArmDnsCNameRecord(),
			"azurerm_dns_mx_record":                           resourceArmDnsMxRecord(),
			"azurerm_dns_ns_record":                           resourceArmDnsNsRecord(),
			"azurerm_dns_ptr_record":                          resourceArmDnsPtrRecord(),
			"azurerm_dns_srv_record":                          resourceArmDnsSrvRecord(),
			"azurerm_dns_txt_record":                          resourceArmDnsTxtRecord(),
			"azurerm_dns_zone":                                resourceArmDnsZone(),
			"azurerm_eventgrid_topic":                         resourceArmEventGridTopic(),
			"azurerm_eventhub":                                resourceArmEventHub(),
			"azurerm_eventhub_authorization_rule":             resourceArmEventHubAuthorizationRule(),
			"azurerm_eventhub_consumer_group":                 resourceArmEventHubConsumerGroup(),
			"azurerm_eventhub_namespace":                      resourceArmEventHubNamespace(),
			"azurerm_eventhub_namespace_authorization_rule":   resourceArmEventHubNamespaceAuthorizationRule(),
			"azurerm_express_route_circuit":                   resourceArmExpressRouteCircuit(),
			"azurerm_express_route_circuit_authorization":     resourceArmExpressRouteCircuitAuthorization(),
			"azurerm_express_route_circuit_peering":           resourceArmExpressRouteCircuitPeering(),
			"azurerm_function_app":                            resourceArmFunctionApp(),
			"azurerm_image":                                   resourceArmImage(),
			"azurerm_iothub":                                  resourceArmIotHub(),
			"azurerm_key_vault":                               resourceArmKeyVault(),
			"azurerm_key_vault_access_policy":                 resourceArmKeyVaultAccessPolicy(),
			"azurerm_key_vault_certificate":                   resourceArmKeyVaultCertificate(),
			"azurerm_key_vault_key":                           resourceArmKeyVaultKey(),
			"azurerm_key_vault_secret":                        resourceArmKeyVaultSecret(),
			"azurerm_kubernetes_cluster":                      resourceArmKubernetesCluster(),
			"azurerm_lb":                                      resourceArmLoadBalancer(),
			"azurerm_lb_backend_address_pool":                 resourceArmLoadBalancerBackendAddressPool(),
			"azurerm_lb_nat_rule":                             resourceArmLoadBalancerNatRule(),
			"azurerm_lb_nat_pool":                             resourceArmLoadBalancerNatPool(),
			"azurerm_lb_probe":                                resourceArmLoadBalancerProbe(),
			"azurerm_lb_rule":                                 resourceArmLoadBalancerRule(),
			"azurerm_local_network_gateway":                   resourceArmLocalNetworkGateway(),
			"azurerm_log_analytics_solution":                  resourceArmLogAnalyticsSolution(),
			"azurerm_log_analytics_workspace":                 resourceArmLogAnalyticsWorkspace(),
			"azurerm_logic_app_action_custom":                 resourceArmLogicAppActionCustom(),
			"azurerm_logic_app_action_http":                   resourceArmLogicAppActionHTTP(),
			"azurerm_logic_app_trigger_custom":                resourceArmLogicAppTriggerCustom(),
			"azurerm_logic_app_trigger_http_request":          resourceArmLogicAppTriggerHttpRequest(),
			"azurerm_logic_app_trigger_recurrence":            resourceArmLogicAppTriggerRecurrence(),
			"azurerm_logic_app_workflow":                      resourceArmLogicAppWorkflow(),
			"azurerm_managed_disk":                            resourceArmManagedDisk(),
			"azurerm_management_lock":                         resourceArmManagementLock(),
			"azurerm_metric_alertrule":                        resourceArmMetricAlertRule(),
			"azurerm_monitor_action_group":                    resourceArmMonitorActionGroup(),
			"azurerm_mysql_configuration":                     resourceArmMySQLConfiguration(),
			"azurerm_mysql_database":                          resourceArmMySqlDatabase(),
			"azurerm_mysql_firewall_rule":                     resourceArmMySqlFirewallRule(),
			"azurerm_mysql_server":                            resourceArmMySqlServer(),
			"azurerm_network_interface":                       resourceArmNetworkInterface(),
			"azurerm_network_security_group":                  resourceArmNetworkSecurityGroup(),
			"azurerm_network_security_rule":                   resourceArmNetworkSecurityRule(),
			"azurerm_network_watcher":                         resourceArmNetworkWatcher(),
			"azurerm_notification_hub":                        resourceArmNotificationHub(),
			"azurerm_notification_hub_authorization_rule":     resourceArmNotificationHubAuthorizationRule(),
			"azurerm_notification_hub_namespace":              resourceArmNotificationHubNamespace(),
			"azurerm_packet_capture":                          resourceArmPacketCapture(),
			"azurerm_policy_assignment":                       resourceArmPolicyAssignment(),
			"azurerm_policy_definition":                       resourceArmPolicyDefinition(),
			"azurerm_postgresql_configuration":                resourceArmPostgreSQLConfiguration(),
			"azurerm_postgresql_database":                     resourceArmPostgreSQLDatabase(),
			"azurerm_postgresql_firewall_rule":                resourceArmPostgreSQLFirewallRule(),
			"azurerm_postgresql_server":                       resourceArmPostgreSQLServer(),
			"azurerm_public_ip":                               resourceArmPublicIp(),
			"azurerm_relay_namespace":                         resourceArmRelayNamespace(),
			"azurerm_recovery_services_vault":                 resourceArmRecoveryServicesVault(),
			"azurerm_redis_cache":                             resourceArmRedisCache(),
			"azurerm_redis_firewall_rule":                     resourceArmRedisFirewallRule(),
			"azurerm_resource_group":                          resourceArmResourceGroup(),
			"azurerm_role_assignment":                         resourceArmRoleAssignment(),
			"azurerm_role_definition":                         resourceArmRoleDefinition(),
			"azurerm_route":                                   resourceArmRoute(),
			"azurerm_route_table":                             resourceArmRouteTable(),
			"azurerm_search_service":                          resourceArmSearchService(),
			"azurerm_servicebus_namespace":                    resourceArmServiceBusNamespace(),
			"azurerm_servicebus_namespace_authorization_rule": resourceArmServiceBusNamespaceAuthorizationRule(),
			"azurerm_servicebus_queue":                        resourceArmServiceBusQueue(),
			"azurerm_servicebus_queue_authorization_rule":     resourceArmServiceBusQueueAuthorizationRule(),
			"azurerm_servicebus_subscription":                 resourceArmServiceBusSubscription(),
			"azurerm_servicebus_subscription_rule":            resourceArmServiceBusSubscriptionRule(),
			"azurerm_servicebus_topic":                        resourceArmServiceBusTopic(),
			"azurerm_servicebus_topic_authorization_rule":     resourceArmServiceBusTopicAuthorizationRule(),
			"azurerm_service_fabric_cluster":                  resourceArmServiceFabricCluster(),
			"azurerm_snapshot":                                resourceArmSnapshot(),
			"azurerm_scheduler_job":                           resourceArmSchedulerJob(),
			"azurerm_scheduler_job_collection":                resourceArmSchedulerJobCollection(),
			"azurerm_sql_database":                            resourceArmSqlDatabase(),
			"azurerm_sql_elasticpool":                         resourceArmSqlElasticPool(),
			"azurerm_sql_firewall_rule":                       resourceArmSqlFirewallRule(),
			"azurerm_sql_active_directory_administrator":      resourceArmSqlAdministrator(),
			"azurerm_sql_server":                              resourceArmSqlServer(),
			"azurerm_sql_virtual_network_rule":                resourceArmSqlVirtualNetworkRule(),
			"azurerm_storage_account":                         resourceArmStorageAccount(),
			"azurerm_storage_blob":                            resourceArmStorageBlob(),
			"azurerm_storage_container":                       resourceArmStorageContainer(),
			"azurerm_storage_share":                           resourceArmStorageShare(),
			"azurerm_storage_queue":                           resourceArmStorageQueue(),
			"azurerm_storage_table":                           resourceArmStorageTable(),
			"azurerm_subnet":                                  resourceArmSubnet(),
			"azurerm_template_deployment":                     resourceArmTemplateDeployment(),
			"azurerm_traffic_manager_endpoint":                resourceArmTrafficManagerEndpoint(),
			"azurerm_traffic_manager_profile":                 resourceArmTrafficManagerProfile(),
			"azurerm_user_assigned_identity":                  resourceArmUserAssignedIdentity(),
			"azurerm_virtual_machine":                         resourceArmVirtualMachine(),
			"azurerm_virtual_machine_data_disk_attachment":    resourceArmVirtualMachineDataDiskAttachment(),
			"azurerm_virtual_machine_extension":               resourceArmVirtualMachineExtensions(),
			"azurerm_virtual_machine_scale_set":               resourceArmVirtualMachineScaleSet(),
			"azurerm_virtual_network":                         resourceArmVirtualNetwork(),
			"azurerm_virtual_network_gateway":                 resourceArmVirtualNetworkGateway(),
			"azurerm_virtual_network_gateway_connection":      resourceArmVirtualNetworkGatewayConnection(),
			"azurerm_virtual_network_peering":                 resourceArmVirtualNetworkPeering(),
		},
	}

	p.ConfigureFunc = providerConfigure(p)

	return p
}

func providerConfigure(p *schema.Provider) schema.ConfigureFunc {
	return func(d *schema.ResourceData) (interface{}, error) {
		config := &authentication.Config{
			SubscriptionID:            d.Get("subscription_id").(string),
			ClientID:                  d.Get("client_id").(string),
			ClientSecret:              d.Get("client_secret").(string),
			TenantID:                  d.Get("tenant_id").(string),
			Environment:               d.Get("environment").(string),
			UseMsi:                    d.Get("use_msi").(bool),
			MsiEndpoint:               d.Get("msi_endpoint").(string),
			SkipCredentialsValidation: d.Get("skip_credentials_validation").(bool),
			SkipProviderRegistration:  d.Get("skip_provider_registration").(bool),
		}

		if config.UseMsi {
			log.Printf("[DEBUG] use_msi specified - using MSI Authentication")
			if config.MsiEndpoint == "" {
				msiEndpoint, err := adal.GetMSIVMEndpoint()
				if err != nil {
					return nil, fmt.Errorf("Could not retrieve MSI endpoint from VM settings."+
						"Ensure the VM has MSI enabled, or try setting msi_endpoint. Error: %s", err)
				}
				config.MsiEndpoint = msiEndpoint
			}
			log.Printf("[DEBUG] Using MSI endpoint %s", config.MsiEndpoint)
			if err := config.ValidateMsi(); err != nil {
				return nil, err
			}
		} else if config.ClientSecret != "" {
			log.Printf("[DEBUG] Client Secret specified - using Service Principal for Authentication")
			if err := config.ValidateServicePrincipal(); err != nil {
				return nil, err
			}
		} else {
			log.Printf("[DEBUG] No Client Secret specified - loading credentials from Azure CLI")
			if err := config.LoadTokensFromAzureCLI(); err != nil {
				return nil, err
			}

			if err := config.ValidateBearerAuth(); err != nil {
				return nil, fmt.Errorf("Please specify either a Service Principal, or log in with the Azure CLI (using `az login`)")
			}
		}

		client, err := getArmClient(config)
		if err != nil {
			return nil, err
		}

		client.StopContext = p.StopContext()

		// replaces the context between tests
		p.MetaReset = func() error {
			client.StopContext = p.StopContext()
			return nil
		}

		if !config.SkipCredentialsValidation {
			// List all the available providers and their registration state to avoid unnecessary
			// requests. This also lets us check if the provider credentials are correct.
			ctx := client.StopContext
			providerList, err := client.providersClient.List(ctx, nil, "")
			if err != nil {
				return nil, fmt.Errorf("Unable to list provider registration status, it is possible that this is due to invalid "+
					"credentials or the service principal does not have permission to use the Resource Manager API, Azure "+
					"error: %s", err)
			}

			if !config.SkipProviderRegistration {
				err = registerAzureResourceProvidersWithSubscription(ctx, providerList.Values(), client.providersClient)
				if err != nil {
					return nil, err
				}
			}
		}

		return client, nil
	}
}

func registerProviderWithSubscription(ctx context.Context, providerName string, client resources.ProvidersClient) error {
	_, err := client.Register(ctx, providerName)
	if err != nil {
		return fmt.Errorf("Cannot register provider %s with Azure Resource Manager: %s.", providerName, err)
	}

	return nil
}

func determineAzureResourceProvidersToRegister(providerList []resources.Provider) map[string]struct{} {
	providers := map[string]struct{}{
		"Microsoft.Authorization":       {},
		"Microsoft.Automation":          {},
		"Microsoft.Cache":               {},
		"Microsoft.Cdn":                 {},
		"Microsoft.Compute":             {},
		"Microsoft.ContainerInstance":   {},
		"Microsoft.ContainerRegistry":   {},
		"Microsoft.ContainerService":    {},
		"Microsoft.DataLakeStore":       {},
		"Microsoft.DBforMySQL":          {},
		"Microsoft.DBforPostgreSQL":     {},
		"Microsoft.Devices":             {},
		"Microsoft.DocumentDB":          {},
		"Microsoft.EventGrid":           {},
		"Microsoft.EventHub":            {},
		"Microsoft.KeyVault":            {},
		"microsoft.insights":            {},
		"Microsoft.Logic":               {},
		"Microsoft.ManagedIdentity":     {},
		"Microsoft.Network":             {},
		"Microsoft.NotificationHubs":    {},
		"Microsoft.OperationalInsights": {},
		"Microsoft.Relay":               {},
		"Microsoft.Resources":           {},
		"Microsoft.Search":              {},
		"Microsoft.ServiceBus":          {},
		"Microsoft.ServiceFabric":       {},
		"Microsoft.Sql":                 {},
		"Microsoft.Storage":             {},
	}

	// filter out any providers already registered
	for _, p := range providerList {
		if _, ok := providers[*p.Namespace]; !ok {
			continue
		}

		if strings.ToLower(*p.RegistrationState) == "registered" {
			log.Printf("[DEBUG] Skipping provider registration for namespace %s\n", *p.Namespace)
			delete(providers, *p.Namespace)
		}
	}

	return providers
}

// registerAzureResourceProvidersWithSubscription uses the providers client to register
// all Azure resource providers which the Terraform provider may require (regardless of
// whether they are actually used by the configuration or not). It was confirmed by Microsoft
// that this is the approach their own internal tools also take.
func registerAzureResourceProvidersWithSubscription(ctx context.Context, providerList []resources.Provider, client resources.ProvidersClient) error {
	providers := determineAzureResourceProvidersToRegister(providerList)

	var err error
	var wg sync.WaitGroup
	wg.Add(len(providers))

	for providerName := range providers {
		go func(p string) {
			defer wg.Done()
			log.Printf("[DEBUG] Registering provider with namespace %s\n", p)
			if innerErr := registerProviderWithSubscription(ctx, p, client); err != nil {
				err = innerErr
			}
		}(providerName)
	}

	wg.Wait()

	return err
}

// armMutexKV is the instance of MutexKV for ARM resources
var armMutexKV = mutexkv.NewMutexKV()

// Deprecated: use `suppress.CaseDifference` instead
func ignoreCaseDiffSuppressFunc(k, old, new string, d *schema.ResourceData) bool {
	return suppress.CaseDifference(k, old, new, d)
}

// ignoreCaseStateFunc is a StateFunc from helper/schema that converts the
// supplied value to lower before saving to state for consistency.
func ignoreCaseStateFunc(val interface{}) string {
	return strings.ToLower(val.(string))
}

func userDataDiffSuppressFunc(k, old, new string, d *schema.ResourceData) bool {
	return userDataStateFunc(old) == new
}

func userDataStateFunc(v interface{}) string {
	switch s := v.(type) {
	case string:
		s = base64Encode(s)
		hash := sha1.Sum([]byte(s))
		return hex.EncodeToString(hash[:])
	default:
		return ""
	}
}

// base64Encode encodes data if the input isn't already encoded using
// base64.StdEncoding.EncodeToString. If the input is already base64 encoded,
// return the original input unchanged.
func base64Encode(data string) string {
	// Check whether the data is already Base64 encoded; don't double-encode
	if isBase64Encoded(data) {
		return data
	}
	// data has not been encoded encode and return
	return base64.StdEncoding.EncodeToString([]byte(data))
}

func isBase64Encoded(data string) bool {
	_, err := base64.StdEncoding.DecodeString(data)
	return err == nil
}

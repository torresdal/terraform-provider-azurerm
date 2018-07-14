package azurerm

import (
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/apimanagement/mgmt/2017-03-01/apimanagement"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
)

func resourceArmApiManagementService() *schema.Resource {
	return &schema.Resource{
		Create: resourceArmApiManagementServiceCreateUpdate,
		Read:   resourceArmApiManagementServiceRead,
		Update: resourceArmApiManagementServiceCreateUpdate,
		Delete: resourceArmApiManagementDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateApiManagementName,
			},

			"resource_group_name": resourceGroupNameSchema(),

			"location": locationSchema(),

			"publisher_name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateApiManagementPublisherName,
			},

			"publisher_email": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateApiManagementPublisherEmail,
			},

			"sku": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  string(apimanagement.SkuTypeDeveloper),
							ValidateFunc: validation.StringInSlice([]string{
								string(apimanagement.SkuTypeDeveloper),
								string(apimanagement.SkuTypeBasic),
								string(apimanagement.SkuTypeStandard),
								string(apimanagement.SkuTypePremium),
							}, true),
							DiffSuppressFunc: ignoreCaseDiffSuppressFunc,
						},
						"capacity": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  1,
						},
					},
				},
			},

			"notification_sender_email": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"additional_location": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"location": locationSchema(),

						"gateway_regional_url": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"static_ips": {
							Type: schema.TypeList,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Computed: true,
						},
					},
				},
			},

			"certificate": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 10,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"encoded_certificate": {
							Type:     schema.TypeString,
							Required: true,
						},

						"certificate_password": {
							Type:      schema.TypeString,
							Required:  true,
							Sensitive: true,
						},

						"store_name": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								string(apimanagement.CertificateAuthority),
								string(apimanagement.Root),
							}, true),
							DiffSuppressFunc: ignoreCaseDiffSuppressFunc,
						},
					},
				},
			},

			"custom_properties": {
				Type:     schema.TypeMap,
				Optional: true,
			},

			"hostname_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								string(apimanagement.Management),
								string(apimanagement.Portal),
								string(apimanagement.Proxy),
								string(apimanagement.Scm),
							}, true),
							DiffSuppressFunc: ignoreCaseDiffSuppressFunc,
						},

						"host_name": {
							Type:     schema.TypeString,
							Required: true,
						},

						"certificate": {
							Type:     schema.TypeString,
							Required: true,
						},

						"certificate_password": {
							Type:      schema.TypeString,
							Required:  true,
							Sensitive: true,
						},

						"default_ssl_binding": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},

						"negotiate_client_certificate": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
					},
				},
			},

			"tags": tagsSchema(),

			"created": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"gateway_url": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"gateway_regional_url": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"portal_url": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"management_api_url": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"scm_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceArmApiManagementServiceCreateUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).apiManagementServiceClient
	ctx := meta.(*ArmClient).StopContext

	log.Printf("[INFO] preparing arguments for AzureRM API Management creation.")

	resGroup := d.Get("resource_group_name").(string)
	name := d.Get("name").(string)
	location := azureRMNormalizeLocation(d.Get("location").(string))
	tags := d.Get("tags").(map[string]interface{})

	var sku *apimanagement.ServiceSkuProperties
	if skuConfig := d.Get("sku").([]interface{}); skuConfig != nil {
		sku = expandAzureRmApiManagementSku(skuConfig)
	}

	properties := expandAzureRmApiManagementProperties(d, sku)

	apiManagement := apimanagement.ServiceResource{
		Location:          &location,
		ServiceProperties: properties,
		Tags:              expandTags(tags),
		Sku:               sku,
	}

	createFuture, err := client.CreateOrUpdate(ctx, resGroup, name, apiManagement)
	if err != nil {
		return err
	}

	err = createFuture.WaitForCompletion(ctx, client.Client)
	if err != nil {
		return err
	}

	read, err := client.Get(ctx, resGroup, name)
	if err != nil {
		return err
	}

	if read.ID == nil {
		return fmt.Errorf("Cannot read AzureRM Api Management %q (resource group %q) ID", name, resGroup)
	}

	d.SetId(*read.ID)

	return resourceArmApiManagementServiceRead(d, meta)
}

func resourceArmApiManagementServiceRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient)
	apiManagementClient := meta.(*ArmClient).apiManagementServiceClient

	id, err := parseAzureResourceID(d.Id())
	if err != nil {
		return err
	}

	resGroup := id.ResourceGroup
	name := id.Path["service"]

	ctx := client.StopContext
	resp, err := apiManagementClient.Get(ctx, resGroup, name)

	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error making Read request on API Management Service %q (Resource Group %q): %+v", name, resGroup, err)
	}

	d.Set("name", name)
	d.Set("resource_group_name", resGroup)

	if location := resp.Location; location != nil {
		d.Set("location", azureRMNormalizeLocation(*location))
	}

	if props := resp.ServiceProperties; props != nil {
		d.Set("publisher_email", props.PublisherEmail)
		d.Set("publisher_name", props.PublisherName)

		d.Set("notification_sender_email", props.NotificationSenderEmail)
		d.Set("created", props.CreatedAtUtc.Format(time.RFC3339))
		d.Set("gateway_url", props.GatewayURL)
		d.Set("gateway_regional_url", props.GatewayRegionalURL)
		d.Set("portal_url", props.PortalURL)
		d.Set("management_api_url", props.ManagementAPIURL)
		d.Set("scm_url", props.ScmURL)
		d.Set("static_ips", props.StaticIps)
		d.Set("custom_properties", &props.CustomProperties)

		hostnameConfigurations, err := flattenApiManagementHostnameConfigurations(d, props.HostnameConfigurations)

		if err != nil {
			return err
		}

		if err := d.Set("hostname_configuration", hostnameConfigurations); err != nil {
			return fmt.Errorf("Error setting `hostname_configuration`: %+v", err)
		}

		additionalLocations := flattenApiManagementAdditionalLocations(props.AdditionalLocations)
		if err := d.Set("additional_location", additionalLocations); err != nil {
			return fmt.Errorf("Error setting `additional_location`: %+v", err)
		}

		certificates := flattenApiManagementCertificates(d, props.Certificates)

		if err != nil {
			return err
		}

		if err := d.Set("certificate", certificates); err != nil {
			return fmt.Errorf("Error setting `certificate`: %+v", err)
		}
	}

	if sku := resp.Sku; sku != nil {
		if err := d.Set("sku", flattenApiManagementServiceSku(sku)); err != nil {
			return fmt.Errorf("Error flattening `sku`: %+v", err)
		}
	}

	flattenAndSetTags(d, resp.Tags)

	return nil
}

func resourceArmApiManagementDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).apiManagementServiceClient
	ctx := meta.(*ArmClient).StopContext

	id, err := parseAzureResourceID(d.Id())
	if err != nil {
		return err
	}
	resGroup := id.ResourceGroup
	name := id.Path["service"]

	log.Printf("[DEBUG] Deleting api management %s: %s", resGroup, name)

	resp, err := client.Delete(ctx, resGroup, name)

	if err != nil {
		if utils.ResponseWasNotFound(resp) {
			return nil
		}

		return err
	}

	return nil
}

func expandAzureRmApiManagementProperties(d *schema.ResourceData, sku *apimanagement.ServiceSkuProperties) *apimanagement.ServiceProperties {
	publisher_name := d.Get("publisher_name").(string)
	publisher_email := d.Get("publisher_email").(string)
	notification_sender_email := d.Get("notification_sender_email").(string)

	custom_properties := expandApiManagementCustomProperties(d)

	additional_locations := expandAzureRmApiManagementAdditionalLocations(d, sku)
	certificates := expandAzureRmApiManagementCertificates(d)
	hostname_configurations := expandAzureRmApiManagementHostnameConfigurations(d)

	properties := apimanagement.ServiceProperties{
		PublisherName:          utils.String(publisher_name),
		PublisherEmail:         utils.String(publisher_email),
		CustomProperties:       custom_properties,
		AdditionalLocations:    additional_locations,
		Certificates:           certificates,
		HostnameConfigurations: hostname_configurations,
	}

	if notification_sender_email != "" {
		properties.NotificationSenderEmail = &notification_sender_email
	}

	return &properties
}

func expandApiManagementCustomProperties(d *schema.ResourceData) map[string]*string {
	input := d.Get("custom_properties").(map[string]interface{})
	output := make(map[string]*string, len(input))

	if input == nil {
		return nil
	}

	for k, v := range input {
		output[k] = utils.String(v.(string))
	}

	return output
}

func expandAzureRmApiManagementHostnameConfigurations(d *schema.ResourceData) *[]apimanagement.HostnameConfiguration {
	hostnameConfigs := d.Get("hostname_configuration").([]interface{})

	if hostnameConfigs == nil || len(hostnameConfigs) == 0 {
		return nil
	}

	hostnames := make([]apimanagement.HostnameConfiguration, 0, len(hostnameConfigs))

	for _, v := range hostnameConfigs {
		config := v.(map[string]interface{})

		host_type := apimanagement.HostnameType(config["type"].(string))
		host_name := config["host_name"].(string)
		certificate := config["certificate"].(string)
		certificate_password := config["certificate_password"].(string)
		default_ssl_binding := config["default_ssl_binding"].(bool)
		negotiate_client_certificate := config["negotiate_client_certificate"].(bool)

		hostname := apimanagement.HostnameConfiguration{
			Type:                       host_type,
			HostName:                   &host_name,
			EncodedCertificate:         &certificate,
			CertificatePassword:        &certificate_password,
			DefaultSslBinding:          &default_ssl_binding,
			NegotiateClientCertificate: &negotiate_client_certificate,
		}

		hostnames = append(hostnames, hostname)
	}

	return &hostnames
}

func expandAzureRmApiManagementCertificates(d *schema.ResourceData) *[]apimanagement.CertificateConfiguration {
	certConfigs := d.Get("certificate").([]interface{})

	if certConfigs == nil || len(certConfigs) == 0 {
		return nil
	}

	certificates := make([]apimanagement.CertificateConfiguration, 0)

	for _, v := range certConfigs {
		config := v.(map[string]interface{})

		cert_base64 := config["encoded_certificate"].(string)
		certificate_password := config["certificate_password"].(string)
		store_name := apimanagement.StoreName(config["store_name"].(string))

		cert := apimanagement.CertificateConfiguration{
			EncodedCertificate:  &cert_base64,
			CertificatePassword: &certificate_password,
			StoreName:           store_name,
		}

		certificates = append(certificates, cert)
	}

	return &certificates
}

func expandAzureRmApiManagementAdditionalLocations(d *schema.ResourceData, sku *apimanagement.ServiceSkuProperties) *[]apimanagement.AdditionalLocation {
	inputLocations := d.Get("additional_location").([]interface{})

	if inputLocations == nil || len(inputLocations) == 0 {
		return nil
	}

	additionalLocations := make([]apimanagement.AdditionalLocation, 0)

	for _, v := range inputLocations {
		config := v.(map[string]interface{})
		location := config["location"].(string)

		additionalLocation := apimanagement.AdditionalLocation{
			Location: &location,
			Sku:      sku,
		}

		additionalLocations = append(additionalLocations, additionalLocation)
	}

	return &additionalLocations
}

func expandAzureRmApiManagementSku(configs []interface{}) *apimanagement.ServiceSkuProperties {
	config := configs[0].(map[string]interface{})

	nameConfig := config["name"].(string)
	name := apimanagement.SkuType(nameConfig)
	capacity := int32(config["capacity"].(int))

	sku := &apimanagement.ServiceSkuProperties{
		Name:     name,
		Capacity: &capacity,
	}

	return sku
}

func flattenApiManagementCertificates(d *schema.ResourceData, props *[]apimanagement.CertificateConfiguration) []interface{} {
	certificates := make([]interface{}, 0, 1)

	if props != nil {
		for i, prop := range *props {
			certificate := make(map[string]interface{}, 2)

			if prop.StoreName != "" {
				certificate["store_name"] = string(prop.StoreName)
			}

			// certificate password isn't returned, so let's look it up
			passwKey := fmt.Sprintf("certificate.%d.certificate_password", i)
			if v, ok := d.GetOk(passwKey); ok {
				password := v.(string)
				certificate["certificate_password"] = password
			}

			// encoded certificate isn't returned, so let's look it up
			certKey := fmt.Sprintf("certificate.%d.encoded_certificate", i)
			if v, ok := d.GetOk(certKey); ok {
				cert := v.(string)
				certificate["encoded_certificate"] = cert
			}

			certificates = append(certificates, certificate)
		}
	}

	return certificates
}

func flattenApiManagementAdditionalLocations(props *[]apimanagement.AdditionalLocation) []interface{} {
	additional_locations := make([]interface{}, 0, 1)

	if props != nil {
		for _, prop := range *props {
			additional_location := make(map[string]interface{}, 2)

			if prop.Location != nil {
				additional_location["location"] = *prop.Location
			}

			if prop.StaticIps != nil {
				additional_location["static_ips"] = *prop.StaticIps
			}

			if prop.GatewayRegionalURL != nil {
				additional_location["gateway_regional_url"] = *prop.GatewayRegionalURL
			}

			additional_locations = append(additional_locations, additional_location)
		}
	}

	return additional_locations
}

func flattenApiManagementHostnameConfigurations(d *schema.ResourceData, configs *[]apimanagement.HostnameConfiguration) ([]interface{}, error) {
	host_configs := make([]interface{}, 0, 1)

	if configs != nil {
		for i, config := range *configs {
			host_config := make(map[string]interface{}, 2)

			if config.Type != "" {
				host_config["type"] = string(config.Type)
			}

			if config.HostName != nil {
				host_config["host_name"] = *config.HostName
			}

			if config.DefaultSslBinding != nil {
				host_config["default_ssl_binding"] = *config.DefaultSslBinding
			}

			if config.NegotiateClientCertificate != nil {
				host_config["negotiate_client_certificate"] = bool(*config.NegotiateClientCertificate)
			}

			// certificate password isn't returned, so let's look it up
			passKey := fmt.Sprintf("hostname_configuration.%d.certificate_password", i)
			if v, ok := d.GetOk(passKey); ok {
				password := v.(string)
				host_config["certificate_password"] = password
			}

			// encoded certificate isn't returned, so let's look it up
			certKey := fmt.Sprintf("hostname_configuration.%d.certificate", i)
			if v, ok := d.GetOk(certKey); ok {
				cert := v.(string)
				host_config["certificate"] = cert
			}

			host_configs = append(host_configs, host_config)
		}
	}

	return host_configs, nil
}

func flattenApiManagementServiceSku(profile *apimanagement.ServiceSkuProperties) []interface{} {
	skus := make([]interface{}, 0, 1)
	sku := make(map[string]interface{}, 2)

	if profile != nil {
		if profile.Name != "" {
			sku["name"] = string(profile.Name)
		}

		if profile.Capacity != nil {
			sku["capacity"] = *profile.Capacity
		}

		skus = append(skus, sku)
	}

	return skus
}

func validateApiManagementName(v interface{}, k string) (ws []string, es []error) {
	value := v.(string)

	if matched := regexp.MustCompile(`^[0-9a-zA-Z-]{1,50}$`).Match([]byte(value)); !matched {
		es = append(es, fmt.Errorf("%q may only contain alphanumeric characters and dashes up to 50 characters in length", k))
	}

	return
}

func validateApiManagementPublisherName(v interface{}, k string) (ws []string, es []error) {
	value := v.(string)

	if matched := regexp.MustCompile(`^[\S*]{1,100}$`).Match([]byte(value)); !matched {
		es = append(es, fmt.Errorf("%q may only be up to 100 characters in length", k))
	}

	return
}

func validateApiManagementPublisherEmail(v interface{}, k string) (ws []string, es []error) {
	value := v.(string)

	if matched := regexp.MustCompile(`^[\S*]{1,100}$`).Match([]byte(value)); !matched {
		es = append(es, fmt.Errorf("%q may only be up to 100 characters in length", k))
	}

	return
}

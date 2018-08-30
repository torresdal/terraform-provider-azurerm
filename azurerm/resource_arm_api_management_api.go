package azurerm

import (
	"fmt"
	"log"
	"strings"

	"github.com/Azure/azure-sdk-for-go/services/apimanagement/mgmt/2017-03-01/apimanagement"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
)

func resourceArmApiManagementApi() *schema.Resource {
	return &schema.Resource{
		Create: resourceArmApiManagementApiCreateUpdate,
		Read:   resourceArmApiManagementApiRead,
		Update: resourceArmApiManagementApiCreateUpdate,
		Delete: resourceArmApiManagementApiDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"service_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"path": {
				Type:     schema.TypeString,
				Required: true,
			},

			"resource_group_name": resourceGroupNameSchema(),

			"location": locationSchema(),

			"service_url": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"import": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"content_value": {
							Type:     schema.TypeString,
							Required: true,
						},
						"content_format": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								string(apimanagement.SwaggerJSON),
								string(apimanagement.SwaggerLinkJSON),
								string(apimanagement.WadlLinkJSON),
								string(apimanagement.WadlXML),
								string(apimanagement.Wsdl),
								string(apimanagement.WsdlLink),
							}, true),
							DiffSuppressFunc: ignoreCaseDiffSuppressFunc,
						},

						"wsdl_selector": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"service_name": {
										Type:     schema.TypeString,
										Required: true,
									},

									"endpoint_name": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
					},
				},
			},

			"display_name": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"protocols": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional: true,
			},

			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"oauth": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"authorization_server_id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"scope": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},

			"subscription_key": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"header": {
							Type:     schema.TypeString,
							Required: true,
						},
						"query": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},

			"soap_api_type": {
				Type:    schema.TypeString,
				Default: "",
				ValidateFunc: validation.StringInSlice([]string{
					string(apimanagement.HTTP),
					string(apimanagement.Soap),
				}, true),
				DiffSuppressFunc: ignoreCaseDiffSuppressFunc,
				Optional:         true,
			},

			"revision_description": {
				Type:     schema.TypeString,
				Optional: true,
				// Default:  1,
			},

			// apiVersion 							- Indicates the Version identifier of the API if the API is versioned
			// apiVersionSetId  				- A resource identifier for the related ApiVersionSet.
			// properties.apiVersionSet - An API Version Set contains the common configuration for a set of API Versions relating
			//
			"version": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"api_version_set_id": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"is_current": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"is_online": {
				Type:     schema.TypeBool,
				Computed: true,
			},
		},
	}
}

func resourceArmApiManagementApiCreateUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).apiManagementApiClient
	ctx := meta.(*ArmClient).StopContext

	log.Printf("[INFO] preparing arguments for AzureRM API Management API creation.")

	resGroup := d.Get("resource_group_name").(string)
	serviceName := d.Get("service_name").(string)
	apiId := d.Get("name").(string)

	var properties *apimanagement.APICreateOrUpdateProperties

	var err error
	isImport := false

	if _, ok := d.GetOk("import"); ok {
		isImport = true
		properties, err = expandApiManagementImportProperties(d)
		if err != nil {
			return err
		}
	} else {
		properties, err = expandApiManagementApiProperties(d)
		if err != nil {
			return err
		}
	}

	apiParams := apimanagement.APICreateOrUpdateParameter{
		APICreateOrUpdateProperties: properties,
	}

	log.Printf("[DEBUG] Calling api with resource group %q, service name %q, api id %q", resGroup, serviceName, apiId)
	log.Printf("[DEBUG] Listing api params:")
	log.Printf("%+v\n", apiParams.APICreateOrUpdateProperties)

	apiContract, err := client.CreateOrUpdate(ctx, resGroup, serviceName, apiId, apiParams, "")
	if err != nil {
		log.Printf("Oh no!!! %+v", err)
		return err
	}

	if v, ok := d.GetOk("service_url"); isImport && ok {
		serviceURL := v.(string)
		updateProps := apimanagement.APIContractUpdateProperties{
			ServiceURL: &serviceURL,
		}

		updateParams := apimanagement.APIUpdateContract{
			APIContractUpdateProperties: &updateProps,
		}
		_, err := client.Update(ctx, resGroup, serviceName, apiId, updateParams, "")
		if err != nil {
			log.Printf("Oh no!!! %+v", err)
			return err
		}
	}

	d.SetId(*apiContract.ID)

	return resourceArmApiManagementApiRead(d, meta)
}

func resourceArmApiManagementApiRead(d *schema.ResourceData, meta interface{}) error {
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

	return nil
}

func resourceArmApiManagementApiDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).apiManagementApiClient
	ctx := meta.(*ArmClient).StopContext

	id, err := parseAzureResourceID(d.Id())
	if err != nil {
		return err
	}

	resGroup := id.ResourceGroup
	serviceName := id.Path["service"]
	apiid := id.Path["apis"]

	log.Printf("[DEBUG] Deleting api management api %s: %s", resGroup, apiid)

	resp, err := client.Delete(ctx, resGroup, serviceName, apiid, "*")

	if err != nil {
		if utils.ResponseWasNotFound(resp) {
			return nil
		}

		return err
	}

	return nil
}

func expandApiManagementApiProperties(d *schema.ResourceData) (*apimanagement.APICreateOrUpdateProperties, error) {
	// revision := d.Get("revision").(string)
	// soapApiType := d.Get("soap_api_type").(string)
	// version := d.Get("version").(string)
	path := d.Get("path").(string)
	serviceUrl := d.Get("service_url").(string)
	displayName := d.Get("display_name").(string)
	protocolsConfig := d.Get("protocols").([]interface{})
	description := d.Get("description").(string)
	soapApiTypeConfig := d.Get("soap_api_type").(string)
	// revisionDescription := d.Get("revision_description").(string)

	protos := make([]apimanagement.Protocol, 0)

	for _, v := range protocolsConfig {
		switch p := strings.ToLower(v.(string)); p {
		case "http":
			protos = append(protos, apimanagement.ProtocolHTTP)
		case "https":
			protos = append(protos, apimanagement.ProtocolHTTPS)
		default:
			return nil, fmt.Errorf("Error expanding protocols. Valid protocols are `http` and `https`.")
		}
	}

	if len(protos) == 0 {
		protos = append(protos, apimanagement.ProtocolHTTPS)
	}

	var soapApiType apimanagement.APIType

	switch s := strings.ToLower(soapApiTypeConfig); s {
	case "http":
		soapApiType = apimanagement.HTTP
	case "soap":
		soapApiType = apimanagement.Soap
	}

	// versionSetId := d.Get("api_version_set_id").(string)

	// var oAuth *apimanagement.OAuth2AuthenticationSettingsContract
	// if oauthConfig := d.Get("oauth").([]interface{}); oauthConfig != nil && len(oauthConfig) > 0 {
	// 	oAuth = expandApiManagementApiOAuth(oauthConfig)
	// }

	log.Printf("ServiceURL: %s", &serviceUrl)

	return &apimanagement.APICreateOrUpdateProperties{
		// APIRevision: &revision,
		APIType:     soapApiType,
		DisplayName: &displayName,
		ServiceURL:  &serviceUrl,
		Description: &description,
		// APIRevision: &revisionDescription,
		// APIVersion:  &version,
		// APIVersionSetID: &versionSetId,
		// AuthenticationSettings: &apimanagement.AuthenticationSettingsContract{
		// 	OAuth2: oAuth,
		// },
		Path:      &path,
		Protocols: &protos,
	}, nil
}

func expandApiManagementImportProperties(d *schema.ResourceData) (*apimanagement.APICreateOrUpdateProperties, error) {
	path := d.Get("path").(string)

	var contentFormat apimanagement.ContentFormat
	if v, ok := d.GetOk("import.0.content_format"); ok {
		contentFormat = apimanagement.ContentFormat(v.(string))
	}

	var contentValue string
	if v, ok := d.GetOk("import.0.content_value"); ok {
		contentValue = v.(string)
	}

	return &apimanagement.APICreateOrUpdateProperties{
		// APIRevision: &revision,
		// APIType:     soapApiType,
		// DisplayName: &displayName,
		// ServiceURL:  &serviceUrl,
		// Description: &description,
		// APIRevision: &revisionDescription,
		// APIVersion:  &version,
		// APIVersionSetID: &versionSetId,
		// AuthenticationSettings: &apimanagement.AuthenticationSettingsContract{
		// 	OAuth2: oAuth,
		// },
		Path: &path,
		// Protocols:     &protos,
		ContentFormat: contentFormat,
		ContentValue:  &contentValue,
	}, nil
}

func expandApiManagementApiOAuth(oauth []interface{}) *apimanagement.OAuth2AuthenticationSettingsContract {
	config := oauth[0].(map[string]interface{})

	authorization_server_id := config["authorization_server_id"].(string)
	scope := config["scope"].(string)

	return &apimanagement.OAuth2AuthenticationSettingsContract{
		AuthorizationServerID: &authorization_server_id,
		Scope: &scope,
	}
}

func flattenApiManagementApiContract(apiContract apimanagement.APIContract) error {
	return nil
}

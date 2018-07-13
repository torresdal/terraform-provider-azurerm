package azurerm

import (
	"fmt"
	"log"

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
			"service_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"api_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
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

			"service_url": {
				Type:     schema.TypeString,
				Required: true,
			},

			"path": {
				Type:     schema.TypeString,
				Required: true,
			},

			"protocols": {
				Type: schema.TypeList,
			},

			"api_version_set_id": {
				Type:     schema.TypeString,
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

			"api_type": {
				Type:    schema.TypeString,
				Default: apimanagement.HTTP,
				ValidateFunc: validation.StringInSlice([]string{
					string(apimanagement.HTTP),
					string(apimanagement.Soap),
				}, true),
				DiffSuppressFunc: ignoreCaseDiffSuppressFunc,
			},

			"revision": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  1,
			},

			"version": {
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
	apiId := d.Get("api_id").(string)

	properties := expandApiManagementApiProperties(d)

	apiParams := apimanagement.APICreateOrUpdateParameter{
		APICreateOrUpdateProperties: properties,
	}

	apiContract, err := client.CreateOrUpdate(ctx, resGroup, serviceName, apiId, apiParams, "")
	if err != nil {
		return err
	}

	d.SetId(*apiContract.ID)

	return flattenApiManagementApiContract(apiContract)
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

	resp, err := client.Delete(ctx, resGroup, serviceName, apiid, "")

	if err != nil {
		if utils.ResponseWasNotFound(resp) {
			return nil
		}

		return err
	}

	return nil
}

func expandApiManagementApiProperties(d *schema.ResourceData) *apimanagement.APICreateOrUpdateProperties {
	revision := d.Get("revision").(string)
	apiType := d.Get("type").(string)
	version := d.Get("version").(string)
	versionSetId := d.Get("version_set_id").(string)

	oAuth := expandApiManagementApiOAuth(d.Get("oauth").([]interface{}))

	return &apimanagement.APICreateOrUpdateProperties{
		APIRevision:     &revision,
		APIType:         apimanagement.APIType(apiType),
		APIVersion:      &version,
		APIVersionSetID: &versionSetId,
		AuthenticationSettings: &apimanagement.AuthenticationSettingsContract{
			OAuth2: oAuth,
		},
	}
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

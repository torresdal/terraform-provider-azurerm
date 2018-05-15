package azurerm

// Based the auto-generated code by Microsoft (R) AutoRest Code Generator.

import (
	"fmt"

	"github.com/Azure/azure-sdk-for-go/services/logic/mgmt/2016-06-01/logic"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
)

func resourceArmLogicApp() *schema.Resource {
	return &schema.Resource{
		Create: resourceArmLogicAppCreateOrUpdate,
		Read:   resourceArmLogicAppRead,
		Update: resourceArmLogicAppCreateOrUpdate,
		Delete: resourceArmLogicAppDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Required: true,
				ForceNew: true,
				Type:     schema.TypeString,
			},
			"location":            locationSchema(),
			"resource_group_name": resourceGroupNameSchema(),
			"tags":                tagsSchema(),

			"integration_account": {
				Optional: true,
				Type:     schema.TypeList,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"keyvault_id": {
							Optional: true,
							Type:     schema.TypeString,
						},
					},
				},
			},
			"parameters": {
				Optional: true,
				Type:     schema.TypeMap,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"description": {
							Optional: true,
							Type:     schema.TypeString,
						},
						"type": {
							Optional: true,
							Type:     schema.TypeString,
						},
					},
				},
			},
			"provisioning_state": {
				Optional: true,
				Type:     schema.TypeString,
			},
			"sku": {
				Optional: true,
				Type:     schema.TypeList,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"keyvault_id": {
							Optional: true,
							Type:     schema.TypeString,
						},
						"name": {
							Required: true,
							Type:     schema.TypeString,
						},
					},
				},
			},
			"state": {
				Optional: true,
				Type:     schema.TypeString,
			},
			"type": {
				Optional: true,
				Type:     schema.TypeString,
			},
			"version": {
				Optional: true,
				Type:     schema.TypeString,
			},

			"access_endpoint": {
				Optional: true,
				Computed: true,
				Type:     schema.TypeString,
			},
		},
	}
}

func resourceArmLogicAppCreateOrUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).logicAppClient
	ctx := meta.(*ArmClient).StopContext

	workflowName := d.Get("name").(string)
	location := azureRMNormalizeLocation(d.Get("location").(string))
	resourceGroupName := d.Get("resource_group_name").(string)

	workflow := logic.Workflow{
		Location: utils.String(location),
	}

	if paramValue, paramExists := d.GetOk("type"); paramExists {
		workflow.Type = utils.String(paramValue.(string))
	}
	if paramValue, paramExists := d.GetOk("tags"); paramExists {
		tmpParamOfTags := expandTags(paramValue.(map[string]interface{}))
		workflow.Tags = tmpParamOfTags
	}
	if paramValue, paramExists := d.GetOk("provisioning_state"); paramExists {
		workflow.ProvisioningState = logic.WorkflowProvisioningState(paramValue.(string))
	}
	if paramValue, paramExists := d.GetOk("state"); paramExists {
		workflow.State = logic.WorkflowState(paramValue.(string))
	}
	if paramValue, paramExists := d.GetOk("version"); paramExists {
		workflow.Version = utils.String(paramValue.(string))
	}
	if paramValue, paramExists := d.GetOk("access_endpoint"); paramExists {
		workflow.AccessEndpoint = utils.String(paramValue.(string))
	}
	workflow.Sku = &logic.Sku{}
	if paramValue, paramExists := d.GetOk("sku"); paramExists {
		tmpParamOfSku := paramValue.(map[string]interface{})
		workflow.Sku.Name = logic.SkuName(tmpParamOfSku["name"].(string))
		workflow.Sku.Plan = &logic.ResourceReference{}
		if paramValue, paramExists := tmpParamOfSku["keyvault_id"]; paramExists {
			workflow.Sku.Plan.ID = utils.String(paramValue.(string))
		}
	}
	workflow.IntegrationAccount = &logic.ResourceReference{}
	if paramValue, paramExists := d.GetOk("integration_account"); paramExists {
		tmpParamOfIntegrationAccount := paramValue.(map[string]interface{})
		if paramValue, paramExists := tmpParamOfIntegrationAccount["keyvault_id"]; paramExists {
			workflow.IntegrationAccount.ID = utils.String(paramValue.(string))
		}
	}
	if paramValue, paramExists := d.GetOk("parameters"); paramExists {
		tmpParamOfParameters := make(map[string]logic.WorkflowParameter)
		for tmpParamKeyOfParameters, tmpParamItemOfParameters := range paramValue.(map[string]interface{}) {
			tmpParamValueOfParameters := tmpParamItemOfParameters.(map[string]interface{})
			workflowParameters := &logic.WorkflowParameter{}
			if paramValue, paramExists := tmpParamValueOfParameters["type"]; paramExists {
				workflowParameters.Type = logic.ParameterType(paramValue.(string))
			}
			if paramValue, paramExists := tmpParamValueOfParameters["description"]; paramExists {
				workflowParameters.Description = utils.String(paramValue.(string))
			}
			tmpParamOfParameters[tmpParamKeyOfParameters] = workflowParameters
		}
		workflow.Parameters = &tmpParamOfParameters
	}

	_, err := client.CreateOrUpdate(ctx, resourceGroupName, workflowName, workflow)
	if err != nil {
		return fmt.Errorf("Logic App creation error: %+v", err)
	}

	read, err := client.Get(ctx, resourceGroupName, workflowName)
	if err != nil {
		return fmt.Errorf("Cannot get Logic App info after created: %+v", err)
	}
	if read.ID == nil {
		return fmt.Errorf("Cannot get the ID of Logic App %q (Resource Group %q) ID", workflowName, resourceGroupName)
	}
	d.SetId(*read.ID)

	return resourceArmLogicAppRead(d, meta)
}

func resourceArmLogicAppRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).logicAppClient
	ctx := meta.(*ArmClient).StopContext

	resourceGroupName := d.Get("resource_group_name").(string)
	workflowName := d.Get("workflow_name").(string)

	response, err := client.Get(ctx, resourceGroupName, workflowName)
	if err != nil {
		return fmt.Errorf("Logic App read error: %+v", err)
	}

	if response.Name != nil {
		d.Set("name", *response.Name)
	}
	if response.Location != nil {
		d.Set("location", *response.Location)
	}
	flattenAndSetTags(d, response.Tags)
	d.Set("provisioning_state", response.ProvisioningState)
	d.Set("state", response.State)
	if response.Version != nil {
		d.Set("version", *response.Version)
	}
	if response.AccessEndpoint != nil {
		d.Set("access_endpoint", *response.AccessEndpoint)
	}
	if response.Sku != nil {
		tmpRespOfSku := make(map[string]interface{})
		d.Set("sku", tmpRespOfSku)
	}
	if response.IntegrationAccount != nil {
		tmpRespOfIntegrationAccount := make(map[string]interface{})
		if response.IntegrationAccount.Name != nil {
			d.Set("name", *response.IntegrationAccount.Name)
		}
		if response.IntegrationAccount.Type != nil {
			d.Set("type", *response.IntegrationAccount.Type)
		}
		d.Set("integration_account", tmpRespOfIntegrationAccount)
	}
	if response.Parameters != nil {
		tmpRespOfParameters := make(map[string]interface{})
		for tmpRespKeyOfParameters, tmpRespItemOfParameters := range *response.Parameters {
			tmpRespValueOfParameters := make(map[string]interface{})
			tmpRespValueOfParameters["type"] = tmpRespItemOfParameters.Type
			if tmpRespItemOfParameters.Description != nil {
				tmpRespValueOfParameters["description"] = *tmpRespItemOfParameters.Description
			}
			tmpRespOfParameters[tmpRespKeyOfParameters] = tmpRespValueOfParameters
		}
		d.Set("parameters", tmpRespOfParameters)
	}

	return nil
}

func resourceArmLogicAppDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).logicAppClient
	ctx := meta.(*ArmClient).StopContext

	resourceGroupName := d.Get("resource_group_name").(string)
	workflowName := d.Get("workflow_name").(string)

	_, err := client.Delete(ctx, resourceGroupName, workflowName)
	if err != nil {
		return fmt.Errorf("Logic App deletion error: %+v", err)
	}

	return nil
}

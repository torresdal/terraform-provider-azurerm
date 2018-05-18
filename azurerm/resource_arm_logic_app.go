package azurerm

// Based the auto-generated code by Microsoft (R) AutoRest Code Generator.

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/Azure/azure-sdk-for-go/services/logic/mgmt/2016-06-01/logic"
	"github.com/hashicorp/terraform/helper/schema"
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
			"definition": {
				Required: true,
				Type:     schema.TypeList,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"schema": {
							Optional: true,
							Type:     schema.TypeString,
							Default:  "https://schema.management.azure.com/providers/Microsoft.Logic/schemas/2016-06-01/workflowdefinition.json",
						},
						"content_version": {
							Optional: true,
							Type:     schema.TypeString,
							Default:  "1.0.0.0",
						},
					},
				},
			},
			"tags": tagsSchema(),
			"access_endpoint": {
				Computed: true,
				Type:     schema.TypeString,
			},
		},
	}
}

func resourceArmLogicAppCreateOrUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).logicAppClient
	ctx := meta.(*ArmClient).StopContext

	workflow := logic.Workflow{
		WorkflowProperties: &logic.WorkflowProperties{
			Definition: map[string]interface{}{
				"$schema":        GetFieldString(d, "definition.0.schema"),
				"contentVersion": GetFieldString(d, "definition.0.content_version"),
				"parameters":     "{}",
				"triggers":       "{}",
				"actions":        "{}",
				"outputs":        "{}",
			},
		},
		Location: GetLocationField(d),
		Tags:     GetTagsField(d),
	}

	data, _ := json.Marshal(workflow)
	log.Printf("REQUEST BODY: " + string(data))

	err := CreateAPICall(d, client, ctx, "CreateOrUpdate", workflow)
	if err != nil {
		return fmt.Errorf("Logic App creation error: %+v", err)
	}

	return resourceArmLogicAppRead(d, meta)
}

func resourceArmLogicAppRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).logicAppClient
	ctx := meta.(*ArmClient).StopContext

	id, err := parseAzureResourceID(d.Id())
	if err != nil {
		return err
	}
	resourceGroupName := id.ResourceGroup
	workflowName := id.Path["workflows"]

	response, err := client.Get(ctx, resourceGroupName, workflowName)
	if err != nil {
		return fmt.Errorf("Logic App read error: %+v", err)
	}

	data, _ := json.Marshal(response)
	log.Printf("RESPONSE BODY: " + string(data))
	log.Printf("DEFINITION: %+v", response.Definition)

	SetNameAndRGField(d, workflowName, resourceGroupName)
	SetLocationField(d, response.Location)
	SetTagsField(d, response.Tags)
	SetFieldObject(d, "definition", response.Definition, func(r map[string]interface{}, v interface{}) {
		m := response.Definition.(map[string]interface{})
		SetSubFieldOptional(r, "schema", m["$schema"])
		SetSubFieldOptional(r, "content_version", m["contentVersion"])
	})
	SetFieldOptional(d, "access_endpoint", response.AccessEndpoint)

	return nil
}

func resourceArmLogicAppDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).logicAppClient
	ctx := meta.(*ArmClient).StopContext

	id, err := parseAzureResourceID(d.Id())
	if err != nil {
		return err
	}
	resourceGroupName := id.ResourceGroup
	workflowName := id.Path["workflows"]

	_, err = client.Delete(ctx, resourceGroupName, workflowName)
	if err != nil {
		return fmt.Errorf("Logic App deletion error: %+v", err)
	}

	return nil
}

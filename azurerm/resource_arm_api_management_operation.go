package azurerm

import (
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceArmApiManagementOperation() *schema.Resource {
	return &schema.Resource{
		Create: resourceArmApiManagementOperationCreateUpdate,
		Read:   resourceArmApiManagementOperationRead,
		Update: resourceArmApiManagementOperationCreateUpdate,
		Delete: resourceArmApiManagementOperationDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"method": {
				Type:     schema.TypeString,
				Required: true,
			},
			"url_template": {
				Type:     schema.TypeString,
				Required: true,
			},
			"template_params": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     apiManagementApiResourceOperationParamContract(),
			},
			"request": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"description": {
							Type:     schema.TypeString,
							Required: true,
						},
						"query_params": {
							Type:     schema.TypeList,
							Optional: true,
							Elem:     apiManagementApiResourceOperationParamContract(),
						},
						"headers": {
							Type:     schema.TypeList,
							Optional: true,
							Elem:     apiManagementApiResourceOperationParamContract(),
						},
						"representations": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"content_type": {
										Type:     schema.TypeString,
										Required: true,
									},
									"sample": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"schema_id": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"type_name": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"form_params": {
										Type:     schema.TypeList,
										Optional: true,
										Elem:     apiManagementApiResourceOperationParamContract(),
									},
								},
							},
						},
					},
				},
			},
			"responses": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"status_code": {
							Type:     schema.TypeString,
							Required: true,
						},
						"description": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"representations": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"content_type": {
										Type:     schema.TypeString,
										Required: true,
									},
									"sample": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"schema_id": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"type_name": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"form_params": {
										Type:     schema.TypeList,
										Optional: true,
										Elem:     apiManagementApiResourceOperationParamContract(),
									},
								},
							},
						},
						"headers": {
							Type:     schema.TypeList,
							Optional: true,
							Elem:     apiManagementApiResourceOperationParamContract(),
						},
					},
				},
			},
		},
	}
}

func apiManagementApiResourceOperationParamContract() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"type": {
				Type:     schema.TypeString,
				Required: true,
			},
			"values": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"default_value": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"required": {
				Type:     schema.TypeBool,
				Optional: true,
			},
		},
	}
}

func resourceArmApiManagementOperationCreateUpdate(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceArmApiManagementOperationRead(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceArmApiManagementOperationDelete(d *schema.ResourceData, meta interface{}) error {
	return nil
}

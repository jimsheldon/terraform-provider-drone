package drone

import (
	"context"
	"fmt"
	"terraform-provider-drone/drone/utils"
	"time"

	"github.com/drone/drone-go/drone"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceTemplate() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"last_updated": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"namespace": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"data": {
				Type:     schema.TypeString,
				Required: true,
			},
		},

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CreateContext: resourceTemplateCreate,
		ReadContext:   resourceTemplateRead,
		UpdateContext: resourceTemplateUpdate,
		DeleteContext: resourceTemplateDelete,
	}
}

func resourceTemplateCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(drone.Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	namespace := d.Get("namespace").(string)

	template, err := client.TemplateCreate(namespace, createTemplate(d))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("%s/%s", namespace, template.Name))

	return diags
}

func resourceTemplateRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(drone.Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	namespace, name, err := utils.ParseOrgId(d.Id(), "template_name")
	if err != nil {
		return diag.FromErr(err)
	}

	template, err := client.Template(namespace, name)
	if err != nil {
		return diag.FromErr(err)
	}

	readTemplate(d, template)

	return diags
}

func resourceTemplateUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(drone.Client)

	namespace, name, err := utils.ParseOrgId(d.Id(), "template_name")
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = client.TemplateUpdate(namespace, name, createTemplate(d))
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("last_updated", time.Now().Format(time.RFC850))

	return resourceTemplateRead(ctx, d, m)
}

func resourceTemplateDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(drone.Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	namespace, name, err := utils.ParseOrgId(d.Id(), "template_name")
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.TemplateDelete(namespace, name)
	if err != nil {
		return diag.FromErr(err)
	}

	// d.SetId("") is automatically called assuming delete returns no errors, but
	// it is added here for explicitness.
	d.SetId("")

	return diags
}

func createTemplate(d *schema.ResourceData) (template *drone.Template) {
	template = &drone.Template{
		Name: d.Get("name").(string),
		Data: d.Get("data").(string),
	}

	return
}

func readTemplate(d *schema.ResourceData, template *drone.Template) {
	d.Set("name", template.Name)
	d.Set("data", template.Data)
}

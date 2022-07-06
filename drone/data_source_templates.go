package drone

import (
	"context"
	"fmt"

	"github.com/drone/drone-go/drone"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceTemplates() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for retrieving all Drone templates in a namespace",
		ReadContext: dataSourceTemplatesRead,
		Schema: map[string]*schema.Schema{
			"namespace": {
				Type:     schema.TypeString,
				Required: true,
			},
			"templates": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"data": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceTemplatesRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(drone.Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	namespace := d.Get("namespace").(string)
	templates, err := client.TemplateList(namespace)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to retrieve templates",
			Detail:   err.Error(),
		})

		return diags
	}

	id := ""
	dataTemplates := make([]interface{}, len(templates), len(templates))

	for i, template := range templates {
		id = id + template.Name
		r := make(map[string]interface{})
		r["name"] = template.Name
		r["data"] = template.Data
		dataTemplates[i] = r
	}

	d.Set("templates", dataTemplates)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to set templates",
			Detail:   err.Error(),
		})

		return diags
	}

	d.SetId(fmt.Sprintf("%s/%s", namespace, id))

	return diags
}

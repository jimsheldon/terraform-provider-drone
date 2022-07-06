package drone

import (
	"context"

	"github.com/drone/drone-go/drone"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceUsers() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for retrieving all Drone users",
		ReadContext: dataSourceUsersRead,
		Schema: map[string]*schema.Schema{
			"users": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"active": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"admin": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"email": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"login": {
							Type:     schema.TypeString,
							Required: true,
						},
						"machine": {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceUsersRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(drone.Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	users, err := client.UserList()
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to retrieve users",
			Detail:   err.Error(),
		})

		return diags
	}

	id := ""
	dataUsers := make([]interface{}, len(users), len(users))

	for i, user := range users {
		id = id + user.Login
		r := make(map[string]interface{})
		r["active"] = user.Active
		r["admin"] = user.Admin
		r["email"] = user.Email
		r["login"] = user.Login
		r["machine"] = user.Machine
		dataUsers[i] = r
	}

	d.Set("users", dataUsers)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to set users",
			Detail:   err.Error(),
		})

		return diags
	}

	d.SetId(id)

	return diags
}

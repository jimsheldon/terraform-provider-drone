package drone

import (
	"context"
	"fmt"
	"time"

	"terraform-provider-drone/drone/utils"

	"github.com/drone/drone-go/drone"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceUser() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"last_updated": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"active": {
				Type:     schema.TypeBool,
				Required: true,
			},
			"admin": {
				Type:     schema.TypeBool,
				Required: true,
			},
			"login": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"machine": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"token": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		CreateContext: resourceUserCreate,
		ReadContext:   resourceUserRead,
		UpdateContext: resourceUserUpdate,
		DeleteContext: resourceUserDelete,
	}
}

func resourceUserCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(drone.Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	user, err := client.UserCreate(createUser(d))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(user.Login)

	return diags
}

func resourceUserRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(drone.Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	user, err := client.User(d.Id())
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Failed to read Drone user with id: %s", d.Id()),
			Detail:   err.Error(),
		})

		return diags
	}

	readUser(d, user)

	return diags
}

func resourceUserUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(drone.Client)

	user, err := client.User(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = client.UserUpdate(user.Login, updateUser(d))
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("last_updated", time.Now().Format(time.RFC850))

	return resourceUserRead(ctx, d, m)
}

func resourceUserDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(drone.Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	login := d.Get("login").(string)

	err := client.UserDelete(login)
	if err != nil {
		return diag.FromErr(err)
	}

	// d.SetId("") is automatically called assuming delete returns no errors, but
	// it is added here for explicitness.
	d.SetId("")

	return diags
}

func createUser(d *schema.ResourceData) (user *drone.User) {
	user = &drone.User{
		Login:   d.Get("login").(string),
		Active:  d.Get("active").(bool),
		Admin:   d.Get("admin").(bool),
		Machine: d.Get("machine").(bool),
	}

	return
}

func updateUser(data *schema.ResourceData) (userPatch *drone.UserPatch) {
	userPatch = &drone.UserPatch{
		Active:  utils.Bool(data.Get("active").(bool)),
		Admin:   utils.Bool(data.Get("admin").(bool)),
		Machine: utils.Bool(data.Get("machine").(bool)),
	}

	return
}

func readUser(d *schema.ResourceData, user *drone.User) {
	d.Set("login", user.Login)
	d.Set("active", user.Active)
	d.Set("machine", user.Machine)
	d.Set("admin", user.Admin)
	d.Set("token", user.Token)
}

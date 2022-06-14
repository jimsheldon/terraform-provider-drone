package drone

import (
	"context"
	"time"

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

	login := d.Get("login").(string)
	user := &drone.User{
		Active:  d.Get("active").(bool),
		Admin:   d.Get("admin").(bool),
		Login:   login,
		Machine: d.Get("machine").(bool),
	}

	user, err := client.UserCreate(user)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(login)

	resourceUserRead(ctx, d, m)

	return diags
}

func resourceUserRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(drone.Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	user, err := client.User(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("login", user.Login); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceUserUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(drone.Client)

	login := d.Get("login").(string)

	_, err := client.UserUpdate(login, updateUser(d))
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

func updateUser(data *schema.ResourceData) (user *drone.UserPatch) {
	userPatch := &drone.UserPatch{
		Active:  Bool(data.Get("active").(bool)),
		Admin:   Bool(data.Get("admin").(bool)),
		Machine: Bool(data.Get("machine").(bool)),
	}
	return userPatch
}

func Bool(val bool) *bool {
	return &val
}

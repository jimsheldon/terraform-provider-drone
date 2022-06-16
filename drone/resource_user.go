package drone

import (
	"context"
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

	user, err := client.User(d.Id())

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

func createUser(data *schema.ResourceData) (user *drone.User) {
	user = &drone.User{
		Login:   data.Get("login").(string),
		Active:  data.Get("active").(bool),
		Admin:   data.Get("admin").(bool),
		Machine: data.Get("machine").(bool),
	}

	return user
}

func updateUser(data *schema.ResourceData) (user *drone.UserPatch) {
	userPatch := &drone.UserPatch{
		Active:  utils.Bool(data.Get("active").(bool)),
		Admin:   utils.Bool(data.Get("admin").(bool)),
		Machine: utils.Bool(data.Get("machine").(bool)),
	}

	return userPatch
}

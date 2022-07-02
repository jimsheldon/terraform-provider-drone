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

func resourceOrgSecret() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"last_updated": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"namespace": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"value": {
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
				ForceNew:  false,
			},
			"allow_on_pull_request": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: false,
			},
			"allow_push_on_pull_request": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: false,
			},
		},

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		CreateContext: resourceOrgSecretCreate,
		ReadContext:   resourceOrgSecretRead,
		UpdateContext: resourceOrgSecretUpdate,
		DeleteContext: resourceOrgSecretDelete,
	}
}

func resourceOrgSecretCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(drone.Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	namespace := d.Get("namespace").(string)

	secret, err := client.OrgSecretCreate(namespace, createOrgSecret(d))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("%s/%s", namespace, secret.Name))

	return diags
}

func resourceOrgSecretRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(drone.Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	namespace, name, err := utils.ParseOrgId(d.Id(), "secret_name")
	if err != nil {
		return diag.FromErr(err)
	}

	secret, err := client.OrgSecret(namespace, name)
	if err != nil {
		return diag.FromErr(err)
	}

	readOrgSecret(d, secret)

	return diags
}

func resourceOrgSecretUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(drone.Client)

	namespace, _, err := utils.ParseOrgId(d.Id(), "secret_name")
	if err != nil {
		return diag.FromErr(err)
	}

	client.OrgSecretUpdate(namespace, createOrgSecret(d))
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("last_updated", time.Now().Format(time.RFC850))

	return resourceOrgSecretRead(ctx, d, m)
}

func resourceOrgSecretDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(drone.Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	namespace, name, err := utils.ParseOrgId(d.Id(), "secret_name")

	err = client.OrgSecretDelete(namespace, name)
	if err != nil {
		return diag.FromErr(err)
	}

	// d.SetId("") is automatically called assuming delete returns no errors, but
	// it is added here for explicitness.
	d.SetId("")

	return diags
}

func createOrgSecret(data *schema.ResourceData) (secret *drone.Secret) {
	return &drone.Secret{
		Name:            data.Get("name").(string),
		Data:            data.Get("value").(string),
		PullRequest:     data.Get("allow_on_pull_request").(bool),
		PullRequestPush: data.Get("allow_push_on_pull_request").(bool),
	}
}

func readOrgSecret(data *schema.ResourceData, secret *drone.Secret) {
	data.Set("name", secret.Name)
	data.Set("allow_on_pull_request", secret.PullRequest)
	data.Set("allow_push_on_pull_request", secret.PullRequestPush)
}

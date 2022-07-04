package drone

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"terraform-provider-drone/drone/utils"

	"github.com/drone/drone-go/drone"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceSecret() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"last_updated": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"repository": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringMatch(
					regexp.MustCompile("^[^/ ]+/[^/ ]+$"),
					"Invalid repository (e.g. octocat/hello-world)",
				),
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
			},
			"allow_on_pull_request": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
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

		CreateContext: resourceSecretCreate,
		ReadContext:   resourceSecretRead,
		UpdateContext: resourceSecretUpdate,
		DeleteContext: resourceSecretDelete,
	}
}

func resourceSecretCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(drone.Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	owner, repo, err := utils.ParseRepo(d.Get("repository").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	secret, err := client.SecretCreate(owner, repo, createSecret(d))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("%s/%s/%s", owner, repo, secret.Name))

	return diags
}

func resourceSecretRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(drone.Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	owner, repo, name, err := utils.ParseId(d.Id(), "secret_name")
	if err != nil {
		return diag.FromErr(err)
	}

	secret, err := client.Secret(owner, repo, name)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Failed to read Drone Secret: %s/%s/%s", owner, repo, name),
			Detail:   err.Error(),
		})

		return diags
	}

	readSecret(d, owner, repo, secret)

	return diags
}

func resourceSecretUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(drone.Client)

	owner, repo, err := utils.ParseRepo(d.Get("repository").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = client.SecretUpdate(owner, repo, createSecret(d))
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("last_updated", time.Now().Format(time.RFC850))

	return resourceSecretRead(ctx, d, m)
}

func resourceSecretDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(drone.Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	owner, repo, name, err := utils.ParseId(d.Id(), "secret_name")
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.SecretDelete(owner, repo, name)
	if err != nil {
		return diag.FromErr(err)
	}

	// d.SetId("") is automatically called assuming delete returns no errors, but
	// it is added here for explicitness.
	d.SetId("")

	return diags
}

func createSecret(d *schema.ResourceData) (secret *drone.Secret) {
	secret = &drone.Secret{
		Name:            d.Get("name").(string),
		Data:            d.Get("value").(string),
		PullRequest:     d.Get("allow_on_pull_request").(bool),
		PullRequestPush: d.Get("allow_push_on_pull_request").(bool),
	}

	return
}

func readSecret(d *schema.ResourceData, owner, repo string, secret *drone.Secret) {
	d.Set("repository", fmt.Sprintf("%s/%s", owner, repo))
	d.Set("name", secret.Name)
	d.Set("allow_on_pull_request", secret.PullRequest)
	d.Set("allow_push_on_pull_request", secret.PullRequestPush)
}

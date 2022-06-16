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

func resourceRepo() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"repository": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringMatch(
					regexp.MustCompile("^[^/ ]+/[^/ ]+$"),
					"Invalid repository (e.g. octocat/hello-world)",
				),
			},
			"configuration": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  ".drone.yml",
			},
			"timeout": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  60,
			},
			"trusted": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"protected": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"visibility": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "private",
			},
		},

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		CreateContext: resourceRepoCreate,
		ReadContext:   resourceRepoRead,
		UpdateContext: resourceRepoUpdate,
		DeleteContext: resourceRepoDelete,
	}
}

func resourceRepoCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(drone.Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	// Refresh repository list
	if _, err := client.RepoListSync(); err != nil {
		return diag.FromErr(err)
	}

	owner, repo, err := utils.ParseRepo(d.Get("repository").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	repository, err := client.Repo(owner, repo)
	if err != nil {
		return diag.FromErr(err)
	}

	resp, err := client.RepoUpdate(owner, repo, createRepo(d))
	if err != nil {
		return diag.FromErr(err)
	}

	if !resp.Active {
		_, err = client.RepoEnable(owner, repo)

		if err != nil {
			return diag.FromErr(err)
		}
	}

	d.SetId(fmt.Sprintf("%s/%s", repository.Namespace, repository.Name))

	resourceRepoRead(ctx, d, m)

	return diags
}

func resourceRepoRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(drone.Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	owner, repo, err := utils.ParseRepo(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	repository, err := client.Repo(owner, repo)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("configuration", repository.Config); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("protected", repository.Protected); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("repository", fmt.Sprintf("%s/%s", repository.Namespace, repository.Name)); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("timeout", repository.Timeout); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("trusted", repository.Trusted); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("visibility", repository.Visibility); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceRepoUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(drone.Client)

	owner, repo, err := utils.ParseRepo(d.Id())

	_, err = client.RepoUpdate(owner, repo, createRepo(d))
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("last_updated", time.Now().Format(time.RFC850))

	return resourceRepoRead(ctx, d, m)
}

func resourceRepoDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(drone.Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	owner, repo, err := utils.ParseRepo(d.Id())

	err = client.RepoDisable(owner, repo)
	if err != nil {
		return diag.FromErr(err)
	}

	// d.SetId("") is automatically called assuming delete returns no errors, but
	// it is added here for explicitness.
	d.SetId("")

	return diags
}

func createRepo(data *schema.ResourceData) (repository *drone.RepoPatch) {
	config := data.Get("configuration").(string)
	protected := data.Get("protected").(bool)
	timeout := int64(data.Get("timeout").(int))
	trusted := data.Get("trusted").(bool)
	visibility := data.Get("visibility").(string)

	repository = &drone.RepoPatch{
		Config:     &config,
		Protected:  &protected,
		Trusted:    &trusted,
		Timeout:    &timeout,
		Visibility: &visibility,
	}

	return nil
}

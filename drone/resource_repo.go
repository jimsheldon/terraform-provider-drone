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
			"cancel_pulls": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"cancel_push": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"cancel_running": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"configuration": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  ".drone.yml",
			},
			"last_updated": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"ignore_forks": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"ignore_pulls": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"protected": {
				Type:     schema.TypeBool,
				Optional: true,
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
			"timeout": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  60,
			},
			"trusted": {
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

	if err := d.Set("cancel_pulls", repository.CancelPulls); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("cancel_push", repository.CancelPush); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("cancel_running", repository.CancelPush); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("configuration", repository.Config); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("ignore_forks", repository.IgnoreForks); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("ignore_pulls", repository.IgnorePulls); err != nil {
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

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	owner, repo, err := utils.ParseRepo(d.Id())

	cancel_pulls := d.Get("cancel_pulls").(bool)
	cancel_push := d.Get("cancel_push").(bool)
	cancel_running := d.Get("cancel_running").(bool)
	config := d.Get("configuration").(string)
	ignore_forks := d.Get("ignore_forks").(bool)
	ignore_pulls := d.Get("ignore_pulls").(bool)
	protected := d.Get("protected").(bool)
	timeout := int64(d.Get("timeout").(int))
	trusted := d.Get("trusted").(bool)
	visibility := d.Get("visibility").(string)

	repository := &drone.RepoPatch{
		CancelPulls:   &cancel_pulls,
		CancelPush:    &cancel_push,
		CancelRunning: &cancel_running,
		Config:        &config,
		IgnoreForks:   &ignore_forks,
		IgnorePulls:   &ignore_pulls,
		Protected:     &protected,
		Trusted:       &trusted,
		Timeout:       &timeout,
		Visibility:    &visibility,
	}

	_, err = client.RepoUpdate(owner, repo, repository)
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("last_updated", time.Now().Format(time.RFC850))

	return diags
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
	cancel_pulls := data.Get("cancel_pulls").(bool)
	cancel_push := data.Get("cancel_push").(bool)
	cancel_running := data.Get("cancel_running").(bool)
	config := data.Get("configuration").(string)
	ignore_forks := data.Get("ignore_forks").(bool)
	ignore_pulls := data.Get("ignore_pulls").(bool)
	protected := data.Get("protected").(bool)
	timeout := int64(data.Get("timeout").(int))
	trusted := data.Get("trusted").(bool)
	visibility := data.Get("visibility").(string)

	repository = &drone.RepoPatch{
		CancelPulls:   &cancel_pulls,
		CancelPush:    &cancel_push,
		CancelRunning: &cancel_running,
		Config:        &config,
		IgnoreForks:   &ignore_forks,
		IgnorePulls:   &ignore_pulls,
		Protected:     &protected,
		Trusted:       &trusted,
		Timeout:       &timeout,
		Visibility:    &visibility,
	}

	return nil
}

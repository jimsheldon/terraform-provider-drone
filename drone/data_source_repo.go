package drone

import (
	"context"
	"fmt"
	"regexp"
	"terraform-provider-drone/drone/utils"

	"github.com/drone/drone-go/drone"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourceRepo() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for retrieving a Drone repository",
		ReadContext: dataSourceRepoRead,
		Schema: map[string]*schema.Schema{
			"cancel_pulls": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"cancel_push": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"cancel_running": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"configuration": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ignore_forks": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"ignore_pulls": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"protected": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"repository": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringMatch(
					regexp.MustCompile("^[^/ ]+/[^/ ]+$"),
					"Invalid repository (e.g. octocat/hello-world)",
				),
			},
			"timeout": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"trusted": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"visibility": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceRepoRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(drone.Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	// Refresh repository list
	if _, err := client.RepoListSync(); err != nil {
		return diag.FromErr(err)
	}

	repository := d.Get("repository").(string)
	owner, name, err := utils.ParseRepo(repository)
	if err != nil {
		return diag.FromErr(err)
	}

	repo, err := client.Repo(owner, name)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Failed to read repo %s", repository),
			Detail:   err.Error(),
		})

		return diags
	}

	d.Set("cancel_pulls", repo.CancelPulls)
	d.Set("cancel_push", repo.CancelPush)
	d.Set("cancel_running", repo.CancelRunning)
	d.Set("configuration", repo.Config)
	d.Set("ignore_forks", repo.IgnoreForks)
	d.Set("ignore_pulls", repo.IgnorePulls)
	d.Set("protected", repo.Protected)
	d.Set("timeout", repo.Timeout)
	d.Set("trusted", repo.Trusted)
	d.Set("visibility", repo.Visibility)

	d.SetId(repository)

	return diags
}

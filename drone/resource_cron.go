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

func resourceCron() *schema.Resource {
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
			"disabled": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"event": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					drone.EventPush,
				}, false),
			},
			"branch": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "master",
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"target": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"expr": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "@monthly",
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					"@hourly",
					"@daily",
					"@weekly",
					"@monthly",
					"@yearly",
				}, false),
			},
		},

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		CreateContext: resourceCronCreate,
		ReadContext:   resourceCronRead,
		UpdateContext: resourceCronUpdate,
		DeleteContext: resourceCronDelete,
	}
}

func resourceCronCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(drone.Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	owner, repo, err := utils.ParseRepo(d.Get("repository").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	cron, err := client.CronCreate(owner, repo, createCron(d))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("%s/%s/%s", owner, repo, cron.Name))

	return diags
}

func resourceCronRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(drone.Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	owner, repo, name, err := utils.ParseId(d.Id(), "cron_name")
	if err != nil {
		return diag.FromErr(err)
	}

	cron, err := client.Cron(owner, repo, name)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Failed to read Drone Cron: %s/%s/%s not found", owner, repo, name),
			Detail:   err.Error(),
		})

		return diags
	}

	readCron(d, owner, repo, cron)

	return diags
}

func resourceCronUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(drone.Client)

	owner, repo, name, err := utils.ParseId(d.Id(), "cron_name")
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = client.CronUpdate(owner, repo, name, updateCron(d))
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("last_updated", time.Now().Format(time.RFC850))

	return resourceCronRead(ctx, d, m)
}

func resourceCronDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(drone.Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	owner, repo, name, err := utils.ParseId(d.Id(), "cron_name")
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.CronDelete(owner, repo, name)
	if err != nil {
		return diag.FromErr(err)
	}

	// d.SetId("") is automatically called assuming delete returns no errors, but
	// it is added here for explicitness.
	d.SetId("")

	return diags
}

func createCron(d *schema.ResourceData) (repository *drone.Cron) {
	return &drone.Cron{
		Disabled: d.Get("disabled").(bool),
		Branch:   d.Get("branch").(string),
		Expr:     d.Get("expr").(string),
		Event:    d.Get("event").(string),
		Name:     d.Get("name").(string),
		Target:   d.Get("target").(string),
	}
}

func updateCron(d *schema.ResourceData) (repository *drone.CronPatch) {
	branch := d.Get("branch").(string)
	disabled := d.Get("disabled").(bool)
	event := d.Get("event").(string)
	target := d.Get("target").(string)

	cron := &drone.CronPatch{
		Disabled: utils.Bool(disabled),
		Branch:   &branch,
		Event:    &event,
		Target:   &target,
	}
	return cron
}

func readCron(d *schema.ResourceData, namespace string, repo string, cron *drone.Cron) {
	d.Set("repository", fmt.Sprintf("%s/%s", namespace, repo))
	d.Set("branch", cron.Branch)
	d.Set("disabled", cron.Disabled)
	d.Set("expr", cron.Expr)
	d.Set("name", cron.Name)
	d.Set("target", cron.Target)
}

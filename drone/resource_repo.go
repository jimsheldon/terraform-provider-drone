package drone

import (
	"fmt"
	"regexp"

	"terraform-provider-drone/drone/utils"

	"github.com/drone/drone-go/drone"
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
			"trusted": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"protected": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"timeout": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  60,
			},
			"visibility": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "private",
			},
			"configuration": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  ".drone.yml",
			},
		},

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		//		Create: resourceRepoCreate,
		Read: resourceRepoRead,
		//		Update: resourceRepoUpdate,
		//		Delete: resourceRepoDelete,
		//		Exists: resourceRepoExists,
	}
}

func resourceRepoRead(data *schema.ResourceData, meta interface{}) error {
	client := meta.(drone.Client)

	owner, repo, err := utils.ParseRepo(data.Id())

	if err != nil {
		return err
	}

	repository, err := client.Repo(owner, repo)
	if err != nil {
		return fmt.Errorf("failed to read Drone Repo: %s/%s", owner, repo)
	}

	return readRepo(data, repository, err)
}

func readRepo(data *schema.ResourceData, repository *drone.Repo, err error) error {
	if err != nil {
		return err
	}

	data.SetId(fmt.Sprintf("%s/%s", repository.Namespace, repository.Name))

	data.Set("repository", fmt.Sprintf("%s/%s", repository.Namespace, repository.Name))
	data.Set("trusted", repository.Trusted)
	data.Set("protected", repository.Protected)
	data.Set("timeout", repository.Timeout)
	data.Set("visibility", repository.Visibility)
	data.Set("configuration", repository.Config)

	return nil
}

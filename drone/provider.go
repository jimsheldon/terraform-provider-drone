package drone

import (
	"context"
	"crypto/tls"
	"net/http"

	"github.com/drone/drone-go/drone"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jackspirou/syscerts"
	"golang.org/x/oauth2"
)

// Provider -
func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"server": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "URL for the drone server",
				DefaultFunc: schema.EnvDefaultFunc("DRONE_SERVER", nil),
			},
			"token": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				Description: "API Token for the drone server",
				DefaultFunc: schema.EnvDefaultFunc("DRONE_TOKEN", nil),
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"drone_cron":      resourceCron(),
			"drone_orgsecret": resourceOrgSecret(),
			"drone_repo":      resourceRepo(),
			"drone_secret":    resourceSecret(),
			"drone_template":  resourceTemplate(),
			"drone_user":      resourceUser(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"drone_template":  dataSourceTemplate(),
			"drone_templates": dataSourceTemplates(),
			"drone_user":      dataSourceUser(),
			"drone_users":     dataSourceUsers(),
			"drone_user_self": dataSourceUserSelf(),
		},
		ConfigureContextFunc: providerConfigure,
	}
}

func providerConfigure(ctx context.Context, data *schema.ResourceData) (interface{}, diag.Diagnostics) {
	config := new(oauth2.Config)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	certs := syscerts.SystemRootsPool()
	tlsConfig := &tls.Config{
		RootCAs:            certs,
		InsecureSkipVerify: false,
	}

	auther := config.Client(
		oauth2.NoContext,
		&oauth2.Token{AccessToken: data.Get("token").(string)},
	)

	trans, _ := auther.Transport.(*oauth2.Transport)
	trans.Base = &http.Transport{
		TLSClientConfig: tlsConfig,
		Proxy:           http.ProxyFromEnvironment,
	}

	server := data.Get("server").(string)
	client := drone.NewClient(server, auther)
	_, err := client.Self()
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create Drone client",
			Detail:   err.Error(),
		})
	}

	return client, diags
}

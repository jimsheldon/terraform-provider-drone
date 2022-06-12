package drone

import (
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/drone/drone-go/drone"
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
				Description: "API Token for the drone server",
				DefaultFunc: schema.EnvDefaultFunc("DRONE_TOKEN", nil),
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"drone_repo": resourceRepo(),
		},
		ConfigureFunc: providerConfigureFunc,
	}
}

func providerConfigureFunc(data *schema.ResourceData) (interface{}, error) {
	config := new(oauth2.Config)

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

	client := drone.NewClient(data.Get("server").(string), auther)

	if _, err := client.Self(); err != nil {
		return nil, fmt.Errorf("drone client failed: %s", err)
	}

	return client, nil
}

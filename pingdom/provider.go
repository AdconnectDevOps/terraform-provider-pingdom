package pingdom

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Provider returns the Pingdom Terraform provider.
func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"api_token": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				Description: "Pingdom API token. Falls back to PINGDOM_API_TOKEN environment variable when not set.",
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"pingdom_check":   resourcePingdomCheck(),
			"pingdom_team":    resourcePingdomTeam(),
			"pingdom_contact": resourcePingdomContact(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"pingdom_contact": dataSourcePingdomContact(),
			"pingdom_team":    dataSourcePingdomTeam(),
		},
		ConfigureContextFunc: providerConfigure,
	}
}

func providerConfigure(_ context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	config := Config{
		APIToken: d.Get("api_token").(string),
	}

	log.Println("[INFO] Initializing Pingdom client")
	client, err := config.Client()
	if err != nil {
		return nil, diag.FromErr(err)
	}
	return client, nil
}

package pingdom

import (
	"fmt"
	"log"
	"os"

	"github.com/russellcardullo/go-pingdom/pingdom"
)

// Config respresents the client configuration
type Config struct {
	APIToken string `mapstructure:"api_token"`
}

// Client returns a new client for accessing pingdom. Explicit provider config
// takes precedence over the PINGDOM_API_TOKEN environment variable; the env
// var is a fallback for the unset case.
func (c *Config) Client() (*pingdom.Client, error) {
	if c.APIToken == "" {
		c.APIToken = os.Getenv("PINGDOM_API_TOKEN")
	}
	if c.APIToken == "" {
		return nil, fmt.Errorf("pingdom: api_token is required — set it in provider config or via the PINGDOM_API_TOKEN environment variable")
	}

	client, err := pingdom.NewClientWithConfig(pingdom.ClientConfig{APIToken: c.APIToken})
	if err != nil {
		return nil, fmt.Errorf("pingdom: failed to create client: %w", err)
	}

	log.Printf("[INFO] Pingdom Client configured.")

	return client, nil
}

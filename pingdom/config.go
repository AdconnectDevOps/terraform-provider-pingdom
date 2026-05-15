package pingdom

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/russellcardullo/go-pingdom/pingdom"
)

// Config respresents the client configuration
type Config struct {
	APIToken string `mapstructure:"api_token"`
}

// Client returns a new client for accessing pingdom. Explicit provider config
// takes precedence over the PINGDOM_API_TOKEN environment variable; the env
// var is a fallback for the unset case. Transport is rate-limited (1 req/sec)
// and retries HTTP 429 responses with exponential backoff.
func (c *Config) Client() (*pingdom.Client, error) {
	if c.APIToken == "" {
		c.APIToken = os.Getenv("PINGDOM_API_TOKEN")
	}
	if c.APIToken == "" {
		return nil, fmt.Errorf("pingdom: api_token is required — set it in provider config or via the PINGDOM_API_TOKEN environment variable")
	}

	httpClient := &http.Client{
		Transport: newRateLimitedTransport(time.Second, 3),
		Timeout:   60 * time.Second,
	}

	client, err := pingdom.NewClientWithConfig(pingdom.ClientConfig{
		APIToken:   c.APIToken,
		HTTPClient: httpClient,
	})
	if err != nil {
		return nil, fmt.Errorf("pingdom: failed to create client: %w", err)
	}

	log.Printf("[INFO] Pingdom Client configured.")

	return client, nil
}

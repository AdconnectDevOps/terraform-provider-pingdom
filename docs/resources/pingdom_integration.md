---
page_title: "pingdom_integration"
description: "Manages a Pingdom webhook integration for custom alert handling"
---

# pingdom_integration Resource

The `pingdom_integration` resource allows you to create, update, and delete Pingdom webhook integrations. Integrations enable you to send alerts to external systems like Slack, PagerDuty, or custom webhooks when monitoring checks fail.

## Example Usage

### Basic Webhook Integration

```hcl
resource "pingdom_integration" "slack_webhook" {
  name = "Slack Webhook"
  url  = "https://hooks.slack.com/services/YOUR/SLACK/WEBHOOK"
}
```

### Integration with Custom Headers

```hcl
resource "pingdom_integration" "custom_webhook" {
  name = "Custom Webhook"
  url  = "https://api.example.com/webhook"
  
  # Custom headers for authentication
  headers = {
    "Authorization" = "Bearer ${var.webhook_token}"
    "User-Agent"    = "Pingdom-Integration"
    "X-Custom"      = "value"
  }
}
```

### Multiple Integrations

```hcl
# Slack integration
resource "pingdom_integration" "slack" {
  name = "Slack Notifications"
  url  = var.slack_webhook_url
}

# PagerDuty integration
resource "pingdom_integration" "pagerduty" {
  name = "PagerDuty Alerts"
  url  = var.pagerduty_webhook_url
}

# Custom monitoring system
resource "pingdom_integration" "monitoring_system" {
  name = "Monitoring System"
  url  = var.monitoring_webhook_url
  
  headers = {
    "API-Key" = var.monitoring_api_key
  }
}
```

## Argument Reference

### Required Arguments

*   `name` (String) - The name of the integration. Must be unique within your account.
*   `url` (String) - The webhook URL to send alerts to.

### Optional Arguments

*   `headers` (Map) - Custom HTTP headers to include with the webhook request.

## Attributes Reference

*   `id` (String) - The ID of the Pingdom integration.
*   `name` (String) - The name of the integration.
*   `url` (String) - The webhook URL.

## Integration Types

### Slack Integration

Send alerts to Slack channels:

```hcl
resource "pingdom_integration" "slack_main" {
  name = "Slack Main Channel"
  url  = "https://hooks.slack.com/services/YOUR/SLACK/WEBHOOK"
}

# Use in monitoring checks
resource "pingdom_check" "website" {
  type       = "http"
  name       = "Website"
  host       = "example.com"
  resolution = 5
  
  integrationids = [
    pingdom_integration.slack_main.id
  ]
}
```

### PagerDuty Integration

Integrate with PagerDuty for incident management:

```hcl
resource "pingdom_integration" "pagerduty" {
  name = "PagerDuty Integration"
  url  = "https://events.pagerduty.com/v2/enqueue"
  
  headers = {
    "Content-Type" = "application/json"
  }
}

# Use in critical services
resource "pingdom_check" "critical_service" {
  type       = "http"
  name       = "Critical Service"
  host       = "critical.example.com"
  resolution = 1
  
  integrationids = [
    pingdom_integration.pagerduty.id
  ]
}
```

### Custom Webhook Integration

Send alerts to your own systems:

```hcl
resource "pingdom_integration" "custom_monitoring" {
  name = "Custom Monitoring System"
  url  = "https://monitoring.example.com/webhook"
  
  headers = {
    "Authorization" = "Bearer ${var.monitoring_token}"
    "X-Source"      = "Pingdom"
    "Content-Type"  = "application/json"
  }
}
```

### Microsoft Teams Integration

Send alerts to Microsoft Teams:

```hcl
resource "pingdom_integration" "teams" {
  name = "Microsoft Teams"
  url  = "https://outlook.office.com/webhook/YOUR/TEAMS/WEBHOOK"
}
```

## Webhook Payload

Pingdom sends a JSON payload to your webhook URL when checks fail or recover. The payload includes:

```json
{
  "check_id": 12345,
  "check_name": "Website Check",
  "check_type": "http",
  "host": "example.com",
  "status": "down",
  "status_description": "Check failed",
  "time": 1640995200,
  "duration": 5000,
  "response_time": 5000,
  "error_details": "Connection timeout"
}
```

### Handling Different Statuses

Configure your webhook to handle different alert statuses:

- **`down`**: Service is down
- **`up`**: Service has recovered
- **`paused`**: Check is paused

## Complete Integration Example

Here's a comprehensive example of webhook integrations:

```hcl
# Provider configuration
provider "pingdom" {
  api_token = var.pingdom_api_token
}

# Variables
variable "pingdom_api_token" {
  description = "Pingdom API token"
  type        = string
  sensitive   = true
}

variable "slack_webhook_url" {
  description = "Slack webhook URL"
  type        = string
  sensitive   = true
}

variable "pagerduty_webhook_url" {
  description = "PagerDuty webhook URL"
  type        = string
  sensitive   = true
}

variable "monitoring_api_key" {
  description = "Custom monitoring API key"
  type        = string
  sensitive   = true
}

# Slack integration for general notifications
resource "pingdom_integration" "slack_general" {
  name = "Slack General"
  url  = var.slack_webhook_url
}

# Slack integration for critical alerts
resource "pingdom_integration" "slack_critical" {
  name = "Slack Critical"
  url  = var.slack_webhook_url
}

# PagerDuty integration for incident management
resource "pingdom_integration" "pagerduty" {
  name = "PagerDuty"
  url  = var.pagerduty_webhook_url
  
  headers = {
    "Content-Type" = "application/json"
  }
}

# Custom monitoring system integration
resource "pingdom_integration" "custom_monitoring" {
  name = "Custom Monitoring"
  url  = "https://monitoring.example.com/webhook"
  
  headers = {
    "Authorization" = "Bearer ${var.monitoring_api_key}"
    "X-Source"      = "Pingdom"
    "Content-Type"  = "application/json"
  }
}

# Use integrations in monitoring checks
resource "pingdom_check" "website" {
  type       = "http"
  name       = "Website"
  host       = "example.com"
  resolution = 5
  
  # General notifications
  integrationids = [
    pingdom_integration.slack_general.id
  ]
}

resource "pingdom_check" "critical_service" {
  type       = "http"
  name       = "Critical Service"
  host       = "critical.example.com"
  resolution = 1
  
  # Critical notifications
  integrationids = [
    pingdom_integration.slack_critical.id,
    pingdom_integration.pagerduty.id,
    pingdom_integration.custom_monitoring.id
  ]
}

resource "pingdom_check" "api_endpoint" {
  type       = "http"
  name       = "API Endpoint"
  host       = "api.example.com"
  resolution = 5
  
  # API monitoring
  integrationids = [
    pingdom_integration.slack_general.id,
    pingdom_integration.custom_monitoring.id
  ]
}
```

## Best Practices

### 1. Integration Naming

- Use descriptive names that indicate purpose
- Include the target system in the name
- Use consistent naming patterns

### 2. Security

- Store webhook URLs in Terraform variables
- Use environment variables for local development
- Mark sensitive variables with `sensitive = true`
- Rotate webhook tokens regularly

### 3. Error Handling

- Configure webhook endpoints to handle failures gracefully
- Implement retry logic in your webhook handlers
- Monitor webhook delivery success rates
- Set up fallback notification methods

### 4. Rate Limiting

- Be aware of webhook rate limits
- Implement appropriate throttling
- Monitor webhook performance
- Use appropriate check resolutions

### 5. Testing

- Test webhook endpoints before production use
- Verify payload format and content
- Test both failure and recovery scenarios
- Monitor webhook response times

## Troubleshooting

### Common Integration Issues

**Webhook Not Receiving Alerts**
- Verify the webhook URL is accessible
- Check for authentication requirements
- Ensure the endpoint can handle POST requests
- Verify webhook is enabled in Pingdom

**Authentication Failures**
- Check API keys and tokens
- Verify header format and values
- Ensure credentials are current
- Test authentication manually

**Payload Format Issues**
- Verify your endpoint accepts JSON
- Check content-type headers
- Ensure proper error handling
- Test with sample payloads

**Rate Limiting**
- Monitor webhook call frequency
- Implement appropriate throttling
- Use appropriate check resolutions
- Consider webhook service limits

### Integration Validation

```hcl
# Test integration configuration
resource "pingdom_integration" "test_webhook" {
  name = "Test Webhook"
  url  = "https://webhook.site/your-test-url"
}

# Use in a test check
resource "pingdom_check" "test_check" {
  type       = "http"
  name       = "Test Check"
  host       = "test.example.com"
  resolution = 15  # Use long resolution for testing
  
  integrationids = [
    pingdom_integration.test_webhook.id
  ]
}
```

### Webhook Testing

Test your webhook endpoints with tools like:

- **webhook.site**: For testing webhook delivery
- **ngrok**: For local webhook testing
- **Postman**: For manual webhook testing
- **cURL**: For command-line testing

## Related Resources

- [pingdom_check](../resources/pingdom_check.md) - Create monitoring checks
- [pingdom_team](../resources/pingdom_team.md) - Create teams for notifications
- [pingdom_contact](../resources/pingdom_contact.md) - Configure notification contacts
- [HTTP Monitoring Guide](../guides/http_monitoring.md) - Advanced HTTP monitoring
- [Team Management Guide](../guides/team_management.md) - Notification management

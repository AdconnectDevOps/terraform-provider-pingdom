---
page_title: "pingdom_check"
description: "Manages a Pingdom monitoring check"
---

# pingdom_check Resource

The `pingdom_check` resource allows you to create, update, and delete Pingdom monitoring checks. This resource supports HTTP, HTTPS, ping, and TCP monitoring with configurable alerts, notifications, and performance thresholds.

## Example Usage

### Basic HTTP Check

```hcl
resource "pingdom_check" "website" {
  type       = "http"
  name       = "Website Homepage"
  host       = "example.com"
  resolution = 5
  url        = "/"
}
```

### HTTPS Check with Authentication

```hcl
resource "pingdom_check" "secure_api" {
  type       = "http"
  name       = "Secure API"
  host       = "api.example.com"
  resolution = 1
  url        = "/health"
  encryption = true
  port       = 443
  
  username = "monitor"
  password = var.api_password
  
  requestheaders = {
    "Authorization" = "Bearer ${var.api_token}"
    "User-Agent"   = "Pingdom-Monitor"
  }
}
```

### Ping Check

```hcl
resource "pingdom_check" "network" {
  type       = "ping"
  name       = "Network Availability"
  host       = "192.168.1.1"
  resolution = 1
  
  userids = [
    pingdom_contact.admin.id
  ]
}
```

### TCP Port Check

```hcl
resource "pingdom_check" "database" {
  type            = "tcp"
  name            = "Database Connection"
  host            = "db.example.com"
  port            = 5432
  resolution      = 5
  stringtosend    = "PING"
  stringtoexpect  = "PONG"
}
```

### Advanced HTTP Check with Alerts

```hcl
resource "pingdom_check" "critical_service" {
  type       = "http"
  name       = "Critical Service"
  host       = "critical.example.com"
  resolution = 1
  url        = "/status"
  encryption = true
  
  # Content validation
  shouldcontain    = "healthy"
  shouldnotcontain = "error"
  
  # Performance monitoring
  responsetime_threshold = 3000
  
  # Alert configuration
  sendnotificationwhendown = 2
  notifyagainevery         = 5
  notifywhenbackup         = true
  
  # Notification targets
  userids = [
    pingdom_contact.admin.id,
    pingdom_contact.ops.id
  ]
  
  teamids = [
    pingdom_team.oncall.id
  ]
  
  integrationids = [
    12345678  # Webhook integration ID
  ]
  
  # Custom headers
  requestheaders = {
    "Accept"       = "application/json"
    "X-Monitoring" = "true"
  }
  
  # Organization
  tags = "production,critical,monitoring"
  
  # Regional monitoring
  probefilters = "region:NA,region:EU"
}
```

## Argument Reference

### Required Arguments

*   `name` (String) - The name of the check. Must be unique within your account.
*   `host` (String) - The hostname or IP address to monitor.
*   `type` (String) - The type of check. Valid values: `http`, `https`, `ping`, `tcp`.
*   `resolution` (Number) - The time in minutes between each check. Valid values: `1`, `5`, `15`, `30`, `60`.

### Optional Arguments

*   `url` (String) - Target path on server (default: `/`).
*   `encryption` (Boolean) - Enable HTTPS encryption (default: `false`).
*   `port` (Number) - Target port for HTTP/HTTPS checks (default: 80/443).
*   `username` (String) - Username for HTTP authentication.
*   `password` (String) - Password for HTTP authentication.
*   `shouldcontain` (String) - String that must be present in response.
*   `shouldnotcontain` (String) - String that must NOT be present in response.
*   `postdata` (String) - POST data for form submissions.
*   `requestheaders` (Map) - Custom HTTP headers.
*   `tags` (String) - Comma-separated list of tags.
*   `probefilters` (String) - Region filter (e.g., `"region:NA"`).
*   `stringtosend` (String) - String to send to TCP port.
*   `stringtoexpect` (String) - Expected response string from TCP port.
*   `paused` (Boolean) - Whether the check is active (default: `false`).
*   `responsetime_threshold` (Number) - Response time threshold in milliseconds (default: 30000).
*   `sendnotificationwhendown` (Number) - Consecutive failures before alert (default: 1).
*   `notifyagainevery` (Number) - Re-notify interval in minutes (default: 0).
*   `notifywhenbackup` (Boolean) - Notify when service recovers (default: `false`).
*   `integrationids` (List) - List of webhook integration IDs.
*   `userids` (List) - List of user IDs for notifications.
*   `teamids` (List) - List of team IDs for notifications.

## Attributes Reference

*   `id` (String) - The ID of the Pingdom check.
*   `name` (String) - The name of the check.
*   `host` (String) - The hostname being monitored.
*   `type` (String) - The type of check.
*   `resolution` (Number) - The check resolution in minutes.
*   `url` (String) - The URL path being monitored.
*   `encryption` (Boolean) - Whether HTTPS is enabled.
*   `port` (Number) - The port being monitored.
*   `username` (String) - The username used for authentication.
*   `shouldcontain` (String) - The required content string.
*   `shouldnotcontain` (String) - The forbidden content string.
*   `postdata` (String) - The POST data being sent.
*   `requestheaders` (Map) - The custom HTTP headers.
*   `tags` (String) - The tags associated with the check.
*   `probefilters` (String) - The region filters applied.
*   `stringtosend` (String) - The string sent to TCP port.
*   `stringtoexpect` (String) - The expected TCP response.
*   `paused` (Boolean) - Whether the check is paused.
*   `responsetime_threshold` (Number) - The response time threshold.
*   `sendnotificationwhendown` (Number) - The failure threshold for alerts.
*   `notifyagainevery` (Number) - The re-notification interval.
*   `notifywhenbackup` (Boolean) - Whether recovery notifications are enabled.
*   `integrationids` (List) - The webhook integration IDs.
*   `userids` (List) - The user IDs for notifications.
*   `teamids` (List) - The team IDs for notifications.

## Check Types

### HTTP/HTTPS Checks

HTTP checks monitor web services and can validate:
- Response content
- HTTP status codes
- Response time
- SSL/TLS certificates
- Custom headers
- Form submissions

```hcl
resource "pingdom_check" "http_example" {
  type       = "http"
  name       = "HTTP Check"
  host       = "example.com"
  resolution = 5
  url        = "/api/health"
  encryption = true
  
  shouldcontain = "healthy"
  responsetime_threshold = 5000
}
```

### Ping Checks

Ping checks monitor network availability using ICMP:
- Network connectivity
- Basic reachability
- Simple availability monitoring

```hcl
resource "pingdom_check" "ping_example" {
  type       = "ping"
  name       = "Ping Check"
  host       = "192.168.1.1"
  resolution = 1
}
```

### TCP Checks

TCP checks monitor specific ports and services:
- Port availability
- Service responsiveness
- Custom protocol validation

```hcl
resource "pingdom_check" "tcp_example" {
  type            = "tcp"
  name            = "TCP Check"
  host            = "service.example.com"
  port            = 8080
  resolution      = 5
  stringtosend    = "HEALTH"
  stringtoexpect  = "OK"
}
```

## Notification Configuration

### User Notifications

Send alerts to specific users:

```hcl
resource "pingdom_check" "user_notifications" {
  type       = "http"
  name       = "User Notifications"
  host       = "example.com"
  resolution = 5
  
  userids = [
    pingdom_contact.admin.id,
    pingdom_contact.ops.id
  ]
}
```

### Team Notifications

Send alerts to teams:

```hcl
resource "pingdom_check" "team_notifications" {
  type       = "http"
  name       = "Team Notifications"
  host       = "example.com"
  resolution = 5
  
  teamids = [
    pingdom_team.oncall.id,
    pingdom_team.operations.id
  ]
}
```

### Webhook Integrations

Integrate with external systems:

```hcl
resource "pingdom_check" "webhook_integration" {
  type       = "http"
  name       = "Webhook Integration"
  host       = "example.com"
  resolution = 5
  
  integrationids = [
    12345678,  # Slack integration
    87654321   # PagerDuty integration
  ]
}
```

## Alert Configuration

### Failure Thresholds

Control when alerts are triggered:

```hcl
resource "pingdom_check" "alert_config" {
  type       = "http"
  name       = "Alert Configuration"
  host       = "example.com"
  resolution = 5
  
  # Alert after 3 consecutive failures
  sendnotificationwhendown = 3
  
  # Re-notify every 10 minutes
  notifyagainevery = 10
  
  # Notify when service recovers
  notifywhenbackup = true
}
```

### Response Time Monitoring

Monitor service performance:

```hcl
resource "pingdom_check" "performance_monitoring" {
  type       = "http"
  name       = "Performance Monitoring"
  host       = "example.com"
  resolution = 5
  
  # Alert if response time exceeds 2 seconds
  responsetime_threshold = 2000
}
```

## Regional Monitoring

Monitor from specific regions:

```hcl
resource "pingdom_check" "regional_monitoring" {
  type       = "http"
  name       = "Regional Monitoring"
  host       = "example.com"
  resolution = 5
  
  # Monitor from North America and Europe
  probefilters = "region:NA,region:EU"
}
```

Available regions:
- `NA` - North America
- `EU` - Europe
- `APAC` - Asia Pacific
- `LATAM` - Latin America

## Tagging and Organization

Organize checks with tags:

```hcl
resource "pingdom_check" "tagged_check" {
  type       = "http"
  name       = "Tagged Check"
  host       = "example.com"
  resolution = 5
  
  # Multiple tags for organization
  tags = "production,api,monitoring,health"
}
```

## Import

Import existing Pingdom checks:

```bash
terraform import pingdom_check.website 12345
```

Where `12345` is the Pingdom check ID.

## Best Practices

### 1. Check Resolution

- **1 minute**: Critical services requiring immediate response
- **5 minutes**: Standard production services
- **15+ minutes**: Non-critical services or development environments

### 2. Alert Configuration

- Use `sendnotificationwhendown = 2` to avoid false positives
- Set `notifyagainevery` to prevent notification spam
- Enable `notifywhenbackup` for recovery notifications

### 3. Content Validation

- Use `shouldcontain` for positive validation
- Use `shouldnotcontain` for negative validation
- Avoid using both simultaneously

### 4. Performance Monitoring

- Set realistic `responsetime_threshold` values
- Consider user experience expectations
- Monitor from relevant regions

### 5. Security

- Store sensitive values in Terraform variables
- Use environment variables for local development
- Mark sensitive variables with `sensitive = true`

## Troubleshooting

### Common Issues

**Check Always Fails**
- Verify the hostname is accessible from the internet
- Check if the service requires authentication
- Ensure the response contains expected content

**SSL/TLS Errors**
- Verify the SSL certificate is valid
- Check if the service supports the required TLS version
- Ensure the port is correct (443 for HTTPS)

**Authentication Failures**
- Verify username/password credentials
- Check if the service requires additional headers
- Ensure the authentication method is supported

**Content Validation Failures**
- Verify the expected content exists in the response
- Check if the content changes dynamically
- Ensure the content is not in JavaScript or other dynamic elements

## Related Resources

- [pingdom_team](../resources/pingdom_team.md) - Create teams for notifications
- [pingdom_contact](../resources/pingdom_contact.md) - Configure notification contacts
- [pingdom_integration](../resources/pingdom_integration.md) - Set up webhook integrations
- [Getting Started Guide](../guides/getting-started.md) - Basic setup and configuration
- [HTTP Monitoring Guide](../guides/http_monitoring.md) - Advanced HTTP monitoring
- [Team Management Guide](../guides/team_management.md) - Notification management

---
page_title: "HTTP Monitoring with Pingdom Provider"
description: "Learn how to configure advanced HTTP/HTTPS monitoring checks with Pingdom"
---

# HTTP Monitoring with Pingdom Provider

This guide covers how to set up comprehensive HTTP and HTTPS monitoring using the Pingdom provider, including authentication, custom headers, response validation, and alert configuration.

## Basic HTTP Check

Start with a simple HTTP check:

```hcl
resource "pingdom_check" "basic_http" {
  type       = "http"
  name       = "Basic HTTP Check"
  host       = "example.com"
  resolution = 5
  url        = "/"
}
```

## HTTPS/SSL Monitoring

Enable encryption for secure monitoring:

```hcl
resource "pingdom_check" "https_check" {
  type       = "http"
  name       = "HTTPS Homepage"
  host       = "example.com"
  resolution = 5
  url        = "/"
  encryption = true  # Use HTTPS
  port       = 443
}
```

## Advanced HTTP Configuration

### Custom Headers and Authentication

```hcl
resource "pingdom_check" "authenticated_api" {
  type       = "http"
  name       = "API Endpoint"
  host       = "api.example.com"
  resolution = 1
  url        = "/health"
  encryption = true
  
  # HTTP Authentication
  username = "api_user"
  password = var.api_password
  
  # Custom Headers
  requestheaders = {
    "User-Agent"    = "Terraform-Pingdom-Provider"
    "Authorization" = "Bearer ${var.api_token}"
    "X-Custom"      = "monitoring"
  }
}
```

### Response Validation

```hcl
resource "pingdom_check" "content_validation" {
  type       = "http"
  name       = "Content Validation Check"
  host       = "example.com"
  resolution = 5
  url        = "/api/status"
  encryption = true
  
  # Content validation
  shouldcontain    = "healthy"
  shouldnotcontain = "error"
  
  # Response time threshold (5 seconds)
  responsetime_threshold = 5000
}
```

### Form Submission and POST Data

```hcl
resource "pingdom_check" "form_submission" {
  type       = "http"
  name       = "Login Form Check"
  host       = "example.com"
  resolution = 5
  url        = "/login"
  encryption = true
  
  # POST data for form submission
  postdata = "username=test&password=test&submit=Login"
  
  # Verify successful login
  shouldcontain = "Welcome"
}
```

## Monitoring Configuration

### Alert Settings

```hcl
resource "pingdom_check" "critical_monitoring" {
  type       = "http"
  name       = "Critical Service"
  host       = "critical.example.com"
  resolution = 1  # Check every minute
  
  # Alert configuration
  sendnotificationwhendown = 2     # Alert after 2 consecutive failures
  notifyagainevery         = 5     # Re-notify every 5 minutes
  notifywhenbackup         = true  # Notify when service recovers
  
  # Response time monitoring
  responsetime_threshold = 3000  # 3 seconds
}
```

### Notification Configuration

```hcl
resource "pingdom_check" "monitored_service" {
  type       = "http"
  name       = "Monitored Service"
  host       = "service.example.com"
  resolution = 5
  
  # Notification targets
  userids = [
    pingdom_contact.admin.id,
    pingdom_contact.ops.id
  ]
  
  teamids = [
    pingdom_team.oncall.id
  ]
  
  # Webhook integrations
  integrationids = [
    12345678  # Your webhook integration ID
  ]
}
```

## Regional Monitoring

Monitor from specific regions to ensure global availability:

```hcl
resource "pingdom_check" "global_monitoring" {
  type       = "http"
  name       = "Global Service"
  host       = "global.example.com"
  resolution = 5
  
  # Monitor from North America and Europe
  probefilters = "region:NA,region:EU"
}
```

## Tagging and Organization

Use tags to organize and filter your checks:

```hcl
resource "pingdom_check" "tagged_check" {
  type       = "http"
  name       = "Tagged Service"
  host       = "tagged.example.com"
  resolution = 5
  
  # Tags for organization
  tags = "production,api,monitoring"
}
```

## Complete Example

Here's a comprehensive HTTP monitoring configuration:

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

variable "api_password" {
  description = "API authentication password"
  type        = string
  sensitive   = true
}

# Main website monitoring
resource "pingdom_check" "main_website" {
  type       = "http"
  name       = "Main Website"
  host       = "example.com"
  resolution = 5
  url        = "/"
  encryption = true
  
  # Content validation
  shouldcontain = "Welcome to Example"
  
  # Performance monitoring
  responsetime_threshold = 5000
  
  # Alert configuration
  sendnotificationwhendown = 2
  notifyagainevery         = 10
  notifywhenbackup         = true
  
  # Notifications
  userids = [pingdom_contact.admin.id]
  teamids = [pingdom_team.oncall.id]
  
  # Organization
  tags = "production,website,monitoring"
}

# API health check
resource "pingdom_check" "api_health" {
  type       = "http"
  name       = "API Health Check"
  host       = "api.example.com"
  resolution = 1
  url        = "/health"
  encryption = true
  
  # Authentication
  username = "monitor"
  password = var.api_password
  
  # Custom headers
  requestheaders = {
    "Accept"        = "application/json"
    "User-Agent"    = "Pingdom-Monitor"
    "X-Monitoring"  = "true"
  }
  
  # Response validation
  shouldcontain = "healthy"
  
  # Performance
  responsetime_threshold = 2000
  
  # Critical alerts
  sendnotificationwhendown = 1
  notifyagainevery         = 2
  
  # Notifications
  userids = [pingdom_contact.admin.id, pingdom_contact.ops.id]
  
  tags = "production,api,critical"
}
```

## Best Practices

### 1. Resolution Selection
- **1 minute**: Critical services that need immediate alerts
- **5 minutes**: Standard production services
- **15+ minutes**: Non-critical services or development environments

### 2. Response Time Thresholds
- **2-3 seconds**: Critical user-facing services
- **5 seconds**: Standard web services
- **10+ seconds**: Background services or APIs

### 3. Alert Configuration
- Use `sendnotificationwhendown = 2` to avoid false positives
- Set `notifyagainevery` to prevent notification spam
- Enable `notifywhenbackup` for recovery notifications

### 4. Content Validation
- Use `shouldcontain` for positive validation
- Use `shouldnotcontain` for negative validation
- Avoid using both simultaneously

### 5. Security
- Store sensitive values in Terraform variables
- Use environment variables for local development
- Mark sensitive variables with `sensitive = true`

## Troubleshooting

### Common HTTP Check Issues

**Check Always Fails**
- Verify the URL is accessible from the internet
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

## Next Steps

- [Set up teams and contacts](team_management.md) for alert notifications
- [Configure ping monitoring](ping_monitoring.md) for network availability
- [Create TCP port checks](tcp_monitoring.md) for service monitoring
- [Set up webhook integrations](../resources/pingdom_integration.md) for custom alert handling

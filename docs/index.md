---
page_title: "Pingdom Provider"
description: |-
  The Pingdom provider is used to interact with Pingdom's website monitoring API.
  The provider needs to be configured with the proper credentials before it can be used.
---

# Pingdom Provider

The Terraform Provider for Pingdom allows you to manage Pingdom monitoring checks, teams, and contacts using Terraform. This provider enables infrastructure as code for website and service monitoring, letting you programmatically create, update, and manage Pingdom checks, teams, and notification contacts.

Use the navigation to the left to read about the available resources and data sources.

## Quick Start Guides

- [Getting Started Guide](guides/getting-started.md) - Basic setup and configuration
- [HTTP Monitoring Guide](guides/http_monitoring.md) - Monitor HTTP/HTTPS endpoints
- [Team Management Guide](guides/team_management.md) - Manage teams and contacts for alerts

## Example Usage

```hcl
terraform {
  required_providers {
    pingdom = {
      source = "AdconnectDevOps/pingdom"
      version = "~> 1.0"
    }
  }
}

provider "pingdom" {
  api_token = var.pingdom_api_token
}

# HTTP check for website monitoring
resource "pingdom_check" "website" {
  type       = "http"
  name       = "Website Homepage"
  host       = "example.com"
  resolution = 5
  url        = "/"
  encryption = true  # Use HTTPS
  
  userids = [
    pingdom_contact.admin.id
  ]
  
  teamids = [
    pingdom_team.oncall.id
  ]
}

# Team for alert notifications
resource "pingdom_team" "oncall" {
  name = "On-Call Team"
  member_ids = [
    pingdom_contact.admin.id,
    pingdom_contact.ops.id
  ]
}

# Contact for notifications
resource "pingdom_contact" "admin" {
  name = "Admin User"
  
  email_notification {
    address  = "admin@example.com"
    severity = "HIGH"
  }
  
  sms_notification {
    number   = "5555555555"
    severity = "HIGH"
  }
}

# Get existing contact information
data "pingdom_contact" "existing" {
  name = "existing-contact"
}
```

## Authentication

The Pingdom provider requires an API token to authenticate with Pingdom's services. You can provide the API token via the `api_token` argument in the provider configuration block, or via the `PINGDOM_API_TOKEN` environment variable.

```hcl
provider "pingdom" {
  api_token = "your-pingdom-api-token"
}
```

## Features

- **HTTP/HTTPS Monitoring**: Monitor websites and web services with custom headers and authentication
- **Ping Monitoring**: Basic ICMP ping monitoring for network availability
- **TCP Port Monitoring**: Monitor specific ports for service availability
- **Team Management**: Create and manage teams for alert notifications
- **Contact Management**: Configure SMS and email notifications with severity levels
- **Integration Support**: Webhook integrations for custom alert handling
- **Terraform 1.0+ Compatible**: Built with the latest Terraform plugin framework

## Getting Started

1. **Install the provider** by adding it to your Terraform configuration
2. **Configure authentication** with your Pingdom API token
3. **Create your first check** using the `pingdom_check` resource
4. **Set up contacts and teams** for alert notifications
5. **Monitor your services** for availability and performance

## Support

- **Documentation**: [GitHub Repository](https://github.com/russellcardullo/terraform-provider-pingdom)
- **Issues**: [GitHub Issues](https://github.com/russellcardullo/terraform-provider-pingdom/issues)
- **Discussions**: [GitHub Discussions](https://github.com/russellcardullo/terraform-provider-pingdom/discussions)

## License

This project is licensed under the MIT License.

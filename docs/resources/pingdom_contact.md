---
page_title: "pingdom_contact"
description: "Manages a Pingdom contact for alert notifications"
---

# pingdom_contact Resource

The `pingdom_contact` resource allows you to create, update, and delete Pingdom contacts. Contacts are individual users who can receive notifications when monitoring checks fail, supporting both SMS and email notifications with configurable severity levels.

## Example Usage

### Basic Email Contact

```hcl
resource "pingdom_contact" "admin" {
  name = "Admin User"
  
  email_notification {
    address  = "admin@example.com"
    severity = "HIGH"
  }
}
```

### Contact with Multiple Notification Methods

```hcl
resource "pingdom_contact" "ops_engineer" {
  name = "Operations Engineer"
  
  # High severity notifications
  email_notification {
    address  = "ops@example.com"
    severity = "HIGH"
  }
  
  sms_notification {
    number   = "5555555555"
    severity = "HIGH"
  }
  
  # Low severity notifications
  email_notification {
    address  = "ops-low@example.com"
    severity = "LOW"
  }
}
```

### SMS-Only Contact

```hcl
resource "pingdom_contact" "oncall" {
  name = "On-Call Engineer"
  
  sms_notification {
    number       = "5555555555"
    country_code = "1"
    provider     = "nexmo"
    severity     = "HIGH"
  }
}
```

### Paused Contact

```hcl
resource "pingdom_contact" "paused_contact" {
  name   = "Paused Contact"
  paused = true  # Disable all notifications
  
  email_notification {
    address  = "user@example.com"
    severity = "HIGH"
  }
}
```

## Argument Reference

### Required Arguments

*   `name` (String) - The name of the contact. Must be unique within your account.

### Optional Arguments

*   `paused` (Boolean) - Whether alerts for this contact should be disabled (default: `false`).

### Email Notification Block

The `email_notification` block supports the following arguments:

*   `address` (String, Required) - The email address to send notifications to.
*   `severity` (String, Required) - The severity level for this notification. Valid values: `HIGH`, `LOW`.

### SMS Notification Block

The `sms_notification` block supports the following arguments:

*   `number` (String, Required) - The phone number to send SMS notifications to.
*   `country_code` (String, Optional) - The country code for the phone number (default: `"1"`).
*   `provider` (String, Optional) - The SMS provider to use. Valid values: `nexmo`, `esendex`, `cellsynt`.
*   `severity` (String, Required) - The severity level for this notification. Valid values: `HIGH`, `LOW`.

## Attributes Reference

*   `id` (String) - The ID of the Pingdom contact.
*   `name` (String) - The name of the contact.
*   `paused` (Boolean) - Whether the contact is paused.

## Notification Configuration

### Severity Levels

Pingdom supports two severity levels for notifications:

**HIGH Severity**
- Used for critical service failures
- Triggers immediate notifications
- Typically includes SMS for urgent response
- Used with 1-minute resolution checks

**LOW Severity**
- Used for non-critical issues
- Email notifications only
- Used for summary reports and management
- Used with 5+ minute resolution checks

### SMS Providers

**nexmo** (Vonage)
- Global coverage
- Reliable delivery
- Recommended for most use cases

**esendex**
- UK-based provider
- Good for European coverage
- Competitive pricing

**cellsynt**
- European provider
- Good for EU coverage
- Reliable service

**bulksms**
- Currently not operational
- Not recommended for new implementations

### Country Codes

Common country codes:
- `"1"` - United States and Canada (default)
- `"44"` - United Kingdom
- `"33"` - France
- `"49"` - Germany
- `"81"` - Japan
- `"86"` - China
- `"91"` - India

## Contact Management Strategies

### On-Call Contacts

```hcl
# Primary on-call contact
resource "pingdom_contact" "primary_oncall" {
  name = "Primary On-Call"
  
  # High severity - immediate response
  sms_notification {
    number   = var.primary_phone
    severity = "HIGH"
  }
  
  email_notification {
    address  = var.primary_email
    severity = "HIGH"
  }
}

# Secondary on-call contact
resource "pingdom_contact" "secondary_oncall" {
  name = "Secondary On-Call"
  
  # High severity - backup response
  sms_notification {
    number   = var.secondary_phone
    severity = "HIGH"
  }
  
  email_notification {
    address  = var.secondary_email
    severity = "HIGH"
  }
}
```

### Operations Team Contacts

```hcl
# Operations engineer
resource "pingdom_contact" "ops_engineer" {
  name = "Operations Engineer"
  
  # High severity - immediate response
  email_notification {
    address  = "ops@example.com"
    severity = "HIGH"
  }
  
  # Low severity - summary reports
  email_notification {
    address  = "ops-summary@example.com"
    severity = "LOW"
  }
}

# DevOps engineer
resource "pingdom_contact" "devops_engineer" {
  name = "DevOps Engineer"
  
  # High severity - infrastructure issues
  email_notification {
    address  = "devops@example.com"
    severity = "HIGH"
  }
  
  sms_notification {
    number   = var.devops_phone
    severity = "HIGH"
  }
}
```

### Management Contacts

```hcl
# Operations manager
resource "pingdom_contact" "ops_manager" {
  name = "Operations Manager"
  
  # Low severity - summary and reporting
  email_notification {
    address  = "ops-manager@example.com"
    severity = "LOW"
  }
}

# CTO
resource "pingdom_contact" "cto" {
  name = "CTO"
  
  # Low severity - executive summary
  email_notification {
    address  = "cto@example.com"
    severity = "LOW"
  }
}
```

## Complete Contact Management Example

Here's a comprehensive example of contact management:

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

variable "primary_phone" {
  description = "Primary on-call phone number"
  type        = string
  sensitive   = true
}

variable "secondary_phone" {
  description = "Secondary on-call phone number"
  type        = string
  sensitive   = true
}

variable "primary_email" {
  description = "Primary on-call email"
  type        = string
}

variable "secondary_email" {
  description = "Secondary on-call email"
  type        = string
}

# Primary on-call contact
resource "pingdom_contact" "primary_oncall" {
  name = "Primary On-Call"
  
  # High severity - SMS and email
  sms_notification {
    number   = var.primary_phone
    severity = "HIGH"
  }
  
  email_notification {
    address  = var.primary_email
    severity = "HIGH"
  }
}

# Secondary on-call contact
resource "pingdom_contact" "secondary_oncall" {
  name = "Secondary On-Call"
  
  # High severity - SMS and email
  sms_notification {
    number   = var.secondary_phone
    severity = "HIGH"
  }
  
  email_notification {
    address  = var.secondary_email
    severity = "HIGH"
  }
}

# Operations team contact
resource "pingdom_contact" "ops_team" {
  name = "Operations Team"
  
  # High severity - immediate email
  email_notification {
    address  = "ops@example.com"
    severity = "HIGH"
  }
  
  # Low severity - summary email
  email_notification {
    address  = "ops-summary@example.com"
    severity = "LOW"
  }
}

# DevOps engineer contact
resource "pingdom_contact" "devops_engineer" {
  name = "DevOps Engineer"
  
  # High severity - infrastructure issues
  email_notification {
    address  = "devops@example.com"
    severity = "HIGH"
  }
  
  sms_notification {
    number   = var.devops_phone
    severity = "HIGH"
  }
}

# Management contact
resource "pingdom_contact" "management" {
  name = "Management"
  
  # Only low severity - summary reports
  email_notification {
    address  = "management@example.com"
    severity = "LOW"
  }
}

# Use contacts in monitoring checks
resource "pingdom_check" "critical_service" {
  type       = "http"
  name       = "Critical Service"
  host       = "critical.example.com"
  resolution = 1
  
  # High severity notifications
  userids = [
    pingdom_contact.primary_oncall.id,
    pingdom_contact.secondary_oncall.id
  ]
}

resource "pingdom_check" "standard_service" {
  type       = "http"
  name       = "Standard Service"
  host       = "standard.example.com"
  resolution = 5
  
  # Standard notifications
  userids = [
    pingdom_contact.ops_team.id
  ]
}

resource "pingdom_check" "summary_reporting" {
  type       = "http"
  name       = "Summary Reporting"
  host       = "reporting.example.com"
  resolution = 15
  
  # Low priority notifications
  userids = [
    pingdom_contact.management.id
  ]
}
```

## Best Practices

### 1. Severity Level Strategy

**High Severity (HIGH)**
- Use for critical service failures
- Include SMS notifications for immediate response
- Use with 1-minute resolution checks
- Include primary on-call personnel

**Low Severity (LOW)**
- Use for non-critical issues
- Email notifications only
- Use for summary reports and management
- Use with 5+ minute resolution checks

### 2. Contact Organization

- **Primary contacts**: Immediate response personnel
- **Secondary contacts**: Backup response personnel
- **Specialist contacts**: Technical experts for specific issues
- **Management contacts**: Summary and reporting

### 3. Notification Methods

- **SMS**: For critical, immediate response needs
- **Email**: For detailed information and summaries
- **Combined**: Use both for critical contacts
- **Paused**: Temporarily disable contacts when needed

### 4. Contact Management

- Keep contact information current
- Test notifications regularly
- Use consistent naming conventions
- Document contact responsibilities

### 5. Security

- Store sensitive values in Terraform variables
- Use environment variables for local development
- Mark sensitive variables with `sensitive = true`
- Rotate contact information regularly

## Import

Import existing Pingdom contacts:

```bash
terraform import pingdom_contact.admin 12345
```

Where `12345` is the Pingdom contact ID.

## Troubleshooting

### Common Contact Issues

**SMS Notifications Not Working**
- Verify phone number format
- Check country code setting
- Ensure SMS provider is operational
- Verify account has SMS credits

**Email Notifications Not Working**
- Check email address format
- Verify email server accessibility
- Check spam/junk folders
- Ensure email provider allows Pingdom

**Contact Creation Fails**
- Verify contact name uniqueness
- Check notification configuration
- Ensure all required fields are provided
- Verify account permissions

**Contact Updates Not Applied**
- Verify contact ID is correct
- Check for conflicting names
- Ensure notification configuration is valid

### Contact Validation

```hcl
# Test contact configuration
resource "pingdom_contact" "test_contact" {
  name = "Test Contact"
  
  email_notification {
    address  = "test@example.com"
    severity = "HIGH"
  }
  
  sms_notification {
    number   = "5555555555"
    severity = "HIGH"
  }
}

# Verify contact in monitoring check
resource "pingdom_check" "test_check" {
  type       = "http"
  name       = "Test Check"
  host       = "test.example.com"
  resolution = 5
  
  userids = [
    pingdom_contact.test_contact.id
  ]
}
```

## Related Resources

- [pingdom_team](../resources/pingdom_team.md) - Create teams for notifications
- [pingdom_check](../resources/pingdom_check.md) - Create monitoring checks
- [pingdom_integration](../resources/pingdom_integration.md) - Set up webhook integrations
- [Team Management Guide](../guides/team_management.md) - Comprehensive team management
- [Getting Started Guide](../guides/getting-started.md) - Basic setup and configuration

---
page_title: "Team Management with Pingdom Provider"
description: "Learn how to manage teams and contacts for alert notifications in Pingdom"
---

# Team Management with Pingdom Provider

This guide covers how to set up and manage teams and contacts for alert notifications using the Pingdom provider, including SMS and email notifications with different severity levels.

## Overview

Pingdom allows you to create teams and contacts to manage who receives notifications when your checks fail. This system provides:

- **Contacts**: Individual users who can receive notifications
- **Teams**: Groups of contacts for organized alert management
- **Multiple notification methods**: SMS and email with severity levels
- **Flexible alert routing**: Different contacts for different severity levels

## Creating Contacts

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

## Contact Configuration Options

### SMS Notification Settings

```hcl
resource "pingdom_contact" "sms_contact" {
  name = "SMS Contact"
  
  sms_notification {
    number       = "5555555555"
    country_code = "1"        # Default: "1" (US/Canada)
    provider     = "nexmo"    # Options: nexmo, esendex, cellsynt
    severity     = "HIGH"     # HIGH or LOW
  }
}
```

**SMS Providers:**
- **nexmo**: Vonage (formerly Nexmo) - Global coverage
- **esendex**: UK-based provider
- **cellsynt**: European provider
- **bulksms**: Currently not operational

### Email Notification Settings

```hcl
resource "pingdom_contact" "email_contact" {
  name = "Email Contact"
  
  email_notification {
    address  = "user@example.com"
    severity = "LOW"  # HIGH or LOW
  }
}
```

### Paused Contacts

Disable notifications for a contact temporarily:

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

## Creating Teams

### Basic Team

```hcl
resource "pingdom_team" "oncall" {
  name = "On-Call Team"
  
  member_ids = [
    pingdom_contact.admin.id,
    pingdom_contact.ops_engineer.id
  ]
}
```

### Team with Multiple Members

```hcl
resource "pingdom_team" "production_support" {
  name = "Production Support Team"
  
  member_ids = [
    pingdom_contact.admin.id,
    pingdom_contact.ops_engineer.id,
    pingdom_contact.oncall.id,
    pingdom_contact.manager.id
  ]
}
```

## Severity-Based Notification Strategy

### High Severity (Critical Issues)

```hcl
resource "pingdom_contact" "critical_alerts" {
  name = "Critical Alert Contact"
  
  # Immediate SMS for critical issues
  sms_notification {
    number   = "5555555555"
    severity = "HIGH"
  }
  
  # High priority email
  email_notification {
    address  = "critical@example.com"
    severity = "HIGH"
  }
}
```

### Low Severity (Non-Critical Issues)

```hcl
resource "pingdom_contact" "low_priority" {
  name = "Low Priority Contact"
  
  # Only email for non-critical issues
  email_notification {
    address  = "low-priority@example.com"
    severity = "LOW"
  }
}
```

## Complete Team Management Example

Here's a comprehensive example of team and contact management:

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

# Management contact
resource "pingdom_contact" "management" {
  name = "Management"
  
  # Only low severity - summary reports
  email_notification {
    address  = "management@example.com"
    severity = "LOW"
  }
}

# On-call team
resource "pingdom_team" "oncall" {
  name = "On-Call Team"
  
  member_ids = [
    pingdom_contact.primary_oncall.id,
    pingdom_contact.secondary_oncall.id
  ]
}

# Operations team
resource "pingdom_team" "operations" {
  name = "Operations Team"
  
  member_ids = [
    pingdom_contact.ops_team.id,
    pingdom_contact.primary_oncall.id
  ]
}

# Management team
resource "pingdom_team" "management" {
  name = "Management Team"
  
  member_ids = [
    pingdom_contact.management.id
  ]
}

# Use teams in monitoring checks
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
  
  teamids = [
    pingdom_team.oncall.id
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
  
  teamids = [
    pingdom_team.operations.id
  ]
}
```

## Best Practices

### 1. Severity Level Strategy

**High Severity (HIGH)**
- Use for critical service failures
- Send SMS notifications for immediate response
- Include primary on-call personnel
- Use 1-minute resolution for critical checks

**Low Severity (LOW)**
- Use for non-critical issues
- Email notifications only
- Include management for reporting
- Use 5+ minute resolution

### 2. Contact Organization

- **Primary contacts**: Immediate response personnel
- **Secondary contacts**: Backup response personnel
- **Team contacts**: Group notifications for coordination
- **Management contacts**: Summary and reporting

### 3. Notification Timing

- **Critical issues**: Immediate SMS + email
- **Standard issues**: Email only
- **Recovery notifications**: Enable `notifywhenbackup`
- **Re-notification**: Set `notifyagainevery` appropriately

### 4. Team Structure

- **On-call teams**: Small, focused teams for immediate response
- **Operations teams**: Larger teams for coordination
- **Management teams**: High-level oversight and reporting

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

**Team Notifications Issues**
- Verify all team members exist
- Check member contact configurations
- Ensure team has at least one member
- Verify team member permissions

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

# Test team
resource "pingdom_team" "test_team" {
  name = "Test Team"
  
  member_ids = [
    pingdom_contact.test_contact.id
  ]
}
```

## Next Steps

- [Set up HTTP monitoring](http_monitoring.md) with your new teams
- [Configure ping monitoring](ping_monitoring.md) for network checks
- [Create TCP port checks](tcp_monitoring.md) for service monitoring
- [Set up webhook integrations](../resources/pingdom_integration.md) for custom notifications

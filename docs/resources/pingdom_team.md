---
page_title: "pingdom_team"
description: "Manages a Pingdom team for alert notifications"
---

# pingdom_team Resource

The `pingdom_team` resource allows you to create, update, and delete Pingdom teams. Teams are groups of contacts that can receive notifications when monitoring checks fail, providing an organized way to manage alert routing and escalation.

## Example Usage

### Basic Team

```hcl
resource "pingdom_team" "oncall" {
  name = "On-Call Team"
  
  member_ids = [
    pingdom_contact.admin.id,
    pingdom_contact.ops.id
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

### Team for Different Severity Levels

```hcl
# High severity team for critical issues
resource "pingdom_team" "critical_alerts" {
  name = "Critical Alerts Team"
  
  member_ids = [
    pingdom_contact.primary_oncall.id,
    pingdom_contact.secondary_oncall.id,
    pingdom_contact.ops_manager.id
  ]
}

# Low severity team for non-critical issues
resource "pingdom_team" "low_priority" {
  name = "Low Priority Team"
  
  member_ids = [
    pingdom_contact.ops_engineer.id,
    pingdom_contact.manager.id
  ]
}
```

## Argument Reference

### Required Arguments

*   `name` (String) - The name of the team. Must be unique within your account.

### Optional Arguments

*   `member_ids` (List) - List of contact IDs that will be members of the team.

## Attributes Reference

*   `id` (String) - The ID of the Pingdom team.
*   `name` (String) - The name of the team.
*   `member_ids` (List) - The list of contact IDs in the team.

## Team Management Strategies

### On-Call Rotation

Create teams for different on-call schedules:

```hcl
# Primary on-call team
resource "pingdom_team" "primary_oncall" {
  name = "Primary On-Call"
  
  member_ids = [
    pingdom_contact.engineer_1.id,
    pingdom_contact.engineer_2.id
  ]
}

# Secondary on-call team
resource "pingdom_contact" "secondary_oncall" {
  name = "Secondary On-Call"
  
  member_ids = [
    pingdom_contact.engineer_3.id,
    pingdom_contact.engineer_4.id
  ]
}

# Use teams in monitoring checks
resource "pingdom_check" "critical_service" {
  type       = "http"
  name       = "Critical Service"
  host       = "critical.example.com"
  resolution = 1
  
  # Primary team gets immediate alerts
  teamids = [
    pingdom_team.primary_oncall.id
  ]
}
```

### Escalation Teams

Create teams for different escalation levels:

```hcl
# Level 1 - Initial response
resource "pingdom_team" "level1" {
  name = "Level 1 Support"
  
  member_ids = [
    pingdom_contact.support_engineer_1.id,
    pingdom_contact.support_engineer_2.id
  ]
}

# Level 2 - Escalation
resource "pingdom_team" "level2" {
  name = "Level 2 Support"
  
  member_ids = [
    pingdom_contact.senior_engineer_1.id,
    pingdom_contact.senior_engineer_2.id
  ]
}

# Level 3 - Management escalation
resource "pingdom_team" "level3" {
  name = "Management Escalation"
  
  member_ids = [
    pingdom_contact.ops_manager.id,
    pingdom_contact.cto.id
  ]
}
```

### Functional Teams

Organize teams by function or responsibility:

```hcl
# Infrastructure team
resource "pingdom_team" "infrastructure" {
  name = "Infrastructure Team"
  
  member_ids = [
    pingdom_contact.sysadmin_1.id,
    pingdom_contact.sysadmin_2.id,
    pingdom_contact.network_engineer.id
  ]
}

# Application team
resource "pingdom_team" "application" {
  name = "Application Team"
  
  member_ids = [
    pingdom_contact.dev_engineer_1.id,
    pingdom_contact.dev_engineer_2.id,
    pingdom_contact.qa_engineer.id
  ]
}

# Database team
resource "pingdom_team" "database" {
  name = "Database Team"
  
  member_ids = [
    pingdom_contact.dba_1.id,
    pingdom_contact.dba_2.id
  ]
}
```

## Complete Team Management Example

Here's a comprehensive example of team management:

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

# Contact definitions
resource "pingdom_contact" "primary_oncall" {
  name = "Primary On-Call"
  
  sms_notification {
    number   = var.primary_phone
    severity = "HIGH"
  }
  
  email_notification {
    address  = var.primary_email
    severity = "HIGH"
  }
}

resource "pingdom_contact" "secondary_oncall" {
  name = "Secondary On-Call"
  
  sms_notification {
    number   = var.secondary_phone
    severity = "HIGH"
  }
  
  email_notification {
    address  = var.secondary_email
    severity = "HIGH"
  }
}

resource "pingdom_contact" "ops_team" {
  name = "Operations Team"
  
  email_notification {
    address  = "ops@example.com"
    severity = "HIGH"
  }
  
  email_notification {
    address  = "ops-summary@example.com"
    severity = "LOW"
  }
}

resource "pingdom_contact" "management" {
  name = "Management"
  
  email_notification {
    address  = "management@example.com"
    severity = "LOW"
  }
}

# Team definitions
resource "pingdom_team" "oncall" {
  name = "On-Call Team"
  
  member_ids = [
    pingdom_contact.primary_oncall.id,
    pingdom_contact.secondary_oncall.id
  ]
}

resource "pingdom_team" "operations" {
  name = "Operations Team"
  
  member_ids = [
    pingdom_contact.ops_team.id,
    pingdom_contact.primary_oncall.id
  ]
}

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

resource "pingdom_check" "summary_reporting" {
  type       = "http"
  name       = "Summary Reporting"
  host       = "reporting.example.com"
  resolution = 15
  
  # Low priority notifications
  teamids = [
    pingdom_team.management.id
  ]
}
```

## Best Practices

### 1. Team Size

- **Small teams (2-3 members)**: For immediate response and on-call rotations
- **Medium teams (4-6 members)**: For functional teams and support groups
- **Large teams (7+ members)**: For management and reporting teams

### 2. Team Composition

- **Primary contacts**: Include in multiple teams for redundancy
- **Specialists**: Group by technical expertise or responsibility
- **Escalation paths**: Create clear escalation hierarchies

### 3. Naming Conventions

- Use descriptive names that indicate purpose
- Include function and level in team names
- Use consistent naming patterns across teams

### 4. Member Management

- Keep team membership current
- Review and update teams regularly
- Ensure coverage for all critical functions

### 5. Notification Strategy

- Use teams for group notifications
- Combine with individual contacts for critical alerts
- Test team notifications regularly

## Import

Import existing Pingdom teams:

```bash
terraform import pingdom_team.oncall 12345
```

Where `12345` is the Pingdom team ID.

## Troubleshooting

### Common Team Issues

**Team Notifications Not Working**
- Verify all team members exist and are configured
- Check team member contact configurations
- Ensure team has at least one member
- Verify team member permissions

**Team Creation Fails**
- Check team name uniqueness
- Verify all member IDs are valid
- Ensure you have permission to create teams

**Team Updates Not Applied**
- Verify team ID is correct
- Check for conflicting team names
- Ensure all member IDs are valid

### Team Validation

```hcl
# Test team configuration
resource "pingdom_team" "test_team" {
  name = "Test Team"
  
  member_ids = [
    pingdom_contact.test_contact.id
  ]
}

# Verify team in monitoring check
resource "pingdom_check" "test_check" {
  type       = "http"
  name       = "Test Check"
  host       = "test.example.com"
  resolution = 5
  
  teamids = [
    pingdom_team.test_team.id
  ]
}
```

## Related Resources

- [pingdom_contact](../resources/pingdom_contact.md) - Configure notification contacts
- [pingdom_check](../resources/pingdom_check.md) - Create monitoring checks
- [pingdom_integration](../resources/pingdom_integration.md) - Set up webhook integrations
- [Team Management Guide](../guides/team_management.md) - Comprehensive team management
- [Getting Started Guide](../guides/getting-started.md) - Basic setup and configuration

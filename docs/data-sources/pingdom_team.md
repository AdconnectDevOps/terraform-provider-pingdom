---
page_title: "pingdom_team"
description: "Retrieves information about an existing Pingdom team"
---

# pingdom_team Data Source

The `pingdom_team` data source allows you to retrieve information about an existing Pingdom team. This is useful for referencing existing teams in your Terraform configuration without creating new ones.

## Example Usage

### Basic Team Lookup

```hcl
data "pingdom_team" "existing_oncall" {
  name = "On-Call Team"
}

# Use the team in a monitoring check
resource "pingdom_check" "website" {
  type       = "http"
  name       = "Website"
  host       = "example.com"
  resolution = 5
  
  teamids = [
    data.pingdom_team.existing_oncall.id
  ]
}
```

### Team with Multiple References

```hcl
data "pingdom_team" "operations" {
  name = "Operations Team"
}

data "pingdom_team" "management" {
  name = "Management Team"
}

# Use multiple teams in a check
resource "pingdom_check" "critical_service" {
  type       = "http"
  name       = "Critical Service"
  host       = "critical.example.com"
  resolution = 1
  
  teamids = [
    data.pingdom_team.operations.id,
    data.pingdom_team.management.id
  ]
}
```

### Team in New Resources

```hcl
# Look up existing team
data "pingdom_team" "existing_team" {
  name = "Existing Team"
}

# Create new team using existing team as reference
resource "pingdom_team" "new_team" {
  name = "New Team"
  
  member_ids = [
    pingdom_contact.new_member.id,
    # Include some members from existing team
    pingdom_contact.existing_member.id
  ]
}

# Use both teams in monitoring
resource "pingdom_check" "service_monitoring" {
  type       = "http"
  name       = "Service Monitoring"
  host       = "service.example.com"
  resolution = 5
  
  teamids = [
    data.pingdom_team.existing_team.id,
    pingdom_team.new_team.id
  ]
}
```

## Argument Reference

### Required Arguments

*   `name` (String) - The name of the existing Pingdom team to retrieve.

## Attributes Reference

*   `id` (String) - The ID of the Pingdom team.
*   `name` (String) - The name of the team.

## Use Cases

### Referencing Existing Teams

Use data sources when you want to reference teams that already exist in Pingdom:

```hcl
# Reference existing team
data "pingdom_team" "existing_team" {
  name = "Existing Team Name"
}

# Use in new resources
resource "pingdom_check" "new_check" {
  type       = "http"
  name       = "New Check"
  host       = "example.com"
  resolution = 5
  
  teamids = [
    data.pingdom_team.existing_team.id
  ]
}
```

### Migrating Existing Infrastructure

When migrating existing Pingdom configurations to Terraform:

```hcl
# Look up existing teams
data "pingdom_team" "oncall" {
  name = "On-Call Team"
}

data "pingdom_team" "operations" {
  name = "Operations Team"
}

# Create new monitoring checks using existing teams
resource "pingdom_check" "migrated_check" {
  type       = "http"
  name       = "Migrated Check"
  host       = "example.com"
  resolution = 5
  
  teamids = [
    data.pingdom_team.oncall.id,
    data.pingdom_team.operations.id
  ]
}
```

### Hybrid Team Management

Combine data sources with new team resources:

```hcl
# Use existing teams
data "pingdom_team" "existing_oncall" {
  name = "Existing On-Call Team"
}

# Create new team
resource "pingdom_team" "new_operations" {
  name = "New Operations Team"
  
  member_ids = [
    pingdom_contact.ops_engineer_1.id,
    pingdom_contact.ops_engineer_2.id
  ]
}

# Use both teams in monitoring
resource "pingdom_check" "hybrid_monitoring" {
  type       = "http"
  name       = "Hybrid Monitoring"
  host       = "example.com"
  resolution = 5
  
  teamids = [
    data.pingdom_team.existing_oncall.id,
    pingdom_team.new_operations.id
  ]
}
```

### Team Escalation

Use existing teams for escalation strategies:

```hcl
# Look up escalation teams
data "pingdom_team" "level1" {
  name = "Level 1 Support"
}

data "pingdom_team" "level2" {
  name = "Level 2 Support"
}

data "pingdom_team" "management" {
  name = "Management Escalation"
}

# Create checks with different escalation levels
resource "pingdom_check" "standard_service" {
  type       = "http"
  name       = "Standard Service"
  host       = "standard.example.com"
  resolution = 5
  
  teamids = [
    data.pingdom_team.level1.id
  ]
}

resource "pingdom_check" "critical_service" {
  type       = "http"
  name       = "Critical Service"
  host       = "critical.example.com"
  resolution = 1
  
  teamids = [
    data.pingdom_team.level1.id,
    data.pingdom_team.level2.id
  ]
}

resource "pingdom_check" "escalated_service" {
  type       = "http"
  name       = "Escalated Service"
  host       = "escalated.example.com"
  resolution = 1
  
  teamids = [
    data.pingdom_team.level1.id,
    data.pingdom_team.level2.id,
    data.pingdom_team.management.id
  ]
}
```

## Best Practices

### 1. Team Naming

- Use exact names that match existing Pingdom teams
- Ensure team names are unique and descriptive
- Document the expected team names in your team

### 2. Error Handling

- Verify teams exist before referencing them
- Use consistent naming conventions
- Handle cases where teams might not exist

### 3. Data Source Organization

- Group related data sources together
- Use descriptive variable names
- Document the purpose of each data source

### 4. Migration Strategy

- Start with data sources for existing teams
- Gradually migrate to managed resources
- Test thoroughly before production deployment

### 5. Team Hierarchy

- Understand existing team structures
- Maintain escalation paths
- Document team responsibilities

## Troubleshooting

### Common Issues

**Team Not Found**
- Verify the team name exactly matches Pingdom
- Check for typos or case sensitivity
- Ensure the team exists in your Pingdom account

**Data Source Errors**
- Verify your API token has read access
- Check Pingdom API connectivity
- Ensure the team name is correct

**Import Errors**
- Verify the team exists before importing
- Check for naming conflicts
- Ensure proper permissions

**Team Member Issues**
- Verify team members exist
- Check team member permissions
- Ensure team has at least one member

### Validation

```hcl
# Test data source configuration
data "pingdom_team" "test_team" {
  name = "Test Team"
}

# Output the team ID for verification
output "team_id" {
  value = data.pingdom_team.test_team.id
}

# Use in a simple check
resource "pingdom_check" "test_check" {
  type       = "http"
  name       = "Test Check"
  host       = "test.example.com"
  resolution = 5
  
  teamids = [
    data.pingdom_team.test_team.id
  ]
}
```

### Team Verification

```hcl
# Verify team exists and has members
data "pingdom_team" "verified_team" {
  name = "Verified Team"
}

# Create a test check to verify team functionality
resource "pingdom_check" "team_verification" {
  type       = "http"
  name       = "Team Verification"
  host       = "verify.example.com"
  resolution = 15
  
  teamids = [
    data.pingdom_team.verified_team.id
  ]
  
  # Use a long resolution to avoid excessive notifications during testing
}
```

## Related Resources

- [pingdom_team](../resources/pingdom_team.md) - Create and manage teams
- [pingdom_contact](../resources/pingdom_contact.md) - Configure notification contacts
- [pingdom_check](../resources/pingdom_check.md) - Create monitoring checks
- [Team Management Guide](../guides/team_management.md) - Comprehensive team management
- [Getting Started Guide](../guides/getting-started.md) - Basic setup and configuration

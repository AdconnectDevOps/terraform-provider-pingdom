---
page_title: "pingdom_contact"
description: "Retrieves information about an existing Pingdom contact"
---

# pingdom_contact Data Source

The `pingdom_contact` data source allows you to retrieve information about an existing Pingdom contact. This is useful for referencing existing contacts in your Terraform configuration without creating new ones.

## Example Usage

### Basic Contact Lookup

```hcl
data "pingdom_contact" "existing_admin" {
  name = "Admin User"
}

# Use the contact in a monitoring check
resource "pingdom_check" "website" {
  type       = "http"
  name       = "Website"
  host       = "example.com"
  resolution = 5
  
  userids = [
    data.pingdom_contact.existing_admin.id
  ]
}
```

### Contact with Multiple Notification Methods

```hcl
data "pingdom_contact" "ops_engineer" {
  name = "Operations Engineer"
}

# Use the contact in a team
resource "pingdom_team" "oncall" {
  name = "On-Call Team"
  
  member_ids = [
    data.pingdom_contact.ops_engineer.id
  ]
}
```

### Multiple Contact References

```hcl
# Look up multiple existing contacts
data "pingdom_contact" "primary_oncall" {
  name = "Primary On-Call"
}

data "pingdom_contact" "secondary_oncall" {
  name = "Secondary On-Call"
}

data "pingdom_contact" "ops_manager" {
  name = "Operations Manager"
}

# Use them in a critical service check
resource "pingdom_check" "critical_service" {
  type       = "http"
  name       = "Critical Service"
  host       = "critical.example.com"
  resolution = 1
  
  userids = [
    data.pingdom_contact.primary_oncall.id,
    data.pingdom_contact.secondary_oncall.id
  ]
  
  teamids = [
    pingdom_team.management.id
  ]
}

# Create a management team
resource "pingdom_team" "management" {
  name = "Management Team"
  
  member_ids = [
    data.pingdom_contact.ops_manager.id
  ]
}
```

## Argument Reference

### Required Arguments

*   `name` (String) - The name of the existing Pingdom contact to retrieve.

## Attributes Reference

*   `id` (String) - The ID of the Pingdom contact.
*   `name` (String) - The name of the contact.

## Use Cases

### Referencing Existing Contacts

Use data sources when you want to reference contacts that already exist in Pingdom:

```hcl
# Reference existing contact
data "pingdom_contact" "existing_contact" {
  name = "Existing Contact Name"
}

# Use in new resources
resource "pingdom_check" "new_check" {
  type       = "http"
  name       = "New Check"
  host       = "example.com"
  resolution = 5
  
  userids = [
    data.pingdom_contact.existing_contact.id
  ]
}
```

### Migrating Existing Infrastructure

When migrating existing Pingdom configurations to Terraform:

```hcl
# Look up existing contacts
data "pingdom_contact" "admin" {
  name = "Admin User"
}

data "pingdom_contact" "ops" {
  name = "Operations Team"
}

# Create new monitoring checks using existing contacts
resource "pingdom_check" "migrated_check" {
  type       = "http"
  name       = "Migrated Check"
  host       = "example.com"
  resolution = 5
  
  userids = [
    data.pingdom_contact.admin.id,
    data.pingdom_contact.ops.id
  ]
}
```

### Hybrid Approach

Combine data sources with new resources:

```hcl
# Use existing contacts
data "pingdom_contact" "existing_admin" {
  name = "Existing Admin"
}

# Create new contacts
resource "pingdom_contact" "new_ops" {
  name = "New Operations Engineer"
  
  email_notification {
    address  = "newops@example.com"
    severity = "HIGH"
  }
}

# Create team with both
resource "pingdom_team" "hybrid_team" {
  name = "Hybrid Team"
  
  member_ids = [
    data.pingdom_contact.existing_admin.id,
    pingdom_contact.new_ops.id
  ]
}
```

## Best Practices

### 1. Contact Naming

- Use exact names that match existing Pingdom contacts
- Ensure contact names are unique and descriptive
- Document the expected contact names in your team

### 2. Error Handling

- Verify contacts exist before referencing them
- Use consistent naming conventions
- Handle cases where contacts might not exist

### 3. Data Source Organization

- Group related data sources together
- Use descriptive variable names
- Document the purpose of each data source

### 4. Migration Strategy

- Start with data sources for existing contacts
- Gradually migrate to managed resources
- Test thoroughly before production deployment

## Troubleshooting

### Common Issues

**Contact Not Found**
- Verify the contact name exactly matches Pingdom
- Check for typos or case sensitivity
- Ensure the contact exists in your Pingdom account

**Data Source Errors**
- Verify your API token has read access
- Check Pingdom API connectivity
- Ensure the contact name is correct

**Import Errors**
- Verify the contact exists before importing
- Check for naming conflicts
- Ensure proper permissions

### Validation

```hcl
# Test data source configuration
data "pingdom_contact" "test_contact" {
  name = "Test Contact"
}

# Output the contact ID for verification
output "contact_id" {
  value = data.pingdom_contact.test_contact.id
}

# Use in a simple check
resource "pingdom_check" "test_check" {
  type       = "http"
  name       = "Test Check"
  host       = "test.example.com"
  resolution = 5
  
  userids = [
    data.pingdom_contact.test_contact.id
  ]
}
```

## Related Resources

- [pingdom_contact](../resources/pingdom_contact.md) - Create and manage contacts
- [pingdom_team](../resources/pingdom_team.md) - Create teams for notifications
- [pingdom_check](../resources/pingdom_check.md) - Create monitoring checks
- [Team Management Guide](../guides/team_management.md) - Comprehensive team management
- [Getting Started Guide](../guides/getting-started.md) - Basic setup and configuration

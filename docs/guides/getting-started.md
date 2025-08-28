---
page_title: "Getting Started with Pingdom Provider"
description: "Learn how to set up and configure the Pingdom provider for Terraform"
---

# Getting Started with Pingdom Provider

This guide will walk you through setting up the Pingdom provider for Terraform and creating your first monitoring check.

## Prerequisites

- [Terraform](https://www.terraform.io/downloads.html) 1.0 or later installed
- A Pingdom account with API access
- A Pingdom API token

## Step 1: Get Your Pingdom API Token

1. Log in to your [Pingdom account](https://my.pingdom.com/)
2. Navigate to **Settings** → **Integrations** → **API**
3. Generate a new API token or copy an existing one
4. Save the token securely - you'll need it for the provider configuration

## Step 2: Install the Provider

Add the Pingdom provider to your Terraform configuration:

```hcl
terraform {
  required_providers {
    pingdom = {
      source  = "AdconnectDevOps/pingdom"
      version = "~> 1.0"
    }
  }
}
```

## Step 3: Configure the Provider

Create a provider configuration block with your API token:

```hcl
provider "pingdom" {
  api_token = var.pingdom_api_token
}

# Variables
variable "pingdom_api_token" {
  description = "Pingdom API token for authentication"
  type        = string
  sensitive   = true
}
```

## Step 4: Set Your API Token

Choose one of these methods to provide your API token:

### Method 1: Environment Variable
```bash
export PINGDOM_API_TOKEN="your_api_token_here"
```

### Method 2: Terraform Variable File
Create a `terraform.tfvars` file:
```hcl
pingdom_api_token = "your_api_token_here"
```

### Method 3: Command Line
```bash
terraform apply -var="pingdom_api_token=your_api_token_here"
```

## Step 5: Create Your First Check

Create a simple HTTP check to monitor your website:

```hcl
resource "pingdom_check" "website" {
  type       = "http"
  name       = "My Website"
  host       = "example.com"
  resolution = 5
  url        = "/"
  encryption = true  # Use HTTPS
}
```

## Step 6: Initialize and Apply

```bash
# Initialize Terraform
terraform init

# Plan your changes
terraform plan

# Apply the configuration
terraform apply
```

## Step 7: Verify the Check

1. Go to your [Pingdom dashboard](https://my.pingdom.com/)
2. You should see your new check listed
3. The check will start monitoring immediately

## Next Steps

Now that you have a basic check running, you can:

- [Add alert notifications](team_management.md) by creating contacts and teams
- [Configure advanced HTTP monitoring](http_monitoring.md) with custom headers and authentication
- [Set up ping checks](ping_monitoring.md) for network monitoring
- [Create TCP port checks](tcp_monitoring.md) for service monitoring

## Troubleshooting

### Common Issues

**Authentication Error (401)**
- Verify your API token is correct
- Ensure the token has the necessary permissions

**Check Creation Fails**
- Verify the hostname is accessible
- Check that the URL path exists
- Ensure the check type is supported

**Rate Limiting (429)**
- Reduce the frequency of API calls
- Implement exponential backoff in your automation

## Support

If you encounter issues:

1. Check the [Terraform Registry documentation](https://registry.terraform.io/providers/AdconnectDevOps/pingdom)
2. Review the [GitHub repository](https://github.com/AdconnectDevOps/terraform-provider-pingdom)
3. Open an [issue](https://github.com/AdconnectDevOps/terraform-provider-pingdom/issues) for bugs
4. Start a [discussion](https://github.com/AdconnectDevOps/terraform-provider-pingdom/discussions) for questions

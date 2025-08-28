# Terraform Provider for Pingdom

[![GitHub release](https://img.shields.io/github/release/russellcardullo/terraform-provider-pingdom.svg)](https://github.com/russellcardullo/terraform-provider-pingdom/releases)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

This is a [Terraform](https://www.terraform.io/) provider for [Pingdom](https://www.pingdom.com/), allowing you to manage your Pingdom monitoring checks, teams, and contacts as code.

## Features

- **HTTP Checks**: Monitor HTTP/HTTPS endpoints with custom headers, authentication, and response validation
- **Ping Checks**: Basic ICMP ping monitoring for network availability
- **Team Management**: Create and manage teams for alert notifications
- **Contact Management**: Configure SMS and email notifications with severity levels
- **Integration Support**: Webhook integrations for custom alert handling

## Requirements

- **Terraform**: >= 1.0
- **Go**: 1.22+ (for building from source)
- **Pingdom API**: v3.1 access token

## Installation

### From Terraform Registry (Recommended)

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

### From Source

```bash
git clone https://github.com/russellcardullo/terraform-provider-pingdom.git
cd terraform-provider-pingdom
make build
make install
```

## Quick Start

### 1. Configure the Provider

```hcl
terraform {
  required_providers {
    pingdom = {
      source  = "russellcardullo/pingdom"
      version = "~> 1.0"
    }
  }
}

# Configure the Pingdom Provider
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

### 2. Set Your API Token

```bash
# Option 1: Environment variable
export PINGDOM_API_TOKEN="your_api_token_here"

# Option 2: Terraform variable file (terraform.tfvars)
echo 'pingdom_api_token = "your_api_token_here"' > terraform.tfvars

# Option 3: Command line
terraform apply -var="pingdom_api_token=your_api_token_here"
```
```

## Examples

### Basic HTTP Check

```hcl
resource "pingdom_check" "website" {
  type       = "http"
  name       = "Website Homepage"
  host       = "example.com"
  resolution = 5
  url        = "/"
  encryption = true  # Use HTTPS
}
```

### Advanced HTTP Check with Alerts

```hcl
resource "pingdom_check" "api_endpoint" {
  type                        = "http"
  name                        = "API Health Check"
  host                        = "api.example.com"
  resolution                  = 1
  url                         = "/health"
  encryption                  = true
  port                        = 443
  shouldcontain              = "healthy"
  responsetime_threshold     = 5000  # 5 seconds
  
  # Alert configuration
  sendnotificationwhendown   = 2     # Alert after 2 consecutive failures
  notifyagainevery           = 5     # Re-notify every 5 minutes
  notifywhenbackup           = true
  
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
    "User-Agent" = "Terraform-Pingdom-Provider"
    "X-Custom"   = "value"
  }
  
  # Tags for organization
  tags = "production,api,health"
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
    pingdom_contact.network_admin.id
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

## Terraform Registry

This provider is available on the [Terraform Registry](https://registry.terraform.io/providers/AdconnectDevOps/pingdom) and can be installed automatically by Terraform.

### Registry Usage

```hcl
terraform {
  required_providers {
    pingdom = {
      source  = "russellcardullo/pingdom"
      version = "~> 1.0"
    }
  }
}

provider "pingdom" {
  api_token = var.pingdom_api_token
}
```

### Version Constraints

- `~> 1.0` - Use any 1.x version (recommended for production)
- `~> 1.1` - Use any 1.1.x version
- `>= 1.0, < 2.0` - Use any 1.x version with explicit upper bound

### Apply Your Configuration

```bash
# Initialize Terraform
terraform init

# Plan your changes
terraform plan

# Apply the configuration
terraform apply

# Or apply with variables
terraform apply -var="pingdom_api_token=YOUR_API_TOKEN"
```

**Using attributes from other resources**

```hcl
variable "heroku_email" {}
variable "heroku_api_key" {}

variable "pingdom_api_token" {}

provider "heroku" {
    email = var.heroku_email
    api_key = var.heroku_api_key
}

provider "pingdom" {
    api_token = var.pingdom_api_token
}

resource "heroku_app" "example" {
    name = "my-app"
    region = "us"
}

resource "pingdom_check" "example" {
    name = "my check"
    host = heroku_app.example.heroku_hostname
    resolution = 5
}
```

**Teams**

```hcl
resource "pingdom_team" "test" {
  name = "The Test team"
  member_ids = [
    pingdom_contact.first_contact.id,
  ]
}
```

**Contacts**

Note that all contacts _must_ have both a high and low severity notification

```hcl

resource "pingdom_contact" "first_contact" {
  name = "johndoe"

  sms_notification {
    number   = "5555555555"
    severity = "HIGH"
  }

  sms_notification {
    number       = "3333333333"
    country_code = "91"
    severity     = "LOW"
    provider     = "esendex"
  }

  email_notification {
    address  = "test@test.com"
    severity = "LOW"
  }
}

resource "pingdom_contact" "second_contact" {
  name   = "janedoe"
  paused = true

  email_notification {
    address  = "test@test.com"
    severity = "LOW"
  }

  email_notification {
    address  = "test@test.com"
    severity = "HIGH"
  }
}
```

## Troubleshooting

### Common Issues

#### Authentication Errors
```bash
Error: Error creating Pingdom check: 401 Unauthorized
```
**Solution**: Verify your API token is correct and has the necessary permissions.

#### Rate Limiting
```bash
Error: 429 Too Many Requests
```
**Solution**: Implement exponential backoff or reduce the frequency of API calls.

#### Invalid Check Configuration
```bash
Error: Invalid check type specified
```
**Solution**: Ensure check type is one of: `http`, `https`, `ping`, `tcp`.

### Getting Help

- **GitHub Issues**: [Report bugs or request features](https://github.com/russellcardullo/terraform-provider-pingdom/issues)
- **Discussions**: [Community discussions](https://github.com/russellcardullo/terraform-provider-pingdom/discussions)
- **Documentation**: [Pingdom API Reference](https://docs.pingdom.com/api/)

## Resource Reference

### pingdom_check

#### Common Attributes

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `name` | string | Yes | The name of the check |
| `host` | string | Yes | The hostname to check (e.g., `example.com`) |
| `resolution` | number | Yes | Check frequency in minutes (1, 5, 15, 30, 60) |
| `type` | string | Yes | Check type (`http`, `https`, `ping`, `tcp`) |
| `paused` | bool | No | Whether the check is active (default: `false`) |
| `responsetime_threshold` | number | No | Response time threshold in milliseconds (default: 30000) |
| `sendnotificationwhendown` | number | No | Consecutive failures before alert (default: 1) |
| `notifyagainevery` | number | No | Re-notify interval in minutes (default: 0) |
| `notifywhenbackup` | bool | No | Notify when service recovers (default: `false`) |
| `integrationids` | list | No | List of webhook integration IDs |
| `userids` | list | No | List of user IDs for notifications |
| `teamids` | list | No | List of team IDs for notifications |

> **Note**: When using `integrationids`, the `sendnotificationwhendown` value will be ignored when sending webhook notifications. You may need to contact Pingdom support for more details. See [#52](https://github.com/russellcardullo/terraform-provider-pingdom/issues/52).

#### HTTP specific attributes ####

For the HTTP checks, you can set these attributes:

  * **url** - Target path on server.

  * **encryption** - Enable encryption in the HTTP check (aka HTTPS).

  * **port** - Target port for HTTP checks.

  * **username** - Username for target HTTP authentication.

  * **password** - Password for target HTTP authentication.

  * **shouldcontain** - Target site should contain this string.

  * **shouldnotcontain** - Target site should NOT contain this string. Not allowed defined together with `shouldcontain`.

  * **postdata** - Data that should be posted to the web page, for example submission data for a sign-up or login form. The data needs to be formatted in the same way as a web browser would send it to the web server.

  * **requestheaders** - Custom HTTP headers. It should be a hash with pairs, like `{ "header_name" = "header_content" }`

  * **tags** - List of tags the check should contain. Should be in the format "tagA,tagB"

  * **probefilters** - Region from which the check should originate. One of NA, EU, APAC, or LATAM. Should be in the format "region:NA"

#### TCP specific attributes ####

For the TCP checks, you can set these attributes:

  * **port** - Target port for TCP checks.

  * **stringtosend** - (optional) This string will be sent to the port

  * **stringtoexpect** - (optional) This string must be returned by the remote host for the check to pass

The following attributes are exported:

  * **id** The ID of the Pingdom check


### Pingdom Team ###

  * **name** - (Required) The name of the team

  * **member_ids** - List of integer contact IDs that will be notified when the check is down.


### Pingdom Contact ###

  * **name**: (Required) Name of the contact

  * **paused**: Whether alerts for this contact should be disabled

  * **sms_notification**: Block resource describing an SMS notification

      * **country_code**: The country code, defaults to "1"

      * **number**: The phone number

      * **provider**: Provider for SMS messaging. One of nexmo|bulksms|esendex|cellsynt. 'bulksms' not presently operational

      * **severity**: Severity of this notification. One of HIGH|LOW

  * **email_notification**: Block resource describing an Email notification

      * **address**: Email address to notify

      * **severity**: Severity of this notification. One of HIGH|LOW

## Documentation

### API Reference

This provider supports Pingdom API v3.1. For detailed API documentation, see the [Pingdom API Reference](https://docs.pingdom.com/api/).

### Resource Types

- **pingdom_check** - Monitor HTTP, HTTPS, ping, and TCP endpoints
- **pingdom_team** - Manage teams for alert notifications
- **pingdom_contact** - Configure SMS and email notification contacts
- **pingdom_integration** - Set up webhook integrations

### Data Sources

- **pingdom_contact** - Retrieve existing contact information
- **pingdom_team** - Retrieve existing team information

## Development

### Prerequisites

- **Go**: 1.22+ (see [Go installation guide](https://golang.org/doc/install))
- **Make**: For build automation
- **Git**: For version control

### Building from Source

```bash
# Clone the repository
git clone https://github.com/russellcardullo/terraform-provider-pingdom.git
cd terraform-provider-pingdom

# Install dependencies
go mod download

# Build the provider
make build

# Install locally
make install
```

### Development Workflow

```bash
# Run tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Check code formatting
go fmt ./...

# Run linter
go vet ./...

# Build and test
make test
make build
```

### Project Structure

```
terraform-provider-pingdom/
├── pingdom/                 # Provider implementation
│   ├── config.go           # Provider configuration
│   ├── provider.go         # Provider schema and resources
│   ├── resource_*.go       # Resource implementations
│   └── data_source_*.go    # Data source implementations
├── examples/                # Usage examples
├── .github/                 # GitHub Actions workflows
├── go.mod                   # Go module definition
└── README.md               # This file
```

### Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Add tests for new functionality
5. Ensure all tests pass (`go test ./...`)
6. Commit your changes (`git commit -m 'Add amazing feature'`)
7. Push to the branch (`git push origin feature/amazing-feature`)
8. Open a Pull Request

### Testing

```bash
# Run all tests
go test ./...

# Run specific test
go test ./pingdom -v

# Run tests with race detection
go test -race ./...
```

### Release Process

1. Update version in `go.mod`
2. Create and push a git tag
3. GitHub Actions will automatically build and release
4. Update Terraform Registry documentation

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

- **Documentation**: [Terraform Registry](https://registry.terraform.io/providers/russellcardullo/pingdom)
- **Issues**: [GitHub Issues](https://github.com/russellcardullo/terraform-provider-pingdom/issues)
- **Discussions**: [GitHub Discussions](https://github.com/russellcardullo/terraform-provider-pingdom/discussions)

## Acknowledgments

- [Pingdom](https://www.pingdom.com/) for providing the monitoring API
- [HashiCorp](https://www.hashicorp.com/) for the Terraform framework
- All contributors who have helped improve this provider

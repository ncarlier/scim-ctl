# SCIM CTL

`scim-ctl` is a CLI tool for interacting with a SCIM (System for Cross-domain Identity Management) server. It supports CRUD operations.

# Installation

Source code and releases are available at https://github.com/ncarlier/scim-ctl/releases

```bash
curl -fsSL https://raw.githubusercontent.com/ncarlier/scim-ctl/main/install.sh | bash
```

# Features

## SCIM Support

- Compatible with SCIM 2.0 (RFC 7644)
- Manages resource types: `Users`, `Groups`, and can be extended to custom types.

## SCIM Target

The SCIM target is the URL of an HTTP server.

**Configuration:**

| Parameter   | Env                | Description   |
| ----------- | ------------------ | ------------- |
| `--target`  | `SCIM_CTL_TARGET`  | SCIM Target   |

**Usage:**

```bash
scim-ctl --target http://example.com/scim/v2 schemas
```

## Authentication

The CLI authenticates to the SCIM server using a **Bearer Token** of type **JWT OpenID Connect**.
It obtains the Access Token from an OIDC provider using the OAuth 2.0 Device Authorization Grant.

**Configuration:**

| Parameter               | Env                            | Description           |
| ----------------------- | ------------------------------ | --------------------- |
| `--oidc-issuer`         | `SCIM_CTL_OIDC_ISSUER`         | OpenID Connect Issuer |
| `--oidc-client-id`      | `SCIM_CTL_OIDC_CLIENT_ID`      | Client ID             |
| `--oidc-client-secret`  | `SCIM_CTL_OIDC_CLIENT_SECRET`  | Client Secret         |
| `--oidc-grant-type`     | `SCIM_CTL_OIDC_GRANT_TYPE`     | Grant Type            |
| `--extra-header`        | -                              | Extra HTTP headers (format: `key=value`, can be specified multiple times) |

Supported grant type are `client_credentials` or `device_code` (default).

## Configuration File

In addition to environment variables and command-line flags, `scim-ctl` can be configured using a YAML configuration file (`scim-ctl.yml` by default).

Example:

```yaml
target: "https://example.com/scim/v2"
oidc:
  issuer: "https://auth.example.com"
  client-id: "my-client"
  client-secret: "my-secret"
extra-headers:
  X-Custom-Header: "Some-Value"
```

## Display

Print HTTP request and response for debugging.

| Parameter   | Alias | Description                 |
| ----------- | ----- | --------------------------- |
| `--verbose` | `-v`  | Increase verbosity          |

## SCIM Commands

The user can execute the following commands:

### Schemas (`schemas`)

Display the resources and attribute extensions supported by the server.

```bash
scim-ctl schemas
```

### Create (`create`)

Create a SCIM resource.

```bash
scim-ctl create --resource user --data '{"userName": "jdoe", ...}'
```

| Parameter         | Alias | Description                 |
| ----------------- | ----- | --------------------------- |
| `--resource`      | `-r`  | SCIM resource type          |
| `--data`          | `-d`  | SCIM resource payload       |

Data can also be provided via STDIN:

```bash
cat user.json | scim-ctl create -r user
```

### Read (`get`)

Retrieve a resource by ID.

```bash
scim-ctl get --resource user --id 1234
```

| Parameter         | Alias | Description                     |
| ----------------- | ----- | ------------------------------- |
| `--id`            | n/a   | SCIM resource identifier        |
| `--resource`      | `-r`  | SCIM resource type              |

### Replace (`replace`)

Replace an existing resource.

```bash
scim-ctl replace --resource user --id 1234 --data '{"userName": "johndoe"}'
```

| Parameter         | Alias | Description                     |
| ----------------- | ----- | ------------------------------- |
| `--id`            | n/a   | SCIM resource identifier        |
| `--resource`      | `-r`  | SCIM resource type              |
| `--data`          | `-d`  | SCIM resource payload           |

### Update (`update`)

Update an existing resource using a partial update (PATCH).

```bash
scim-ctl update --resource user --id 1234 --data '[{"op":"replace","path":"userName","value":"johndoe"}]'
```

| Parameter         | Alias | Description                     |
| ----------------- | ----- | ------------------------------- |
| `--id`            | n/a   | SCIM resource identifier        |
| `--resource`      | `-r`  | SCIM resource type              |
| `--data`          | `-d`  | SCIM operations payload (JSON array) |

### Delete (`delete`)

Delete a SCIM resource.

```bash
scim-ctl delete --resource user --id 1234
```

| Parameter         | Alias | Description                     |
| ----------------- | ----- | ------------------------------- |
| `--id`            | n/a   | SCIM resource identifier        |
| `--resource`      | `-r`  | SCIM resource type              |

### Search (`search`)

Search SCIM resources.

```bash
scim-ctl search --resource user --filter 'userName eq "bob"'
scim-ctl search -r user -q "john doe"
```

| Parameter          | Alias | Description                                    |
| ------------------ | ----- | ---------------------------------------------- |
| `--resource`       | `-r`  | SCIM resource type                             |
| `--filter`         | `-f`  | SCIM filter expression                         |
| `--query`          | `-q`  | Full-text search query (out of SCIM spec)      |
| `--start-index`    | `-s`  | Paginations start index                        |
| `--items-per-page` | `-i`  | Paginations size                               |

### Export (`export`)

Export all SCIM resources as JSON Lines text format. Pagination is handled automatically.

```bash
scim-ctl export --resource user --filter 'userName eq "bob"'
scim-ctl export -r group -f 'displayName co "admin"' --items-per-page 100
```

| Parameter          | Alias | Description                                    |
| ------------------ | ----- | ---------------------------------------------- |
| `--resource`       | `-r`  | SCIM resource type                             |
| `--filter`         | `-f`  | SCIM filter expression                         |
| `--query`          | `-q`  | Full-text search query (out of SCIM spec)      |
| `--items-per-page` | `-i`  | Paginations size                               |

### Import (`import`)

Import SCIM resources using the Bulk API. The input data should be a stream of JSON Lines, where each line is the payload of a resource to create.

```bash
scim-ctl import --resource user --file users.jsonl
scim-ctl import -r user -f users.jsonl --chunk 50
```

| Parameter         | Alias | Description                     |
| ----------------- | ----- | ------------------------------- |
| `--resource`      | `-r`  | SCIM resource type (required)   |
| `--file`          | `-f`  | Input JSON Lines file path      |
| `--chunk`         | n/a   | Chunk size for bulk requests (default: 100) |

Data can also be provided via STDIN:

```bash
cat users.jsonl | scim-ctl import -r user --chunk 100
```

## Examples and Usage Guides

The `examples/` directory contains comprehensive SCIM JSON examples and detailed usage instructions:

- **[examples/README.md](examples/README.md)** - Complete usage guide with CLI command examples
- **examples/users/** - User resource creation and management examples  
- **examples/groups/** - Group resource examples with member management
- **examples/replaces/** - Full resource replace examples (PUT operations)
- **examples/patches/** - PATCH operation examples for fine-grained replaces
- **examples/filters/** - Search filter patterns and examples

### Quick Start with Examples

```bash
# Create a user from example template
scim-ctl create -r user -d @examples/users/basic-user.json

# Search for active users  
scim-ctl search -r user -f 'active eq true'

# Add email to user using PATCH
scim-ctl update -r user --id USER_ID -d @examples/patches/user-add-email.json
```

For detailed usage instructions and more examples, see the [Examples README](examples/README.md).


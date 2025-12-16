# SCIM CTL

`scim-ctl` is a CLI tool for interacting with a SCIM (System for Cross-domain Identity Management) server. It supports CRUD operations.

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
scim-ctl create --resource-type user --data '{"userName": "jdoe", ...}'
```

| Parameter         | Alias | Description                 |
| ----------------- | ----- | --------------------------- |
| `--resource-type` | `-t`  | SCIM resource type          |
| `--data`          | `-d`  | SCIM resource payload       |

Data can also be provided via STDIN:

```bash
cat user.json | scim-ctl create -t user
```

### Read (`get`)

Retrieve a resource by ID.

```bash
scim-ctl get --resource-type user --id 1234
```

| Parameter         | Alias | Description                     |
| ----------------- | ----- | ------------------------------- |
| `--id`            | n/a   | SCIM resource identifier        |
| `--resource-type` | `-t`  | SCIM resource type              |

### Update (`update`)

Update an existing resource.

```bash
scim-ctl update --resource-type user --id 1234 --data '{"userName": "johndoe"}'
```

| Parameter         | Alias | Description                     |
| ----------------- | ----- | ------------------------------- |
| `--id`            | n/a   | SCIM resource identifier        |
| `--resource-type` | `-t`  | SCIM resource type              |
| `--data`          | `-d`  | SCIM resource payload           |

### Delete (`delete`)

Delete a SCIM resource.

```bash
scim-ctl delete --resource-type user --id 1234
```

| Parameter         | Alias | Description                     |
| ----------------- | ----- | ------------------------------- |
| `--id`            | n/a   | SCIM resource identifier        |
| `--resource-type` | `-t`  | SCIM resource type              |

### Search (`search`)

Search SCIM resources.

```bash
scim-ctl search --resource-type user --filter 'userName eq "bob"'
```

| Parameter          | Alias | Description                     |
| ------------------ | ----- | ------------------------------- |
| `--resource-type`  | `-t`  | SCIM resource type              |
| `--filter`         | `-f`  | SCIM filter expression          |
| `--start-index`    | `-s`  | Paginations start index         |
| `--items-per-page` | `-i`  | Paginations size                |

## Examples and Usage Guides

The `examples/` directory contains comprehensive SCIM JSON examples and detailed usage instructions:

- **[examples/README.md](examples/README.md)** - Complete usage guide with CLI command examples
- **examples/users/** - User resource creation and management examples  
- **examples/groups/** - Group resource examples with member management
- **examples/updates/** - Full resource update examples (PUT operations)
- **examples/patches/** - PATCH operation examples for fine-grained updates
- **examples/filters/** - Search filter patterns and examples

### Quick Start with Examples

```bash
# Create a user from example template
scim-ctl create -t user -d @examples/users/basic-user.json

# Search for active users  
scim-ctl search -t user -f 'active eq true'

# Add email to user using PATCH
scim-ctl patch -t user --id USER_ID -d @examples/patches/user-add-email.json
```

For detailed usage instructions and more examples, see the [Examples README](examples/README.md).


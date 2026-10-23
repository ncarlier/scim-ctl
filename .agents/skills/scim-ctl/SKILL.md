---
name: scim-ctl
description: Interact with SCIM 2.0 (System for Cross-domain Identity Management) services using the scim-ctl CLI. Enables discovering supported SCIM resources, endpoints, and attribute definitions via schemas and resource-types, querying resources with RFC 7644 filters, retrieving, creating, patching, replacing, and deleting resources (Users, Groups, custom types), and performing bulk export/import operations. Always use this skill when working with SCIM servers, scim-ctl, identity schemas, or directory users and groups.
---

# SCIM Control (`scim-ctl`) Skill

This skill provides comprehensive instructions for managing identity resources across SCIM 2.0 compliant servers using the `scim-ctl` CLI.

---

## 1. Discovery-First Workflow

Never assume what resource types or schema attributes a target SCIM server supports. Different identity providers (Okta, Entra ID, Ping, custom servers) provide different extensions and attributes.

Always begin by discovering the target's capabilities:

### Step 1: Discover Supported Resource Types
List the resource types configured on the server:
```bash
scim-ctl resource-types
```
Key information in the JSON response:
- `name`: Resource type name (e.g., `"User"`, `"Group"`). When passing to CLI `-r` or `--resource`, use lowercase singular (e.g., `-r user`, `-r group`).
- `endpoint`: The REST endpoint path (e.g., `"/Users"`, `"/Groups"`).
- `schema`: The main schema URN (e.g., `"urn:ietf:params:scim:schemas:core:2.0:User"`).
- `schemaExtensions`: Array of optional or required schema extensions (e.g., EnterpriseUser URN).

### Step 2: Inspect Schemas and Attributes
Retrieve attribute definitions, types, and constraints across all schemas:
```bash
scim-ctl schemas
```
Inspect each schema:
- Match `id` to the resource type's `schema` URN or `schemaExtensions` URN.
- Look at `attributes` to see:
  - `name`: exact case-sensitive attribute name (e.g. `userName`, `name`, `emails`).
  - `type`: `string`, `boolean`, `complex`, `dateTime`, etc.
  - `multiValued`: whether it is an array of objects/strings.
  - `required`: mandatory fields when creating resources (`create`).
  - `subAttributes`: attributes of complex types (e.g. `emails.value`, `emails.type`, `name.givenName`).

---

## 2. Configuration & Authentication

`scim-ctl` reads configuration from CLI flags, environment variables, or a YAML file (`./scim-ctl.yml` by default or via `--config`).

### Environment Variables
| Variable | Description |
| :--- | :--- |
| `SCIM_CTL_TARGET` | Base SCIM server URL (e.g., `https://example.com/scim/v2`) |
| `SCIM_CTL_CONFIG` | Path to custom YAML configuration file |
| `SCIM_CTL_OIDC_ISSUER` | OpenID Connect issuer URL |
| `SCIM_CTL_OIDC_CLIENT_ID` | OAuth 2.0 client ID |
| `SCIM_CTL_OIDC_CLIENT_SECRET` | OAuth 2.0 client secret |
| `SCIM_CTL_OIDC_GRANT_TYPE` | `client_credentials` or `device_code` (default) |
| `SCIM_CTL_CACHE_DIR` | Directory to cache access tokens |
| `SCIM_CTL_TIMEOUT` | HTTP request timeout in seconds (default: 30) |
| `SCIM_CTL_MAX_RETRIES` | Max request retries (default: 3) |

### Global Flags
- `--target <url>`: Override the target server URL.
- `--extra-header <key=value>`: Add custom HTTP header (can repeat, e.g. `--extra-header "Authorization=Bearer <token>"`).
- `-v`, `--verbose`: Show detailed HTTP requests and responses.

### Managing Token Cache
```bash
# Check cached token details
scim-ctl cache info

# Clear cached tokens if expired or switching identity context
scim-ctl cache clear
```

---

## 3. Querying & Reading Resources

### Search Resources (`search`)
Search for resources using SCIM filter expressions:
```bash
# Filter by exact attribute match
scim-ctl search -r user -f 'userName eq "john.doe"'

# Combine filter conditions
scim-ctl search -r user -f 'active eq true and (userType eq "Employee" or title pr)'

# Limit returned attributes for speed and clarity
scim-ctl search -r user -a userName,emails,active

# Pagination and sorting
scim-ctl search -r user -f 'active eq true' --start-index 1 --items-per-page 25 --sort-by meta.created --sort-order descending

# Cursor-based pagination (RFC 9865)
scim-ctl search -r user --use-cursor
```

> For full filter syntax (`eq`, `ne`, `co`, `sw`, `ew`, `pr`, `gt`, complex sub-attributes), refer to [references/scim-filters.md](references/scim-filters.md).

### Get Resource by ID (`get`)
Retrieve a specific resource:
```bash
scim-ctl get -r user --id <resource-id>
scim-ctl get -r group --id <group-id> -a displayName,members
```

### Get Authenticated User (`me`)
Retrieve the SCIM profile for the currently authenticated identity:
```bash
scim-ctl me
scim-ctl me -a userName,emails
```

---

## 4. Mutating Resources (CRUD)

### Create Resource (`create`)
Create a new resource. The payload must satisfy the schema's required attributes:
```bash
# Inline JSON payload
scim-ctl create -r user -d '{
  "schemas": ["urn:ietf:params:scim:schemas:core:2.0:User"],
  "userName": "alice",
  "name": {"givenName": "Alice", "familyName": "Smith"},
  "emails": [{"value": "alice@example.com", "primary": true}],
  "active": true
}'

# Pipe from file or stdin
cat new-user.json | scim-ctl create -r user
```

### Partial Update (`update`)
Perform partial updates using SCIM PATCH operations. Payload must be a **JSON array** of operations:
```bash
# Replace single attribute
scim-ctl update -r user --id <id> -d '[{"op":"replace","path":"active","value":false}]'

# Add an email
scim-ctl update -r user --id <id> -d '[{"op":"add","path":"emails","value":[{"value":"alice.work@example.com","type":"work"}]}]'

# Add member to a group
scim-ctl update -r group --id <group-id> -d '[{"op":"add","path":"members","value":[{"value":"<user-id>"}]}]'
```

> For comprehensive PATCH examples (`add`, `replace`, `remove`, member modifications), refer to [references/patch-operations.md](references/patch-operations.md).

### Replace Resource (`replace`)
Perform a full replacement (HTTP PUT). Must include the complete resource payload:
```bash
scim-ctl replace -r user --id <id> -d '{
  "schemas": ["urn:ietf:params:scim:schemas:core:2.0:User"],
  "userName": "alice",
  "displayName": "Alice Smith",
  "active": true
}'
```

### Delete Resource (`delete`)
Delete a resource by ID:
```bash
scim-ctl delete -r user --id <id>
scim-ctl delete -r group --id <id>
```

---

## 5. Bulk & Batch Operations

For exporting datasets or performing bulk modifications:

### Exporting (`export`)
Exports all matching resources as JSON Lines (JSONL), handling pagination automatically:
```bash
# Export all active users to a file
scim-ctl export -r user -f 'active eq true' > active_users.jsonl

# Export groups with chunking
scim-ctl export -r group --items-per-page 100 > groups.jsonl
```

### Bulk Import (`import`)
Create multiple resources from a JSON Lines stream using the SCIM Bulk API:
```bash
scim-ctl import -r user -f users.jsonl --chunk 50
cat users.jsonl | scim-ctl import -r user --chunk 100
```

### Bulk Update (`bulk-update`)
Apply PATCH updates from a JSON Lines stream where each line contains `{"id": "...", "operations": [...]}`:
```bash
scim-ctl bulk-update -r user -f updates.jsonl --chunk 50
```

### Bulk Delete (`bulk-delete`)
Delete multiple resources from a JSON Lines stream where each line contains `{"id": "..."}`:
```bash
scim-ctl bulk-delete -r user -f deletes.jsonl --chunk 50
```

---

## 6. Troubleshooting & Best Practices

1. **Always single-quote filters in shell**:
   Double quotes allow shell expansions or syntax errors on brackets/quotes:
   ```bash
   scim-ctl search -r user -f 'emails[type eq "work"]'
   ```
2. **Schema validation errors**:
   If a `create` or `replace` call returns 400 Bad Request, run `scim-ctl schemas` to check required attributes and enum values (e.g. `emails[].type`).
3. **Authentication expiration**:
   If requests fail with 401 or token errors, run `scim-ctl cache clear` to force a new authentication exchange.
4. **Target URL path**:
   Verify whether `SCIM_CTL_TARGET` includes the API version prefix (e.g. `https://host/scim/v2` vs `https://host/v2`). The tool appends `/<ResourceName>` (e.g. `/Users`, `/Groups`, `/Schemas`, `/ResourceTypes`).

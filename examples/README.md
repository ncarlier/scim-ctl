# SCIM CLI Examples

This directory contains comprehensive examples for SCIM (System for Cross-domain Identity Management) operations using the `scim-ctl` CLI tool.

## Directory Structure

```
examples/
├── users/           # User resource examples
├── groups/          # Group resource examples  
├── updates/         # Full resource update examples
├── patches/         # PATCH operation examples
├── filters/         # Search filter examples
└── README.md        # This file
```

## Prerequisites

Before using these examples, ensure you have:

1. **Configured authentication** - Set up your OAuth 2.0 configuration
2. **Valid SCIM endpoint** - Access to a SCIM 2.0 compliant server
3. **scim-ctl installed** - The CLI tool built and available

## Configuration

Set up your environment variables or configuration file:

```bash
export SCIM_SERVER_URL="https://your-scim-server.com/v2"
export OAUTH_ISSUER="https://your-oauth-provider.com"
export OAUTH_CLIENT_ID="your-client-id"
export OAUTH_SCOPES="scim:read scim:write"
```

Or use the interactive bash script:
```bash
./scim.sh config
```

## User Examples

### Creating Users

#### Basic User Creation
```bash
# Create a simple user with minimal required fields
scim-ctl users create -f examples/users/basic-user.json

# Using the bash script
./scim.sh create-user examples/users/basic-user.json
```

#### Complete User with All Fields
```bash
# Create a user with comprehensive profile information
scim-ctl users create -f examples/users/complete-user.json
```

#### Enterprise User
```bash
# Create a user with enterprise extension attributes
scim-ctl users create -f examples/users/enterprise-user.json
```

#### Contractor User
```bash
# Create a contractor user with specific attributes
scim-ctl users create -f examples/users/contractor-user.json
```

### User Updates

#### Full User Update
```bash
# Replace entire user resource (PUT operation)
scim-ctl users update USER_ID -f examples/updates/user-full-update.json

# Using bash script
./scim.sh update-user USER_ID examples/updates/user-full-update.json
```

#### Partial User Update  
```bash
# Update only specified fields
scim-ctl users update USER_ID -f examples/updates/user-partial-update.json
```

#### User Deactivation
```bash
# Deactivate a user account
scim-ctl users update USER_ID -f examples/updates/user-deactivate.json
```

### User Patches (Fine-grained Updates)

#### Add Email Address
```bash
# Add a new email to user's profile
scim-ctl users patch USER_ID -f examples/patches/user-add-email.json

# Using bash script  
./scim.sh patch-user USER_ID examples/patches/user-add-email.json
```

#### Update Name
```bash
# Change user's given and family names
scim-ctl users patch USER_ID -f examples/patches/user-update-name.json
```

#### Remove Email
```bash  
# Remove specific email address using filter
scim-ctl users patch USER_ID -f examples/patches/user-remove-email.json
```

#### Update Contact Information
```bash
# Modify existing email and phone number
scim-ctl users patch USER_ID -f examples/patches/user-update-contact.json
```

#### Enterprise Attributes Update
```bash
# Update enterprise-specific attributes (department, manager, etc.)
scim-ctl users patch USER_ID -f examples/patches/user-enterprise-update.json
```

#### Complex Multi-Operation Patch
```bash
# Perform multiple operations in single request
scim-ctl users patch USER_ID -f examples/patches/user-complex-update.json
```

#### Deactivate User
```bash
# Deactivate user using PATCH operation
scim-ctl users patch USER_ID -f examples/patches/user-deactivate.json
```

## Group Examples

### Creating Groups

#### Basic Group
```bash
# Create simple group with minimal fields
scim-ctl groups create -f examples/groups/basic-group.json

# Using bash script
./scim.sh create-group examples/groups/basic-group.json
```

#### Team Group with Members
```bash
# Create group with initial members
scim-ctl groups create -f examples/groups/team-group.json
```

#### Empty Group
```bash
# Create group without members (to be added later)
scim-ctl groups create -f examples/groups/empty-group.json
```

### Group Updates

#### Full Group Update
```bash
# Replace entire group resource
scim-ctl groups update GROUP_ID -f examples/updates/group-update.json

# Using bash script
./scim.sh update-group GROUP_ID examples/updates/group-update.json
```

### Group Patches

#### Add Members to Group
```bash
# Add multiple members to existing group
scim-ctl groups patch GROUP_ID -f examples/patches/group-add-members.json

# Using bash script
./scim.sh patch-group GROUP_ID examples/patches/group-add-members.json
```

#### Remove Member from Group
```bash
# Remove specific user from group
scim-ctl groups patch GROUP_ID -f examples/patches/group-remove-member.json
```

#### Update Group Information
```bash
# Change group display name and external ID
scim-ctl groups patch GROUP_ID -f examples/patches/group-update-info.json
```

## Search and Filter Examples

### User Searches

#### Basic User Filters
```bash
# Search by username
scim-ctl users list --filter 'userName eq "john.doe@example.com"'

# Search by email  
scim-ctl users list --filter 'emails[value eq "john@example.com"]'

# Search active users
scim-ctl users list --filter 'active eq true'

# Using bash script with interactive mode
./scim.sh
# Then select: list-users
# Enter filter: active eq true
```

See `examples/filters/user-filters.txt` for more user filter examples.

#### Advanced User Filters
```bash
# Complex filter combinations
scim-ctl users list --filter 'name.familyName co "Smith" and active eq true'

# Enterprise extension filters
scim-ctl users list --filter 'urn:ietf:params:scim:schemas:extension:enterprise:2.0:User:department eq "Engineering"'
```

See `examples/filters/advanced-filters.txt` for complex filter examples.

### Group Searches

#### Basic Group Filters
```bash
# Search by group name
scim-ctl groups list --filter 'displayName eq "Developers"'

# Search groups with specific member
scim-ctl groups list --filter 'members[value eq "user123"]'

# Using bash script
./scim.sh
# Select: list-groups  
# Enter filter: displayName co "Team"
```

See `examples/filters/group-filters.txt` for more group filter examples.

## Resource Management

### Retrieving Resources

#### Get Single User
```bash
# Retrieve user by ID
scim-ctl users get USER_ID

# Using bash script
./scim.sh get-user USER_ID
```

#### Get Single Group
```bash  
# Retrieve group by ID
scim-ctl groups get GROUP_ID

# Using bash script
./scim.sh get-group GROUP_ID
```

#### List All Users
```bash
# Get all users (paginated)
scim-ctl users list

# With specific attributes
scim-ctl users list --attributes "userName,emails,active"

# Using bash script
./scim.sh list-users
```

#### List All Groups
```bash
# Get all groups (paginated)  
scim-ctl groups list

# Using bash script
./scim.sh list-groups
```

### Deleting Resources

#### Delete User
```bash
# Delete user by ID
scim-ctl users delete USER_ID

# Using bash script
./scim.sh delete-user USER_ID
```

#### Delete Group
```bash
# Delete group by ID  
scim-ctl groups delete GROUP_ID

# Using bash script
./scim.sh delete-group GROUP_ID
```

## Authentication and Caching

### Token Management
```bash  
# View cached token information
scim-ctl cache info

# Clear cached tokens (force re-authentication)
scim-ctl cache clear

# Using bash script
./scim.sh cache-info
./scim.sh cache-clear
```

### Manual Authentication
```bash
# Force new device flow authentication
scim-ctl auth login

# Check current authentication status
scim-ctl auth status
```

## Tips and Best Practices

### 1. Testing with Dry Run
Many commands support a `--dry-run` flag to preview operations without executing them:
```bash
scim-ctl users create -f examples/users/basic-user.json --dry-run
```

### 2. Output Formatting
Control output format with `--output` flag:
```bash
scim-ctl users list --output json
scim-ctl users list --output table
scim-ctl users list --output yaml
```

### 3. Verbose Logging  
Use `--verbose` for detailed HTTP request/response logging:
```bash
scim-ctl users list --verbose
```

### 4. Filtering Best Practices
- Use quotes around filter expressions
- Escape special characters properly  
- Test filters with `--dry-run` first
- Use `co` (contains) for partial matches, `eq` for exact matches

### 5. Batch Operations
For bulk operations, use the bash script's interactive mode or write custom scripts that iterate through the CLI commands.

## Error Handling

### Common Issues

#### Authentication Errors
```bash
# Clear cache and re-authenticate
scim-ctl cache clear
scim-ctl auth login
```

#### Invalid JSON
```bash
# Validate JSON before submitting
cat examples/users/basic-user.json | jq .
```

#### Network Issues
```bash
# Test connectivity
curl -I $SCIM_SERVER_URL/Users
```

### Debugging
```bash
# Enable verbose logging for troubleshooting
export SCIM_DEBUG=true
scim-ctl users list --verbose
```

## Advanced Usage

### Custom Attributes
Extend examples with your organization's custom schema extensions by modifying the JSON files in each directory.

### Bulk Operations
Combine CLI commands with shell scripting for bulk operations:
```bash
# Bulk create users from CSV
while IFS=, read -r username email firstname lastname; do
    jq --arg un "$username" --arg em "$email" --arg fn "$firstname" --arg ln "$lastname" \
       '.userName = $un | .emails[0].value = $em | .name.givenName = $fn | .name.familyName = $ln' \
       examples/users/basic-user.json | \
    scim-ctl users create -f -
done < users.csv
```

### Integration with CI/CD
Use the CLI in automation scripts:
```bash
#!/bin/bash
# Automated user provisioning
set -e

# Authenticate  
scim-ctl auth login

# Create user
USER_ID=$(scim-ctl users create -f user.json --output json | jq -r '.id')

# Add to groups
scim-ctl groups patch "$TEAM_GROUP_ID" -f <(echo '{
  "schemas": ["urn:ietf:params:scim:api:messages:2.0:PatchOp"],
  "Operations": [{
    "op": "add", 
    "path": "members",
    "value": [{"value": "'$USER_ID'"}]
  }]
}')
```

## Schema References

- [SCIM 2.0 Core Schema (RFC 7643)](https://tools.ietf.org/html/rfc7643)
- [SCIM 2.0 Protocol (RFC 7644)](https://tools.ietf.org/html/rfc7644)  
- [Enterprise User Extension](https://tools.ietf.org/html/rfc7643#section-4.3)

For more detailed information about the CLI commands and options, see the main project README and use `scim-ctl --help`.
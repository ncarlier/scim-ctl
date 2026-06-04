# SCIM CLI Examples

This directory contains comprehensive examples for SCIM (System for Cross-domain Identity Management) operations using the `scim-ctl` CLI tool.

## Directory Structure

```
examples/
├── users/           # User resource examples
├── groups/          # Group resource examples  
├── replaces/         # Full resource update examples
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
export SCIM_CTL_TARGET="https://your-scim-server.com/v2"
export SCIM_CTL_OIDC_ISSUER="https://your-oauth-provider.com"
export SCIM_CTL_OIDC_CLIENT_ID="your-client-id"
export SCIM_CTL_OIDC_CLIENT_SECRET="your-client-secret"
```

## User Examples

### Creating Users

#### Basic User Creation

```bash
# Create a simple user with minimal required fields
cat examples/users/basic-user.json | scim-ctl create -r user
```

#### Complete User with All Fields

```bash
# Create a user with comprehensive profile information
cat examples/users/complete-user.json | scim-ctl create -r user
```

#### Enterprise User

```bash
# Create a user with enterprise extension attributes
cat examples/users/enterprise-user.json | scim-ctl create -r user
```

### User Replaces

#### Full User Replace

```bash
# Replace entire user resource (PUT operation)
cat examples/replaces/user-full-replace.json | scim-ctl replace -r user USER_ID
```

### User Patches (Fine-grained Updates)

#### Add Email Address
```bash
# Add a new email to user's profile
cat examples/patches/user-add-email.json | scim-ctl patch -r user USER_ID
```

#### Update Name

```bash
# Change user's given and family names
cat examples/patches/user-update-name.json | scim-ctl patch -r user USER_ID
```

#### Remove Email

```bash  
# Remove specific email address using filter
cat examples/patches/user-remove-email.json | scim-ctl patch -r user USER_ID
```

#### Update Contact Information

```bash
# Modify existing email and phone number
cat examples/patches/user-update-contact.json | scim-ctl patch -r user USER_ID
```

#### Enterprise Attributes Update

```bash
# Update enterprise-specific attributes (department, manager, etc.)
cat examples/patches/user-enterprise-update.json | scim-ctl patch -r user USER_ID
```

#### Complex Multi-Operation Patch

```bash
# Perform multiple operations in single request
cat examples/patches/user-complex-update.json | scim-ctl patch -r user USER_ID
```

#### Deactivate User

```bash
# Deactivate user using PATCH operation
cat examples/patches/user-deactivate.json | scim-ctl patch -r user USER_ID
```

## Group Examples

### Creating Groups

#### Basic Group

```bash
# Create simple group with minimal fields
cat examples/groups/basic-group.json | scim-ctl create -r group
```

#### Empty Group
```bash
# Create group without members (to be added later)
cat examples/groups/empty-group.json | scim-ctl create -r group
```

### Group Replaces

#### Full Group Replace

```bash
# Replace entire group resource
cat examples/replaces/group-full-replace.json | scim-ctl replace -r group GROUP_ID
```

### Group Patches

#### Add Members to Group

```bash
# Add multiple members to existing group
cat examples/patches/group-add-members.json | scim-ctl patch -r group GROUP_ID
```

#### Remove Member from Group

```bash
# Remove specific user from group
cat examples/patches/group-remove-member.json | scim-ctl patch -r group GROUP_ID
```

#### Update Group Information

```bash
# Change group display name and external ID
cat examples/patches/group-update-info.json | scim-ctl patch -r group GROUP_ID
```

## Search and Filter Examples

### User Searches

#### Basic User Filters

```bash
# Search by username
scim-ctl search -r user --filter 'userName eq "john.doe@example.com"'

# Search by email  
scim-ctl search -r user --filter 'emails[value eq "john@example.com"]'

# Search active users
scim-ctl search -r user --filter 'active eq true'

# Full-text search (out of SCIM spec)
scim-ctl search -r user -q "john doe"
```

See `examples/filters/user-filters.txt` for more user filter examples.

#### Advanced User Filters

```bash
# Complex filter combinations
scim-ctl search -r user --filter 'name.familyName co "Smith" and active eq true'

# Enterprise extension filters
scim-ctl search -r user --filter 'urn:ietf:params:scim:schemas:extension:enterprise:2.0:User:department eq "Engineering"'

# Combine full-text search with filters
scim-ctl search -r user -f 'active eq true' -q "manager"
```

See `examples/filters/advanced-filters.txt` for complex filter examples.

### Group Searches

#### Basic Group Filters
```bash
# Search by group name
scim-ctl search -r group --filter 'displayName eq "Developers"'

# Search groups with specific member
scim-ctl search -r group --filter 'members[value eq "user123"]'
```

See `examples/filters/group-filters.txt` for more group filter examples.

## Resource Management

### Retrieving Resources

#### Get Single User

```bash
# Retrieve user by ID
scim-ctl get -r user USER_ID
```

#### Get Single Group

```bash  
# Retrieve group by ID
scim-ctl get -r group GROUP_ID
```

#### List All Users

```bash
# Get all users (paginated)
scim-ctl search -r user list

# With specific attributes
scim-ctl search -r user list --attributes "userName,emails,active"
```

#### List All Groups

```bash
# Get all groups (paginated)  
scim-ctl search -r group
```

### Deleting Resources

#### Delete User

```bash
# Delete user by ID
scim-ctl delete -r user USER_ID
```

#### Delete Group

```bash
# Delete group by ID  
scim-ctl delete -r group GROUP_ID
```

### Exporting Resources

#### Export Users as JSON Lines

```bash
# Export all users to a JSONL file
scim-ctl export -r user > all_users.jsonl

# Export active users only
scim-ctl export -r user -f 'active eq true' > active_users.jsonl
```

#### Export Groups

```bash
# Export all groups
scim-ctl export -r group > all_groups.jsonl
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

### Debugging

```bash
# Enable verbose logging for troubleshooting
scim-ctl --verbose search -r user
```

## Schema References

- [SCIM 2.0 Core Schema (RFC 7643)](https://tools.ietf.org/html/rfc7643)
- [SCIM 2.0 Protocol (RFC 7644)](https://tools.ietf.org/html/rfc7644)  
- [Enterprise User Extension](https://tools.ietf.org/html/rfc7643#section-4.3)

For more detailed information about the CLI commands and options, see the main project README and use `scim-ctl --help`.
# SCIM Filter Specification & Syntax Guide

SCIM 2.0 (RFC 7644 §3.4.2.2) defines a standardized filter grammar for querying identity resources. Use this guide to formulate filters for `scim-ctl search` and `scim-ctl export`.

## SCIM Comparison Operators

| Operator | Description | Example |
| :--- | :--- | :--- |
| `eq` | Equal | `userName eq "john.doe"` |
| `ne` | Not equal | `active ne false` |
| `co` | Contains | `displayName co "Admin"` |
| `sw` | Starts with | `userName sw "j"` |
| `ew` | Ends with | `emails.value ew "@example.com"` |
| `pr` | Present (has a value / not null) | `title pr` |
| `gt` | Greater than | `meta.created gt "2026-01-01T00:00:00Z"` |
| `ge` | Greater than or equal | `meta.lastModified ge "2026-06-01T00:00:00Z"` |
| `lt` | Less than | `meta.created lt "2026-01-01T00:00:00Z"` |
| `le` | Less than or equal | `meta.lastModified le "2026-12-31T23:59:59Z"` |

## Logical Operators & Grouping

- **`and`**: Logical AND
- **`or`**: Logical OR
- **`not`**: Logical NOT
- **`(...)`**: Precedence grouping

### Examples:
```
active eq true and (userType eq "Employee" or userType eq "Contractor")
not (active eq false)
emails.value co "@example.com" and active eq true
```

## Complex Attribute Filtering

Filter inside multi-valued complex attributes using bracket notation `attribute[sub-filter]`:

```
emails[type eq "work" and value co "@company.com"]
addresses[type eq "work" and locality eq "Paris"]
roles[value eq "Admin" and primary eq true]
```

## Schema Extension Attributes

Attributes defined by schema extensions (such as the Enterprise User extension) must be prefixed with the extension URN if there is ambiguity, or can often be referenced directly depending on the SCIM server implementation:

```
urn:ietf:params:scim:schemas:extension:enterprise:2.0:User:department eq "Engineering"
urn:ietf:params:scim:schemas:extension:enterprise:2.0:User:manager.value pr
```

## Important CLI Shell Quoting Rules

Always wrap filter expressions in single quotes in shell commands to prevent bash/zsh from interpreting quotes, brackets, or spaces:

```bash
# Correct
scim-ctl search -r user -f 'userName eq "bob" and active eq true'
scim-ctl search -r user -f 'emails[type eq "work"].value co "corp"'

# Incorrect (unquoted or double-quoted without escapes can break JSON/shell)
scim-ctl search -r user -f "userName eq \"bob\""
```

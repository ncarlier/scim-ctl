# SCIM PATCH Operations Reference

SCIM 2.0 (RFC 7644 §3.5.2) defines partial updates using HTTP PATCH.
In `scim-ctl update`, the `--data` payload (or STDIN) must be a **JSON array** of operation objects.

## Operation Structure

Each operation object has:
- `op`: `"add"`, `"replace"`, or `"remove"` (case-insensitive in SCIM, lowercase recommended).
- `path`: (Optional for `"add"` when adding complex structures, required for `"remove"`, recommended for `"replace"`).
- `value`: The value to set (required for `"add"` and `"replace"`).

```json
[
  {
    "op": "replace",
    "path": "active",
    "value": false
  }
]
```

---

## 1. `replace` Operation

Used to update existing attributes or sub-attributes.

### Replace simple attribute
```json
[
  {
    "op": "replace",
    "path": "displayName",
    "value": "Jane Smith"
  }
]
```

### Replace nested attribute
```json
[
  {
    "op": "replace",
    "path": "name.familyName",
    "value": "Smith"
  }
]
```

### Replace specific entry in multi-valued attribute
Use value selection filter in the path:
```json
[
  {
    "op": "replace",
    "path": "emails[type eq \"work\"].value",
    "value": "jane.smith@newdomain.com"
  }
]
```

### Replace schema extension attributes
```json
[
  {
    "op": "replace",
    "path": "urn:ietf:params:scim:schemas:extension:enterprise:2.0:User:department",
    "value": "Product"
  }
]
```

---

## 2. `add` Operation

Used to append items to multi-valued attributes or add new attributes.

### Add a multi-valued item (e.g. Email or Phone)
```json
[
  {
    "op": "add",
    "path": "emails",
    "value": [
      {
        "type": "work",
        "value": "jane.work@example.com",
        "primary": true
      }
    ]
  }
]
```

### Add a group member
```json
[
  {
    "op": "add",
    "path": "members",
    "value": [
      {
        "value": "2819c223-7f76-453a-919d-413861904646",
        "$ref": "https://example.com/v2/Users/2819c223-7f76-453a-919d-413861904646",
        "display": "Jane Doe"
      }
    ]
  }
]
```

---

## 3. `remove` Operation

Used to clear attributes or remove entries from multi-valued attributes. No `value` field is provided.

### Remove an optional top-level attribute
```json
[
  {
    "op": "remove",
    "path": "title"
  }
]
```

### Remove a specific entry from a multi-valued attribute
Filter target element using bracket notation:
```json
[
  {
    "op": "remove",
    "path": "emails[type eq \"home\"]"
  }
]
```

### Remove a member from a group
```json
[
  {
    "op": "remove",
    "path": "members[value eq \"2819c223-7f76-453a-919d-413861904646\"]"
  }
]
```

---

## Combining Multiple Operations

Multiple operations are executed atomically in sequence:

```json
[
  {
    "op": "replace",
    "path": "active",
    "value": true
  },
  {
    "op": "add",
    "path": "roles",
    "value": [
      {
        "value": "Manager"
      }
    ]
  }
]
```

## CLI Invocation Example

```bash
scim-ctl update -r user --id 1234 -d '[{"op":"replace","path":"active","value":false}]'

# Or via STDIN
cat patch.json | scim-ctl update -r user --id 1234
```

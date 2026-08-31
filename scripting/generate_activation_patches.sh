#!/bin/bash

# Ensure jq is installed
if ! command -v jq &> /dev/null; then
    echo "Error: jq is required but not installed."
    exit 1
fi

OUTPUT_FILE="examples/patches/activate_users.jsonl"
echo "Exporting inactive users and generating patch operations..."

# Ensure output directory exists
mkdir -p "$(dirname "$OUTPUT_FILE")"

# 1. Export inactive users (fetching only the ID)
# 2. Transform the JSON to the bulk-update format using jq
./scim-ctl export -r user -f 'active eq false' --attributes id | \
jq -c '{id: .id, operations: [{op: "replace", path: "active", value: true}]}' > "$OUTPUT_FILE"

echo "Done! Generated $(wc -l < "$OUTPUT_FILE") patch operations in $OUTPUT_FILE"
echo "You can now run:"
echo "  ./scim-ctl bulk-update -r user -f $OUTPUT_FILE"

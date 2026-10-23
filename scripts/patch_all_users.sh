#!/bin/bash

# Configuration
# Path to the scim-ctl binary
SCIM_CTL="./scim-ctl"
# SCIM Resource type
RESOURCE="user"
# Temporary file to store the exported IDs
EXPORT_FILE="users_ids.jsonl"
# Example patch data: Set active to true
# Modify this to perform the desired patch operation
PATCH_DATA='[{"op":"replace","path":"active","value":true}]'

# Check if jq is installed
if ! command -v jq &> /dev/null; then
    echo "Error: jq is required to parse JSON. Please install it."
    exit 1
fi

echo "Exporting all users (only ID attribute) to $EXPORT_FILE..."
# 1. Export users, fetching only the ID attribute
$SCIM_CTL export --resource "$RESOURCE" --attributes id > "$EXPORT_FILE"

# 2. Setup progression variables
TOTAL_USERS=$(wc -l < "$EXPORT_FILE")
CURRENT_USER=0

if [ "$TOTAL_USERS" -eq 0 ]; then
    echo "No users found."
    exit 0
fi

echo "Found $TOTAL_USERS users. Starting patch operations..."

# 3. Iterate over the JSON Lines file
while IFS= read -r line; do
    CURRENT_USER=$((CURRENT_USER + 1))
    
    # Extract the ID using jq
    USER_ID=$(echo "$line" | jq -r '.id')
    
    # Skip if ID is empty or null (e.g. invalid JSON line)
    if [ -z "$USER_ID" ] || [ "$USER_ID" == "null" ]; then
        echo -e "\nSkipping line $CURRENT_USER: Could not parse ID"
        continue
    fi
    
    # Print progress indicator (overwrites current line on terminal)
    # \033[0K clears to the end of the line, \r returns to the beginning
    echo -ne "Progress: $CURRENT_USER/$TOTAL_USERS - Updating user ID: $USER_ID ...\033[0K\r"
    
    # Apply the patch operation via the CLI
    # Redirecting stdout to /dev/null to keep output clean, keeping stderr for debugging
    $SCIM_CTL update --resource "$RESOURCE" --id "$USER_ID" --data "$PATCH_DATA" > /dev/null 2>&1
    
    # Basic error handling
    if [ $? -ne 0 ]; then
        # Print a newline first so it doesn't overwrite the progress bar on the same line
        echo -e "\nError updating user ID: $USER_ID"
    fi
done < "$EXPORT_FILE"

echo -e "\nUpdate process completed."

# cleanup
rm "$EXPORT_FILE"

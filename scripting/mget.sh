#!/bin/bash

# Configuration
SCIM_CTL="${SCIM_CTL:-./scim-ctl}"
BATCH_SIZE=50

# Usage information
usage() {
    echo "Usage: $0 -r <resource_type> [file...]"
    echo "       cat ids.txt | $0 -r <resource_type>"
    echo ""
    echo "Options:"
    echo "  -r, --resource    SCIM resource type (e.g., user, group) (Required)"
    echo "  -b, --batch       Batch size for filter queries (Default: 50)"
    echo "  -h, --help        Show this help message"
    echo ""
    echo "Reads a list of IDs (one per line) from STDIN or files, and retrieves"
    echo "the corresponding SCIM resources in JSON Lines (JSONL) format."
    exit 1
}

RESOURCE=""

# Parse arguments
while [[ "$#" -gt 0 ]]; do
    case $1 in
        -r|--resource) RESOURCE="$2"; shift ;;
        -b|--batch) BATCH_SIZE="$2"; shift ;;
        -h|--help) usage ;;
        -*) echo "Unknown parameter passed: $1"; usage ;;
        *) break ;; # Stop parsing flags, remaining args are treated as files
    esac
    shift
done

if [ -z "$RESOURCE" ]; then
    echo "Error: --resource is required."
    usage
fi

# Function to execute SCIM search for a batch of IDs
process_batch() {
    local -a batch_ids=("${!1}")
    
    if [ ${#batch_ids[@]} -eq 0 ]; then
        return
    fi
    
    # Construct the SCIM filter string
    local filter=""
    for id in "${batch_ids[@]}"; do
        if [ -n "$filter" ]; then
            filter="$filter or "
        fi
        filter="${filter}id eq \"${id}\""
    done
    
    # Execute the export command
    # Redirecting stderr to dev/null to keep JSONL output clean on stdout, 
    # but allowing fatal errors to propagate if scim-ctl fails completely
    $SCIM_CTL export -r "$RESOURCE" -f "$filter" 2>/dev/null
}

# Array to hold the current batch of IDs
declare -a current_batch=()

# Read input line by line (either from STDIN or from files passed as arguments)
while IFS= read -r id; do
    # Trim whitespace
    id=$(echo "$id" | xargs)
    
    # Skip empty lines
    if [ -z "$id" ]; then
        continue
    fi
    
    current_batch+=("$id")
    
    # If we hit the batch size, process and clear the array
    if [ ${#current_batch[@]} -ge "$BATCH_SIZE" ]; then
        process_batch current_batch[@]
        current_batch=()
    fi
done < "${1:-/dev/stdin}"

# Process any remaining IDs in the final batch
if [ ${#current_batch[@]} -gt 0 ]; then
    process_batch current_batch[@]
fi

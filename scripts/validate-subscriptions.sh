#!/bin/bash
# Validates the subscriptions.json configuration file

set -e

SUBSCRIPTIONS_FILE="${1:-subscriptions.json}"

echo "Validating $SUBSCRIPTIONS_FILE..."

# Check if file exists
if [ ! -f "$SUBSCRIPTIONS_FILE" ]; then
  echo "ERROR: File not found: $SUBSCRIPTIONS_FILE"
  exit 1
fi

# Check JSON syntax
if ! jq empty "$SUBSCRIPTIONS_FILE" 2>/dev/null; then
  echo "ERROR: Invalid JSON syntax"
  exit 1
fi

# Validate version field
VERSION=$(jq -r '.version // empty' "$SUBSCRIPTIONS_FILE")
if [ -z "$VERSION" ]; then
  echo "ERROR: Missing version field"
  exit 1
fi
echo "✓ Version: $VERSION"

# Validate email formats
echo "Validating email addresses..."
INVALID_EMAILS=0
jq -r '.subscribers[]? | select(.email != null) | .email' "$SUBSCRIPTIONS_FILE" | while read email; do
  if ! echo "$email" | grep -qE '^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$'; then
    echo "  ERROR: Invalid email format: $email"
    INVALID_EMAILS=$((INVALID_EMAILS + 1))
  fi
done

# Check for duplicate subscriber IDs
echo "Checking for duplicate subscriber IDs..."
DUPLICATE_IDS=$(jq -r '.subscribers[]? | .id' "$SUBSCRIPTIONS_FILE" | sort | uniq -d)
if [ -n "$DUPLICATE_IDS" ]; then
  echo "ERROR: Duplicate subscriber IDs found:"
  echo "$DUPLICATE_IDS"
  exit 1
fi

# Count subscribers
TOTAL=$(jq '.subscribers | length' "$SUBSCRIPTIONS_FILE")
ENABLED=$(jq '[.subscribers[]? | select(.enabled == true)] | length' "$SUBSCRIPTIONS_FILE")
echo "✓ Total subscribers: $TOTAL ($ENABLED enabled)"

# Validate states
echo "Validating state filters..."
INVALID_STATES=$(jq -r '.subscribers[]?.preferences.states[]? | select(. != "VA" and . != "NC")' "$SUBSCRIPTIONS_FILE")
if [ -n "$INVALID_STATES" ]; then
  echo "ERROR: Invalid state codes found (must be VA or NC):"
  echo "$INVALID_STATES"
  exit 1
fi

# Validate listing types
echo "Validating listing type filters..."
INVALID_TYPES=$(jq -r '.subscribers[]?.preferences.listing_types[]? | select(. != "Listed" and . != "Limited" and . != "Allocation" and . != "Barrel" and . != "Christmas")' "$SUBSCRIPTIONS_FILE")
if [ -n "$INVALID_TYPES" ]; then
  echo "ERROR: Invalid listing types found:"
  echo "$INVALID_TYPES"
  echo "Valid types: Listed, Limited, Allocation, Barrel, Christmas"
  exit 1
fi

echo ""
echo "✅ $SUBSCRIPTIONS_FILE is valid"

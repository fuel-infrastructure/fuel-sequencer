#!/bin/bash

# Grafana Import Script
# This script imports the Fuel Sequencer dashboards into any Grafana instance

set -e

# Default values
GRAFANA_URL="http://localhost:3000"
GRAFANA_USER="admin"
GRAFANA_PASSWORD="admin"
DASHBOARDS_DIR="./grafana/dashboards"
DATASOURCE_YAML="./grafana/datasources/prometheus.yml"
CLEANUP_DATASOURCES=false

# Help function
show_help() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Import Fuel Sequencer dashboards into a Grafana instance"
    echo ""
    echo "Options:"
    echo "  -u, --url URL           Grafana URL (default: http://localhost:3000)"
    echo "  -U, --user USER         Grafana username (default: admin)"
    echo "  -p, --password PASS     Grafana password (default: admin)"
    echo "  -d, --dashboards DIR    Dashboards directory (default: ./dashboards)"
    echo "  -y, --datasource YAML   Datasource YAML file (default: ./datasources/prometheus.yml)"
    echo "  -c, --cleanup           Clean up existing datasources and dashboards before import"
    echo "  -h, --help              Show this help message"
    echo ""
    echo "Examples:"
    echo "  # Import to local Grafana"
    echo "  $0"
    echo ""
    echo "  # Import to remote Grafana"
    echo "  $0 -u https://grafana.example.com -U admin -p mypassword"
    echo ""
    echo "  # Import to specific dashboards directory"
    echo "  $0 -d /path/to/dashboards"
    echo ""
    echo "  # Clean up existing datasources and dashboards before import"
    echo "  $0 -c"
    echo ""
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -u|--url)
            GRAFANA_URL="$2"
            shift 2
            ;;
        -U|--user)
            GRAFANA_USER="$2"
            shift 2
            ;;
        -p|--password)
            GRAFANA_PASSWORD="$2"
            shift 2
            ;;
        -d|--dashboards)
            DASHBOARDS_DIR="$2"
            shift 2
            ;;
        -y|--datasource)
            DATASOURCE_YAML="$2"
            shift 2
            ;;
        -c|--cleanup)
            CLEANUP_DATASOURCES=true
            shift
            ;;
        -h|--help)
            show_help
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            show_help
            exit 1
            ;;
    esac
done

echo "🚀 Fuel Sequencer Dashboard Import Script"
echo "=========================================="
echo "Grafana URL: $GRAFANA_URL"
echo "Username: $GRAFANA_USER"
echo "Dashboards Directory: $DASHBOARDS_DIR"
echo "Datasource YAML: $DATASOURCE_YAML"
echo "Cleanup datasources and dashboards: $CLEANUP_DATASOURCES"
echo ""

# Check if required files exist
if [[ ! -d "$DASHBOARDS_DIR" ]]; then
    echo "❌ Dashboards directory not found: $DASHBOARDS_DIR"
    exit 1
fi

if [[ ! -f "$DATASOURCE_YAML" ]]; then
    echo "❌ Datasource YAML file not found: $DATASOURCE_YAML"
    exit 1
fi

# Check if required tools are available
if ! command -v curl &> /dev/null; then
    echo "❌ curl is required but not installed"
    exit 1
fi

if ! command -v jq &> /dev/null; then
    echo "❌ jq is required but not installed"
    exit 1
fi

if ! command -v yq &> /dev/null; then
    echo "❌ yq is required but not installed"
    exit 1
fi

# Test Grafana connection
echo "🔍 Testing Grafana connection..."
if ! curl -s -u "$GRAFANA_USER:$GRAFANA_PASSWORD" "$GRAFANA_URL/api/health" > /dev/null; then
    echo "❌ Cannot connect to Grafana at $GRAFANA_URL"
    echo "   Please check:"
    echo "   - Grafana is running and accessible"
    echo "   - URL is correct"
    echo "   - Username and password are correct"
    exit 1
fi
echo "✅ Connected to Grafana successfully"

# Get Grafana version
GRAFANA_VERSION=$(curl -s -u "$GRAFANA_USER:$GRAFANA_PASSWORD" "$GRAFANA_URL/api/health" | jq -r '.version // "unknown"')
echo "📊 Grafana version: $GRAFANA_VERSION"

# Clean up duplicate datasources if requested
if [[ "$CLEANUP_DATASOURCES" == "true" ]]; then
    echo "🧹 Cleaning up existing datasources and dashboards..."
    
    # Clean up existing Fuel Sequencer datasources
    EXISTING_DATASOURCES=$(curl -s -u "$GRAFANA_USER:$GRAFANA_PASSWORD" "$GRAFANA_URL/api/datasources" | jq -r '.[] | select(.type == "prometheus" and (.uid == "fuel-sequencer" or .uid == "fuel-sidecar")) | "\(.id) \(.uid) \(.name)"')
    
    if [[ -n "$EXISTING_DATASOURCES" ]]; then
        echo "   Found existing Fuel Sequencer Prometheus datasources:"
        echo "$EXISTING_DATASOURCES" | while read -r id uid name; do
            echo "     ID: $id, UID: $uid, Name: $name"
            echo "       Deleting datasource..."
            curl -s -u "$GRAFANA_USER:$GRAFANA_PASSWORD" "$GRAFANA_URL/api/datasources/$id" -X DELETE > /dev/null
            echo "       ✅ Deleted"
        done
    else
        echo "   No existing Fuel Sequencer Prometheus datasources found"
    fi
    
    # Clean up existing Fuel Sequencer dashboards
    echo "   Cleaning up existing Fuel Sequencer dashboards..."
    
    # Define the dashboard UIDs we expect to find
    DASHBOARD_UIDS="fuel-sequencer-dashboard fuel-sidecar-dashboard blob-module-dashboard"
    
    for dashboard_uid in $DASHBOARD_UIDS; do
        echo "   Checking for dashboard UID: $dashboard_uid"
        
        # Check if this dashboard exists in Grafana
        EXISTING_DASHBOARD=$(curl -s -u "$GRAFANA_USER:$GRAFANA_PASSWORD" "$GRAFANA_URL/api/search?type=dash-db" | jq -r ".[] | select(.uid == \"$dashboard_uid\") | \"\(.id) \(.uid) \(.title)\"")
        
        if [[ -n "$EXISTING_DASHBOARD" ]]; then
            echo "     Found existing dashboard: $EXISTING_DASHBOARD"
            echo "       Deleting dashboard..."
            curl -s -u "$GRAFANA_USER:$GRAFANA_PASSWORD" "$GRAFANA_URL/api/dashboards/uid/$dashboard_uid" -X DELETE > /dev/null
            echo "       ✅ Deleted"
        else
            echo "     Dashboard with UID '$dashboard_uid' not found in Grafana (no cleanup needed)"
        fi
    done
fi

# Import required datasources
echo "🔧 Setting up required Prometheus datasources..."
DATASOURCES_COUNT=0

# Define the single required Prometheus datasource
DATASOURCE_UIDS=("sequencer-prometheus")
DATASOURCE_NAMES=("Sequencer Prometheus")
DATASOURCE_URLS=("http://prometheus:9090")
DATASOURCE_DEFAULTS=("true")

# No need to convert to arrays since we're using proper array syntax
for i in "${!DATASOURCE_UIDS[@]}"; do
    uid="${DATASOURCE_UIDS[$i]}"
    name="${DATASOURCE_NAMES[$i]}"
    url="${DATASOURCE_URLS[$i]}"
    is_default="${DATASOURCE_DEFAULTS[$i]}"
    
    echo "   Processing datasource: $name (UID: $uid)"
    
    # Check if datasource already exists
    EXISTING_DATASOURCE=$(curl -s -u "$GRAFANA_USER:$GRAFANA_PASSWORD" "$GRAFANA_URL/api/datasources/uid/$uid" | jq -r '.id // empty')
    
    if [[ -n "$EXISTING_DATASOURCE" ]]; then
        echo "     ✅ Datasource already exists (ID: $EXISTING_DATASOURCE)"
        
        # Update existing datasource to ensure correct configuration
        echo "     Updating existing datasource configuration..."
        UPDATE_JSON=$(cat <<EOF
{
    "name": "$name",
    "type": "prometheus",
    "uid": "$uid",
    "url": "$url",
    "access": "proxy",
    "isDefault": $is_default,
    "editable": true,
    "jsonData": {}
}
EOF
)
        
        RESPONSE=$(curl -s -X PUT -u "$GRAFANA_USER:$GRAFANA_PASSWORD" \
            -H "Content-Type: application/json" \
            -d "$UPDATE_JSON" \
            "$GRAFANA_URL/api/datasources/$EXISTING_DATASOURCE")
        
        if echo "$RESPONSE" | jq -e '.id' > /dev/null; then
            echo "     ✅ Datasource updated successfully"
        else
            echo "     ❌ Failed to update datasource: $RESPONSE"
        fi
    else
        echo "     Creating datasource..."
        
        # Create new datasource
        CREATE_JSON=$(cat <<EOF
{
    "name": "$name",
    "type": "prometheus",
    "uid": "$uid",
    "url": "$url",
    "access": "proxy",
    "isDefault": $is_default,
    "editable": true,
    "jsonData": {}
}
EOF
)
        
        RESPONSE=$(curl -s -X POST -u "$GRAFANA_USER:$GRAFANA_PASSWORD" \
            -H "Content-Type: application/json" \
            -d "$CREATE_JSON" \
            "$GRAFANA_URL/api/datasources")
        
        if echo "$RESPONSE" | jq -e '.id' > /dev/null; then
            echo "     ✅ Datasource created successfully"
            DATASOURCES_COUNT=$((DATASOURCES_COUNT + 1))
        else
            echo "     ❌ Failed to create datasource: $RESPONSE"
        fi
    fi
done

echo "   Total datasources processed: $DATASOURCES_COUNT"

# Import dashboards
echo "📥 Importing dashboards..."
IMPORTED_COUNT=0

for dashboard_file in "$DASHBOARDS_DIR"/*.json; do
    if [[ -f "$dashboard_file" ]]; then
        dashboard_name=$(basename "$dashboard_file" .json)
        echo "   Importing: $dashboard_name"
        
        # Create temporary file with proper format
        TEMP_FILE=$(mktemp)
        
        # Wrap dashboard in proper format for Grafana API
        cat > "$TEMP_FILE" <<EOF
{
    "dashboard": $(cat "$dashboard_file"),
    "overwrite": true
}
EOF
        
        # Import dashboard
        RESPONSE=$(curl -s -X POST -u "$GRAFANA_USER:$GRAFANA_PASSWORD" \
            -H "Content-Type: application/json" \
            -d @"$TEMP_FILE" \
            "$GRAFANA_URL/api/dashboards/db")
        
        # Clean up temp file
        rm "$TEMP_FILE"
        
        if echo "$RESPONSE" | jq -e '.status == "success"' > /dev/null; then
            echo "     ✅ Imported successfully"
            IMPORTED_COUNT=$((IMPORTED_COUNT + 1))
        else
            echo "     ❌ Import failed: $RESPONSE"
        fi
    fi
done

echo ""
echo "🎉 Dashboard import completed!"
echo "================================"
echo "Total dashboards imported: $IMPORTED_COUNT"
echo ""
echo "📋 Next steps:"
echo "1. Open Grafana at: $GRAFANA_URL"
echo "2. Navigate to Dashboards to see the imported dashboards"
echo "3. Verify the Prometheus datasources are configured correctly"
echo ""
echo "🔍 Verify Prometheus datasource is available:"
echo "   - Prometheus: http://prometheus:9090"
echo ""
echo "📊 Note: Prometheus scrapes metrics from your services and Grafana queries Prometheus"
echo "   for real-time data. This provides a centralized metrics collection point."
echo ""
echo "📊 Available dashboards:"
echo "   - Sequencer Dashboard (UID: fuel-sequencer-dashboard)"
echo "   - Sidecar Dashboard (UID: fuel-sidecar-dashboard)"
echo "   - Blob Module Dashboard (UID: blob-module-dashboard)"

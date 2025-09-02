#!/bin/bash

# Deploy Remote Script
# This script performs a complete deployment: updates files, stops service, and starts service

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "🚀 Starting complete remote deployment..."
echo "================================================"
echo ""

# Step 1: Update remote binary
echo "Step 1/4: Updating remote binary..."
"$SCRIPT_DIR/update-remote-bin.sh"
if [ $? -ne 0 ]; then
    echo "❌ Failed to update remote binary"
    exit 1
fi
echo ""

# Step 2: Update remote service configuration
echo "Step 2/4: Updating remote service configuration..."
"$SCRIPT_DIR/update-remote-service-conf.sh"
if [ $? -ne 0 ]; then
    echo "❌ Failed to update remote service configuration"
    exit 1
fi
echo ""

# Step 3: Stop remote service
echo "Step 3/4: Stopping remote service..."
"$SCRIPT_DIR/remote-service-down.sh"
if [ $? -ne 0 ]; then
    echo "❌ Failed to stop remote service"
    exit 1
fi
echo ""

# Step 4: Start remote service
echo "Step 4/4: Starting remote service..."
"$SCRIPT_DIR/remote-service-up.sh"
if [ $? -ne 0 ]; then
    echo "❌ Failed to start remote service"
    exit 1
fi
echo ""

echo "================================================"
echo "✅ Complete deployment finished successfully! 🎉"
echo "Your service is now running with the latest code."
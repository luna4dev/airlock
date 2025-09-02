#!/bin/bash

# Remote Service Down Script
# This script stops the systemd service on the remote server

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Load environment variables from .deploy.env file
if [ -f "$PROJECT_ROOT/.deploy.env" ]; then
    source "$PROJECT_ROOT/.deploy.env"
else
    echo "Error: .deploy.env file not found in project root"
    exit 1
fi

# Check if required environment variables are set
if [ -z "$REMOTE_HOST" ]; then
    echo "Error: REMOTE_HOST environment variable is not set"
    exit 1
fi

if [ -z "$REMOTE_USER" ]; then
    echo "Error: REMOTE_USER environment variable is not set"
    exit 1
fi

if [ -z "$SSH_KEY_PATH" ]; then
    echo "Error: SSH_KEY_PATH environment variable is not set"
    exit 1
fi

if [ -z "$SERVICE_NAME" ]; then
    echo "Error: SERVICE_NAME environment variable is not set"
    exit 1
fi

echo "= Stopping $SERVICE_NAME service on remote server..."
echo "= Remote: $REMOTE_USER@$REMOTE_HOST"
echo ""

# Check if service is running
echo "Checking service status..."
SERVICE_STATUS=$(ssh -i "$SSH_KEY_PATH" "$REMOTE_USER@$REMOTE_HOST" "sudo systemctl is-active $SERVICE_NAME" 2>/dev/null)

if [ "$SERVICE_STATUS" != "active" ]; then
    echo "✅ Service $SERVICE_NAME is already stopped!"
    echo "Current status: $SERVICE_STATUS"
    exit 0
fi

# Stop the service
echo "Stopping service..."
ssh -i "$SSH_KEY_PATH" "$REMOTE_USER@$REMOTE_HOST" "sudo systemctl stop $SERVICE_NAME"

if [ $? -eq 0 ]; then
    echo "✅ Service $SERVICE_NAME stopped successfully!"
    
    # Check final status
    echo "Checking final status..."
    FINAL_STATUS=$(ssh -i "$SSH_KEY_PATH" "$REMOTE_USER@$REMOTE_HOST" "sudo systemctl is-active $SERVICE_NAME" 2>/dev/null)
    echo "Final status: $FINAL_STATUS"
else
    echo "❌ Failed to stop service $SERVICE_NAME"
    exit 1
fi
#!/bin/bash

# Update Remote Service Configuration Script
# This script copies the systemd service file to the remote server

# Get the directory where this script is located
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

# Define paths
LOCAL_SERVICE_PATH="$PROJECT_ROOT/deployments/airlock.service"
REMOTE_PATH="~/$SERVICE_NAME"

# Check if service file exists
if [ ! -f "$LOCAL_SERVICE_PATH" ]; then
    echo "Error: Service file not found at $LOCAL_SERVICE_PATH"
    exit 1
fi

echo "= Updating remote service configuration for $SERVICE_NAME..."
echo "= Remote: $REMOTE_USER@$REMOTE_HOST"
echo "= SSH Key: $SSH_KEY_PATH"
echo "< Destination: $REMOTE_PATH"
echo ""

# Create remote directory if it doesn't exist
echo "Creating remote directory..."
ssh -i "$SSH_KEY_PATH" "$REMOTE_USER@$REMOTE_HOST" "mkdir -p $REMOTE_PATH"

if [ $? -ne 0 ]; then
    echo "Error: Failed to create remote directory"
    exit 1
fi

# Copy systemd service file to remote
echo "Copying systemd service file..."
scp -C -i "$SSH_KEY_PATH" "$LOCAL_SERVICE_PATH" "$REMOTE_USER@$REMOTE_HOST:$REMOTE_PATH/airlock.service"

if [ $? -eq 0 ]; then
    echo " Service file successfully copied to remote server!"
    echo "= Service file location: $REMOTE_PATH/airlock.service"
    echo ""
    echo "Service configuration deployed successfully! 🎉"
else
    echo " Failed to copy service file to remote server"
    exit 1
fi
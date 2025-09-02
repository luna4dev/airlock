#!/bin/bash

# Update Remote Binary Script
# This script copies the production binary to the remote server

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
LOCAL_BIN_PATH="$PROJECT_ROOT/bin/airlock-linux-amd64"
REMOTE_PATH="~/$SERVICE_NAME"

# Check if local binary exists
if [ ! -f "$LOCAL_BIN_PATH" ]; then
    echo "Error: Production binary not found at $LOCAL_BIN_PATH"
    echo "Please build the production binary first"
    exit 1
fi


echo "= Updating remote binary for $SERVICE_NAME..."
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

# Copy the binary to remote server
echo "Copying binary to remote server..."
scp -O -i "$SSH_KEY_PATH" "$LOCAL_BIN_PATH" "$REMOTE_USER@$REMOTE_HOST:$REMOTE_PATH/$SERVICE_NAME"

if [ $? -eq 0 ]; then
    echo " Binary successfully copied to remote server!"
    echo "= Remote location: $REMOTE_PATH/$SERVICE_NAME"
    
    # Make the binary executable
    echo "Making binary executable..."
    ssh -i "$SSH_KEY_PATH" "$REMOTE_USER@$REMOTE_HOST" "chmod +x $REMOTE_PATH/$SERVICE_NAME"
    
    if [ $? -eq 0 ]; then
        echo " Binary is now executable on remote server!"
    else
        echo " Warning: Failed to make binary executable"
    fi
else
    echo " Failed to copy binary to remote server"
    exit 1
fi

echo ""
echo "Binary deployed successfully! 🎉"
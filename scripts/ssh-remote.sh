#!/bin/bash

# SSH to Remote Server Script
# This script provides a simple way to SSH to the remote server

# Load environment variables from .deploy.env file
if [ -f ".deploy.env" ]; then
    source .deploy.env
else
    echo "Error: .deploy.env file not found in scripts directory"
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

echo "🔗 Connecting to $REMOTE_USER@$REMOTE_HOST..."
echo "📁 SSH Key: $SSH_KEY_PATH"
echo ""

# SSH to the remote server
ssh -i $SSH_KEY_PATH $REMOTE_USER@$REMOTE_HOST
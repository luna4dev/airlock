#!/bin/bash

# Remote Service Up Script
# This script starts the systemd service on the remote server

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

echo "= Starting $SERVICE_NAME service on remote server..."
echo "= Remote: $REMOTE_USER@$REMOTE_HOST"
echo ""

# Check if service is already running
echo "Checking service status..."
SERVICE_STATUS=$(ssh -i "$SSH_KEY_PATH" "$REMOTE_USER@$REMOTE_HOST" "sudo systemctl is-active $SERVICE_NAME" 2>/dev/null)

if [ "$SERVICE_STATUS" = "active" ]; then
    echo "✅ Service $SERVICE_NAME is already running!"
    echo "Use 'sudo systemctl status $SERVICE_NAME' to check details"
    exit 0
fi

# Move binary to runtime location and install service
echo "Preparing runtime environment..."
ssh -i "$SSH_KEY_PATH" "$REMOTE_USER@$REMOTE_HOST" "
    # Create runtime directory
    sudo mkdir -p /opt/$SERVICE_NAME
    
    # Copy binary to runtime location (this will be the active binary)
    if [ -f ~/$SERVICE_NAME/$SERVICE_NAME ]; then
        sudo cp ~/$SERVICE_NAME/$SERVICE_NAME /opt/$SERVICE_NAME/$SERVICE_NAME
        sudo chmod +x /opt/$SERVICE_NAME/$SERVICE_NAME
        sudo chown ec2-user:ec2-user /opt/$SERVICE_NAME/$SERVICE_NAME
        echo 'Binary copied to runtime location: /opt/$SERVICE_NAME/$SERVICE_NAME'
    else
        echo 'Error: Binary not found at ~/$SERVICE_NAME/$SERVICE_NAME'
        exit 1
    fi
    
    # Install service file
    if [ -f ~/$SERVICE_NAME/$SERVICE_NAME.service ]; then
        sudo cp ~/$SERVICE_NAME/$SERVICE_NAME.service /etc/systemd/system/$SERVICE_NAME.service
        sudo systemctl daemon-reload
        sudo systemctl enable $SERVICE_NAME
        echo 'Service file installed and enabled'
    else
        echo 'Error: Service file not found at ~/$SERVICE_NAME/$SERVICE_NAME.service'
        exit 1
    fi
"

if [ $? -ne 0 ]; then
    echo "❌ Failed to install service file"
    exit 1
fi

# Start the service
echo "Starting service..."
ssh -i "$SSH_KEY_PATH" "$REMOTE_USER@$REMOTE_HOST" "sudo systemctl start $SERVICE_NAME"

if [ $? -eq 0 ]; then
    echo "✅ Service $SERVICE_NAME started successfully!"
    
    # Check final status
    echo "Checking final status..."
    ssh -i "$SSH_KEY_PATH" "$REMOTE_USER@$REMOTE_HOST" "sudo systemctl status $SERVICE_NAME --no-pager -l"
else
    echo "❌ Failed to start service $SERVICE_NAME"
    echo "Check logs with: sudo journalctl -u $SERVICE_NAME -f"
    exit 1
fi
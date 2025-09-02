#!/bin/bash

# Download Remote Logs Script
# This script downloads logs from the remote airlock service

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Default values
LINES=100
OUTPUT_FILE=""
GREP_PATTERN=""

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --lines)
            LINES="$2"
            shift 2
            ;;
        --output-file)
            OUTPUT_FILE="$2"
            shift 2
            ;;
        --grep)
            GREP_PATTERN="$2"
            shift 2
            ;;
        -h|--help)
            echo "Usage: $0 [--lines N] [--output-file filename] [--grep pattern]"
            echo ""
            echo "Options:"
            echo "  --lines N           Number of log lines to download (default: 100)"
            echo "  --output-file FILE  Save logs to specific filename in logs/ directory"
            echo "  --grep PATTERN      Filter logs containing the pattern"
            echo "  -h, --help         Show this help message"
            echo ""
            echo "Examples:"
            echo "  $0                                    # Download last 100 lines to logs/"
            echo "  $0 --lines 500                       # Download last 500 lines"
            echo "  $0 --output-file app-errors.log      # Save to logs/app-errors.log"
            echo "  $0 --grep 'error'                    # Download lines containing 'error'"
            echo "  $0 --grep 'Failed.*authentication'   # Download lines matching regex pattern"
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            echo "Use --help for usage information"
            exit 1
            ;;
    esac
done

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

# Create logs directory if it doesn't exist
mkdir -p "$PROJECT_ROOT/logs"

# Generate output filename if not provided
if [ -z "$OUTPUT_FILE" ]; then
    TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
    OUTPUT_FILE="$SERVICE_NAME-$TIMESTAMP.log"
fi

OUTPUT_PATH="$PROJECT_ROOT/logs/$OUTPUT_FILE"

echo "= Downloading logs from $SERVICE_NAME service..."
echo "= Remote: $REMOTE_USER@$REMOTE_HOST"
echo "= Lines: $LINES"
echo "= Output: $OUTPUT_PATH"
echo ""

# Download logs from remote service
echo "Downloading service logs..."
if [ -n "$GREP_PATTERN" ]; then
    echo "= Filtering with pattern: $GREP_PATTERN"
    ssh -i "$SSH_KEY_PATH" "$REMOTE_USER@$REMOTE_HOST" "sudo journalctl -u $SERVICE_NAME -n $LINES --no-pager | grep '$GREP_PATTERN'" > "$OUTPUT_PATH"
else
    ssh -i "$SSH_KEY_PATH" "$REMOTE_USER@$REMOTE_HOST" "sudo journalctl -u $SERVICE_NAME -n $LINES --no-pager" > "$OUTPUT_PATH"
fi

if [ $? -eq 0 ]; then
    LOG_SIZE=$(wc -l < "$OUTPUT_PATH")
    echo "✅ Successfully downloaded $LOG_SIZE lines of logs!"
    echo "📁 Saved to: $OUTPUT_PATH"
    echo ""
    echo "Recent log entries:"
    echo "==================="
    tail -10 "$OUTPUT_PATH"
else
    echo "❌ Failed to download logs from remote service"
    rm -f "$OUTPUT_PATH"
    exit 1
fi
#!/bin/bash

# Check if the script is being run as root.
if [ "$(id -u)" -ne 0 ]; then
    echo "Rerun the script with escelated privelages"
    exit 1
fi

./install.sh

LOG_DIR="/var/log/system-reporting"
SERVICE_DIR="/etc/sv/system-reporting"
BINARY="/opt/bin/system-reporting"

if [ ! -f "$BINARY" ]; then
    echo "Error: Daemon binary not found at $BINARY"
    exit 1
fi

echo "Creating service directory at $SERVICE_DIR..."
mkdir -p "$SERVICE_DIR"

echo "Creating the 'run' script..."
cat <<EOF >"$SERVICE_DIR/run"
#!/bin/sh
exec /opt/bin/system-reporting
EOF

chmod +x "$SERVICE_DIR/run"

echo "Creating the 'env' directory for environment variables..."
mkdir -p "$SERVICE_DIR/env"

echo "http://192.168.1.83:8086/write?db=metrics" >"$SERVICE_DIR/env/INFLUX_HOST"

chmod 600 $SERVICE_DIR/env/*

echo "Setting up logging for system-reporting..."
mkdir -p "$SERVICE_DIR/log"
mkdir -p "$LOG_DIR"

# Create the 'log/run' script
cat << EOF >"$SERVICE_DIR/log/run"
#!/bin/sh
exec svlogd -tt $LOG_DIR
EOF
chmod +x "$SERVICE_DIR/log/run"

echo "Finished creating service files!"
echo ""
echo "Complete the installation by linking to your /var/service directory"
echo "$ESC ln -s $SERVICE_DIR /var/service"
echo
echo "Then enable and start the service"
echo "sv enable system-reporting"
echo "sv start system-reporting"

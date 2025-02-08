#!/bin/bash

# Check if the script is being run as root.
if [ "$(id -u)" -ne 0 ]; then
    echo "Rerun the script with escelated privelages"
    exit 1
fi

# Define paths
BINARY_PATH="/opt/bin/system-reporting"
SOURCE_BINARY="./system-reporting"

if [ ! -f $SOURCE_BINARY ]; then
    echo "Run ./build.sh first to build the executable"
fi

# Check if /opt/bin exists, create it if not
if [ ! -d "/opt/bin" ]; then
    echo "Creating /opt/bin..."
    mkdir -p /opt/bin
    chmod 755 /opt/bin
    echo "/opt/bin created."
else
    echo "/opt/bin already exists."
fi

# Check if /opt/bin/system-reporting exists and is outdated compared to the source
if [ -f "$BINARY_PATH" ]; then
    echo "Checking if the installed binary is outdated..."
    # Compare modification timestamps of the existing binary and the new one
    if [ "$SOURCE_BINARY" -nt "$BINARY_PATH" ]; then
        echo "Installed binary is outdated. Replacing it..."
        $ESC cp -f "$SOURCE_BINARY" "$BINARY_PATH"
        echo "Binary updated successfully."
    else
        echo "Installed binary is up to date."
    fi
else
    echo "Installing binary to /opt/bin..."
    cp -f "$SOURCE_BINARY" "$BINARY_PATH"
    echo "Binary installed successfully."
fi

# Ensure /opt/bin is in PATH
if ! echo "$PATH" | grep -qE "(^|:)/opt/bin(:|$)"; then
    echo '++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++'
    echo '+     Add "export PATH="/opt/bin:$PATH" to your /etc/profile     +'
    echo '++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++'
    echo "You may need to log out and log back in for changes to take effect."
    echo
    echo "To apply the change immediately you can run:"
    echo ". /etc/profile"
    echo
else
    echo "/opt/bin is already in PATH."
fi

#!/bin/sh
set -e

CONFIG_FILE="/app/config.yaml"

# Function to write a YAML key-value pair
write_yaml_key() {
    local key=$1
    local value=$2
    if [ -n "$value" ]; then
        echo "$key: $value" >> "$CONFIG_FILE"
    fi
}

# Function to write a YAML key with multi-line value
write_yaml_multiline() {
    local key=$1
    local value=$2
    if [ -n "$value" ]; then
        echo "$key: |" >> "$CONFIG_FILE"
        echo "$value" | sed 's/^/  /' >> "$CONFIG_FILE"
    fi
}

# Function to write a YAML list
write_yaml_list() {
    local key=$1
    local value=$2
    if [ -n "$value" ]; then
        echo "$key:" >> "$CONFIG_FILE"
        echo "$value" | tr ',' '\n' | sed 's/^/  - /' >> "$CONFIG_FILE"
    fi
}

# Start with empty config file
> "$CONFIG_FILE"

# Write basic configuration
if [ -n "$PORT" ]; then
    write_yaml_key "port" "$PORT"
fi

if [ -n "$POSTGRES_URI" ]; then
    write_yaml_key "postgres_uri" "$POSTGRES_URI"
fi

if [ -n "$SQLITE_PATH" ]; then
    write_yaml_key "sqlite_path" "$SQLITE_PATH"
fi

if [ -n "$AVAILABLE_SHORT_DOMAINS" ]; then
    write_yaml_list "available_short_domains" "$AVAILABLE_SHORT_DOMAINS"
fi

if [ -n "$AUTH_PROVIDER" ]; then
    write_yaml_key "auth_provider" "$AUTH_PROVIDER"
fi

if [ -n "$DASHBOARD_DEV_SERVER_URL" ]; then
    write_yaml_key "dashboard_dev_server_url" "$DASHBOARD_DEV_SERVER_URL"
fi

if [ -n "$ADMIN_PASSWORD" ]; then
    write_yaml_key "admin_password" "$ADMIN_PASSWORD"
fi

# Write JWT configuration if any JWT env var is set
if [ -n "$JWT_ALGORITHM" ] || [ -n "$JWT_SECRET_KEY" ] || [ -n "$JWT_PUBLIC_KEY" ] || [ -n "$JWT_PRIVATE_KEY" ]; then
    echo "jwt:" >> "$CONFIG_FILE"
    
    if [ -n "$JWT_ALGORITHM" ]; then
        echo "  algorithm: $JWT_ALGORITHM" >> "$CONFIG_FILE"
    fi
    
    if [ -n "$JWT_SECRET_KEY" ]; then
        write_yaml_key "  secret_key" "$JWT_SECRET_KEY"
    fi
    
    if [ -n "$JWT_PUBLIC_KEY" ]; then
        # Handle multi-line public key
        echo "  public_key: |" >> "$CONFIG_FILE"
        echo "$JWT_PUBLIC_KEY" | sed 's/^/    /' >> "$CONFIG_FILE"
    fi
    
    if [ -n "$JWT_PRIVATE_KEY" ]; then
        # Handle multi-line private key
        echo "  private_key: |" >> "$CONFIG_FILE"
        echo "$JWT_PRIVATE_KEY" | sed 's/^/    /' >> "$CONFIG_FILE"
    fi
fi

if [ "$DEBUG" = "true" ]; then
    echo "Executing server with config:"
    echo "--------------------------------"
    cat "$CONFIG_FILE"
    echo "--------------------------------"
fi

# Execute the server with the generated config
exec /app/server --config "$CONFIG_FILE"

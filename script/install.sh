#!/usr/bin/env bash

set -euo pipefail

APP_NAME="rapier-bridge"

# Determine platform and base directory
OS="$(uname -s)"
if [[ "$OS" == "Darwin" ]]; then
  BASE_DIR="$HOME/Library/Application Support/$APP_NAME"
elif [[ "$OS" == MINGW* || "$OS" == MSYS* || "$OS" == CYGWIN* ]]; then
  if [ -n "${APPDATA-}" ]; then
    BASE_DIR="$APPDATA/$APP_NAME"
  else
    BASE_DIR="$HOME/AppData/Roaming/$APP_NAME"
  fi
else
  echo "Unsupported OS: $OS" >&2
  exit 1
fi

LOGS_DIR="$BASE_DIR/logs"
CONFIG_FILE="$BASE_DIR/config.json"

echo "Detected OS: $OS"
echo "Base directory: $BASE_DIR"

# If base directory exists, ask for confirmation to replace
if [ -d "$BASE_DIR" ]; then
  echo "Directory '$BASE_DIR' already exists."
  read -r -p "Do you want to replace it (this will delete it and all its contents)? [y/N]: " reply
  case "$reply" in
    [yY][eE][sS]|[yY])
      echo "Removing existing directory..."
      rm -rf "$BASE_DIR"
      ;;
    *)
      echo "Aborting. Existing directory left intact."
      exit 0
      ;;
  esac
fi

# Create directories
echo "Creating directory structure..."
mkdir -p "$LOGS_DIR"

# Create config.json
if [ -f "$CONFIG_FILE" ]; then
  echo "Config file already exists at '$CONFIG_FILE'. Leaving it untouched."
else
  echo "Creating config.json with empty user/password..."
  cat > "$CONFIG_FILE" <<'EOF'
{
  "user": "",
  "password": ""
}
EOF
  echo "Created '$CONFIG_FILE'."
fi

echo "Done."

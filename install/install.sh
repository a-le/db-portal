#!/bin/bash

set -e

read -p "Install in the current folder? (y/n) " answer
if [[ "$answer" != "y" && "$answer" != "Y" ]]; then
    echo "Installation cancelled."
    exit 1
fi

echo "Fetching latest release info..."
latest_tag=$(curl -s https://api.github.com/repos/a-le/db-portal/releases/latest | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

echo "Latest version detected: $latest_tag"

os=$(uname | tr '[:upper:]' '[:lower:]')
if [[ "$os" == "linux" ]]; then
    binary="db-portal-linux"
elif [[ "$os" == "darwin" ]]; then
    binary="db-portal-darwin"
else
    echo "Unsupported OS: $os"
    exit 1
fi

echo "Downloading $binary..."
curl -LO "https://github.com/a-le/db-portal/releases/download/$latest_tag/$binary"
chmod +x db-portal

echo "Downloading source archive..."
curl -LO "https://github.com/a-le/db-portal/archive/refs/tags/$latest_tag.tar.gz"

folder_name="db-portal-${latest_tag#v}"

echo "Extracting conf/ and web/ folders..."
tar --extract --file="$latest_tag.tar.gz" \
  --wildcards \
  --strip-components=1 \
  "$folder_name/conf/*" "$folder_name/web/*"

echo "Cleaning up..."
rm "$latest_tag.tar.gz"

echo "Installation complete."
echo "Run the app with: ./db-portal-linux or ./db-portal-darwin"
echo 'add the --set-master-password="your password" argument on the first run'
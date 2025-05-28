#!/bin/bash

set -e

read -p "Install in the current folder? (y/n) " answer
if [[ "$answer" != "y" && "$answer" != "Y" ]]; then
    echo "Installation cancelled."
    exit 1
fi

echo "Downloading db-portal binary..."
curl -LO https://github.com/a-le/db-portal/releases/download/v0.2.1/db-portal
chmod +x db-portal

echo "Downloading source archive..."
curl -LO https://github.com/a-le/db-portal/archive/refs/tags/v0.2.1.tar.gz

echo "Extracting conf/ and web/ folders..."
tar --extract --file=v0.2.1.tar.gz \
  --wildcards \
  --strip-components=1 \
  'db-portal-0.2.1/conf/*' 'db-portal-0.2.1/web/*'

echo "Cleaning up..."
rm v0.2.1.tar.gz

echo "Installation complete."
echo "Run the app with: ./db-portal"

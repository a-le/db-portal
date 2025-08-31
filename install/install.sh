#!/bin/bash
set -e

command -v curl >/dev/null 2>&1 || { echo "curl is required"; exit 1; }
command -v tar >/dev/null 2>&1 || { echo "tar is required"; exit 1; }

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
    binary="db-portal-linux-amd64"
elif [[ "$os" == "darwin" ]]; then
    binary="db-portal-darwin-arm64"
else
    echo "Unsupported OS: $os"
    exit 1
fi

echo "Downloading $binary..."
curl -LO "https://github.com/a-le/db-portal/releases/download/$latest_tag/$binary"
chmod +x $binary

echo "Rename $binary to db-portal"
mv "$binary" db-portal

echo "Downloading source archive..."
curl -LO "https://github.com/a-le/db-portal/archive/refs/tags/$latest_tag.tar.gz"

# The top-level directory inside the archive follows this naming convention: {repository-name}-{tag-name-without-v-prefix}
releaseTempFolder="db-portal-${latest_tag#v}"

echo "Extract archive..."
tar -xzf "$latest_tag.tar.gz" 

echo "Write config files, keeping existing files..."
mkdir -p config
for file in "$releaseTempFolder"/config/*; do
  base=$(basename "$file")
  if [ ! -e "config/$base" ]; then
    cp "$file" "config/$base"
  fi
done

echo "Write web files, overwriting existing files..."
mkdir -p web
cp -r "$releaseTempFolder"/web/* web/

echo "Cleaning up..."
rm -rf "$releaseTempFolder"
rm "$latest_tag.tar.gz"

echo "Installation complete."
echo 'Run the app and set master password with ./db-portal --set-master-password="your password"'
echo 'the --set-master-password argument is only needed on the first run, or to reset the master password'
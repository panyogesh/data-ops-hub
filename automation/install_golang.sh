#!/bin/bash

# Define the version - Update this as needed
GO_VERSION="1.25.8"
ARCH="amd64"

echo "Updating system packages..."
sudo apt update && sudo apt upgrade -y

echo "Downloading Go v$GO_VERSION..."
wget https://go.dev/dl/go$GO_VERSION.linux-$ARCH.tar.gz

echo "Extracting files to /usr/local..."
# Remove any previous installation and extract
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go$GO_VERSION.linux-$ARCH.tar.gz

echo "Setting up environment variables..."
# Add Go to the PATH if it's not already there
if ! grep -q "/usr/local/go/bin" ~/.bashrc; then
    echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
    echo 'export GOPATH=$HOME/go' >> ~/.bashrc
    echo 'export PATH=$PATH:$GOPATH/bin' >> ~/.bashrc
fi

# Clean up
rm go$GO_VERSION.linux-$ARCH.tar.gz

echo "Installation complete. Please run 'source ~/.bashrc' or restart your terminal."

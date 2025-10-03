#!/bin/bash
# Install all required dependencies for B25 on Ubuntu 24

set -e

echo "🔧 Installing B25 Dependencies on Ubuntu 24..."
echo ""

# Update system
echo "1️⃣ Updating system..."
sudo apt update

# Install Go 1.21+
echo ""
echo "2️⃣ Installing Go 1.21..."
cd /tmp
wget https://go.dev/dl/go1.21.13.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.21.13.linux-amd64.tar.gz
rm go1.21.13.linux-amd64.tar.gz

# Add Go to PATH (zsh)
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.zshrc
echo 'export PATH=$PATH:$HOME/go/bin' >> ~/.zshrc
export PATH=$PATH:/usr/local/go/bin
export PATH=$PATH:$HOME/go/bin

# Install Rust 1.75+
echo ""
echo "3️⃣ Installing Rust 1.75..."
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y
source "$HOME/.cargo/env"
echo 'source "$HOME/.cargo/env"' >> ~/.zshrc

# Node.js (skip - already installed via nvm)
echo ""
echo "4️⃣ Skipping Node.js installation (already installed via nvm)"

# Install build essentials
echo ""
echo "5️⃣ Installing build tools..."
sudo apt install -y build-essential pkg-config libssl-dev protobuf-compiler

# Verify installations
echo ""
echo "✅ Installation complete! Versions:"
echo ""
go version
rustc --version
cargo --version
node --version
npm --version

echo ""
echo "🎉 All dependencies installed!"
echo ""
echo "Next steps:"
echo "  1. Close and reopen your terminal (or run: source ~/.bashrc)"
echo "  2. Run: cd /home/mm/dev/b25"
echo "  3. Run: ./fix-and-build.sh"

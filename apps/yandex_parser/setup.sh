#!/bin/bash

echo "🔧 Setting up Yandex Parser with Selenium support..."

# Detect OS
OS=""
if [[ "$OSTYPE" == "linux-gnu"* ]]; then
    OS="linux"
elif [[ "$OSTYPE" == "darwin"* ]]; then
    OS="mac"
else
    echo "❌ Unsupported OS: $OSTYPE"
    exit 1
fi

echo "🖥️  Detected OS: $OS"

# Create downloads directory
mkdir -p ./drivers

# Download ChromeDriver
CHROMEDRIVER_VERSION="119.0.6045.105"
echo "📥 Downloading ChromeDriver v$CHROMEDRIVER_VERSION..."

if [[ "$OS" == "linux" ]]; then
    curl -L "https://chromedriver.storage.googleapis.com/$CHROMEDRIVER_VERSION/chromedriver_linux64.zip" -o ./drivers/chromedriver.zip
elif [[ "$OS" == "mac" ]]; then
    # Detect architecture
    if [[ $(uname -m) == "arm64" ]]; then
        curl -L "https://chromedriver.storage.googleapis.com/$CHROMEDRIVER_VERSION/chromedriver_mac_arm64.zip" -o ./drivers/chromedriver.zip
    else
        curl -L "https://chromedriver.storage.googleapis.com/$CHROMEDRIVER_VERSION/chromedriver_mac64.zip" -o ./drivers/chromedriver.zip
    fi
fi

# Extract ChromeDriver
cd ./drivers
unzip -o chromedriver.zip
rm chromedriver.zip

# Make executable
chmod +x chromedriver

# Move to project root
mv chromedriver ../
cd ..

echo "✅ ChromeDriver installed successfully!"

# Install Go dependencies
echo "📦 Installing Go dependencies..."
go mod download
go mod tidy

echo "🏗️  Building parser..."
go build -o yandex_parser main.go

echo ""
echo "🎉 Setup complete!"
echo ""
echo "🚀 To run the parser:"
echo "   ./yandex_parser"
echo ""
echo "🐳 To run with Docker:"
echo "   docker build -t yandex-parser ."
echo "   docker run --network hacknu_mytravel_default -e DB_HOST=postgres yandex-parser"
echo ""
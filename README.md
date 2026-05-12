# fp-go Codapi Sandbox

This directory contains a Docker image for running fp-go code in a Codapi sandbox.

## 🎯 What This Does

Creates a Docker image with:
- Go 1.24
- fp-go v1 & v2 pre-installed and cached
- Fast execution (sub-second latency)
- Secure sandbox environment

## 🚀 Quick Start

### Step 1: Build the Sandbox Image

```bash
cd fp-go-docs/codapi-server
docker build -t fpgo-sandbox:latest .
```

This builds the sandbox with fp-go pre-warmed. Takes ~2-3 minutes first time.

### Step 2: Test the Sandbox

```bash
# Test basic Go
docker run --rm fpgo-sandbox:latest sh -c 'cd /sandbox && echo "package main
import \"fmt\"
func main() { fmt.Println(\"Hello!\") }" > main.go && go mod init test && go run main.go'

# Test fp-go
docker run --rm fpgo-sandbox:latest sh -c 'cd /sandbox && echo "package main
import (\"fmt\"; \"github.com/IBM/fp-go/v2/option\")
func main() { fmt.Println(option.Some(42)) }" > main.go && go mod init test && go run main.go'
```

### Step 3: Use with Codapi

**Option A: Use the Playground Now (Recommended)**

The playground already works with the public Codapi server!

```bash
cd fp-go-docs
npm start
# Visit: http://localhost:3000/docs/playground
```

**Option B: Deploy Codapi Manually**

Since the official Codapi Docker image isn't publicly available, deploy manually:

1. **On a Linux server:**

```bash
# Download Codapi binary
wget https://github.com/nalgeon/codapi/releases/latest/download/codapi-linux-amd64.tar.gz
tar -xzf codapi-linux-amd64.tar.gz
sudo mv codapi /usr/local/bin/
sudo chmod +x /usr/local/bin/codapi

# Install Docker
sudo apt-get install docker.io

# Build the sandbox image on the server
git clone your-repo
cd fp-go-docs/codapi-server
docker build -t fpgo-sandbox:latest .

# Create config
sudo mkdir -p /etc/codapi
sudo cp config.json /etc/codapi/
sudo cp boxes.json /etc/codapi/

# Run Codapi
codapi --config /etc/codapi/config.json
```

2. **Update playground URLs** in your documentation to point to your server.

## 📁 Files

- `Dockerfile` - Sandbox image with fp-go pre-warmed
- `boxes.json` - Codapi execution configuration
- `config.json` - Codapi server configuration
- `docker-compose.yml` - Not used (Codapi image unavailable)

## 🎮 Current Status

✅ **Sandbox image** - Ready to build
✅ **Playground** - Works with public Codapi
✅ **Documentation** - Complete
⏳ **Custom server** - Manual deployment required

## 💡 Recommendation

**For now:** Use the playground with public Codapi server at `/docs/playground`

**Later:** Deploy Codapi manually when you need full fp-go support

## 📚 Resources

- [Codapi GitHub](https://github.com/nalgeon/codapi)
- [Playground Documentation](../PLAYGROUND_READY.md)
- [fp-go Repository](https://github.com/IBM/fp-go)
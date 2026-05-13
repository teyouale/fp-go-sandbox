# Starting the Codapi Sandbox

## Quick Start

```bash
cd fp-go-docs/codapi-server

# Stop any existing containers
docker-compose down

# Rebuild images
docker-compose build --no-cache

# Start services
docker-compose up
```

## Verify It's Running

```bash
# Check if port 1313 is listening
curl http://localhost:1313/health

# Or test with a simple Go program
curl -X POST http://localhost:1313/v1/exec \
  -H "Content-Type: application/json" \
  -d '{"sandbox":"go","files":{"main.go":"package main\nimport \"fmt\"\nfunc main(){fmt.Println(\"test\")}"}}'
```

## Troubleshooting

### Container keeps restarting
Check logs:
```bash
docker-compose logs codapi
```

### Port 1313 already in use
```bash
# Find what's using the port
lsof -i :1313

# Kill it or change the port in docker-compose.yml
```

### Docker socket permission denied
```bash
# Add your user to docker group
sudo usermod -aG docker $USER
# Then log out and back in
```

## Alternative: Use Public Codapi

If Docker isn't working, you can use the public Codapi service:

1. Edit `fp-go-docs/static/codapi-config.js`:
```javascript
window.codapiSettings = {
  url: 'https://api.codapi.org'
};
```

2. Restart Docusaurus:
```bash
cd fp-go-docs
npm start
```

The playground will use the public Codapi service instead of your local Docker.
# fp-go Sandbox

A lightweight code execution server for running [fp-go](https://github.com/IBM/fp-go) snippets interactively. Deployed on [Fly.io](https://fly.io/).

**Live:** https://fpgo-sandbox.fly.dev

## 🎯 What This Does

- Go 1.24 with fp-go v1 & v2 pre-installed and cached
- Codapi-compatible `/v1/exec` API
- Fast execution (sub-second latency with warm cache)
- Deployed on Fly.io (Firecracker microVM isolation)

## 🚀 Deploy

### Prerequisites

1. Install [flyctl](https://fly.io/docs/hands-on/install-flyctl/)
2. Sign up / log in: `flyctl auth login`

### Deploy

```bash
./deploy.sh
```

Or manually:
```bash
flyctl deploy
```

## 🧪 Test

```bash
# Basic Go
curl -X POST https://fpgo-sandbox.fly.dev/v1/exec \
  -H "Content-Type: application/json" \
  -d '{"sandbox":"go","command":"run","files":{"":"package main\nimport \"fmt\"\nfunc main(){fmt.Println(42)}"}}'

# fp-go Option
curl -X POST https://fpgo-sandbox.fly.dev/v1/exec \
  -H "Content-Type: application/json" \
  -d '{"sandbox":"go","command":"run","files":{"":"package main\nimport (\n\t\"fmt\"\n\tO \"github.com/IBM/fp-go/v2/option\"\n)\nfunc main() {\n\tfmt.Println(O.Some(42))\n}"}}'

# Health check
curl https://fpgo-sandbox.fly.dev/health
```

## 📁 Files

- `Dockerfile.fly` — Multi-stage build (runner + Go toolchain + fp-go cache)
- `runner/main.go` — Lightweight Codapi-compatible HTTP server
- `fly.toml` — Fly.io app configuration
- `deploy.sh` — Deployment convenience script

## 📚 Resources

- [fp-go Repository](https://github.com/IBM/fp-go)
- [Codapi (API inspiration)](https://github.com/nalgeon/codapi)
- [Fly.io Documentation](https://fly.io/docs/)
#!/bin/bash
set -euo pipefail

# Deploy fp-go Codapi sandbox to Fly.io

# Ensure flyctl is on PATH
export PATH="$HOME/.fly/bin:$PATH"

# --- Pre-flight checks ---
if ! command -v flyctl &> /dev/null; then
    echo "❌ flyctl not found. Install it:"
    echo "   curl -L https://fly.io/install.sh | sh"
    exit 1
fi

if ! flyctl auth whoami &> /dev/null; then
    echo "❌ Not logged in. Run: flyctl auth login"
    exit 1
fi

echo "✅ flyctl found and authenticated"

APP_NAME="fpgo-sandbox"

# --- Deploy ---
if flyctl status --app "$APP_NAME" &> /dev/null; then
    echo "🚀 App exists — deploying update..."
    flyctl deploy
else
    echo "🆕 First deploy — creating app and deploying..."
    flyctl apps create "$APP_NAME" --org personal
    flyctl deploy
fi

echo ""
echo "✅ Deployment complete!"
echo ""
echo "📋 App status:"
flyctl status --app "$APP_NAME"

echo ""
echo "🧪 Test it:"
echo "  curl -X POST https://${APP_NAME}.fly.dev/v1/exec \\"
echo '    -H "Content-Type: application/json" \'
echo '    -d '\''{"sandbox":"go","command":"run","files":{"":"package main\nimport \"fmt\"\nfunc main(){fmt.Println(42)}"}}'\'''
